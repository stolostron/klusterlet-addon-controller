// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/bindata"
	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	appmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/workmgr/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/utils"
	"github.com/open-cluster-management/library-go/pkg/applier"
	"github.com/open-cluster-management/library-go/pkg/templateprocessor"
	ocinfrav1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	infrastructureConfigName = "cluster"
	apiserverConfigName      = "cluster"
	openshiftConfigNamespace = "openshift-config"
)

var addonsArray = []addons.KlusterletAddon{
	appmgr.AddonAppMgr{},
	certpolicyctrl.AddonCertPolicyCtrl{},
	iampolicyctrl.AddonIAMPolicyCtrl{},
	policyctrl.AddonPolicyCtrl{},
	search.AddonSearch{},
	workmgr.AddonWorkMgr{},
}
var merger applier.Merger = func(current,
	new *unstructured.Unstructured,
) (
	future *unstructured.Unstructured,
	update bool,
) {
	if spec, ok := new.Object["spec"]; ok &&
		!reflect.DeepEqual(spec, current.Object["spec"]) {
		update = true
		current.Object["spec"] = spec
	}
	if rules, ok := new.Object["rules"]; ok &&
		!reflect.DeepEqual(rules, current.Object["rules"]) {
		update = true
		current.Object["rules"] = rules
	}
	if roleRef, ok := new.Object["roleRef"]; ok &&
		!reflect.DeepEqual(roleRef, current.Object["roleRef"]) {
		update = true
		current.Object["roleRef"] = roleRef
	}
	if subjects, ok := new.Object["subjects"]; ok &&
		!reflect.DeepEqual(subjects, current.Object["subjects"]) {
		update = true
		current.Object["subjects"] = subjects
	}
	return current, update
}

//deleteOutDatedRoleRoleBindings deletes old role/rolebinding with klusterletaddonconfig ownerRef (controller).
//it returns nil if no role/rolebinding exist, and it returns error when failed to delete the role/rolebinding
func deleteOutDatedRoleRoleBinding(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
) error {
	if klusterletaddonconfig == nil {
		return nil
	}
	// name used in previous addon role & rolebinding
	name := klusterletaddonconfig.Name + "-" + addon.GetAddonName()
	// check if the role/rolebinding exist
	role := &rbacv1.Role{}
	rolebinding := &rbacv1.RoleBinding{}
	objs := make([]runtime.Object, 0)
	objs = append(objs, role)
	objs = append(objs, rolebinding)
	var retErr error
	retErr = nil
	for _, o := range objs {
		if err := client.Get(
			context.TODO(),
			types.NamespacedName{
				Name:      name,
				Namespace: klusterletaddonconfig.Namespace,
			}, o); err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			retErr = err
			continue
		}

		// verify ownerRef
		if objMetaAccessor, ok := o.(metav1.ObjectMetaAccessor); !ok {
			log.V(2).Info("Failed to get ObjectMeta")
			continue
		} else {
			ownerRef := metav1.GetControllerOf(objMetaAccessor.GetObjectMeta())
			if ownerRef == nil {
				log.V(2).Info("No controller reference of the role, skipping")
				continue
			}
			ownerGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
			if err != nil {
				log.V(2).Info("Failed to get group from object ownerRef")
				continue
			}
			if ownerRef.Kind != klusterletaddonconfig.Kind ||
				ownerRef.Name != klusterletaddonconfig.Name ||
				ownerGV.Group != klusterletaddonconfig.GroupVersionKind().Group {
				log.V(2).Info("Object is not owned by klusterletaddonconfig. Skipping")
				continue
			}
		}

		if err := client.Delete(context.TODO(), o); err != nil && !errors.IsNotFound(err) {
			retErr = err
			continue
		}
	}

	return retErr
}

func createOrUpdateHubKubeConfigResources(
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	r *ReconcileKlusterletAddon,
	addon addons.KlusterletAddon) error {
	componentName := addon.GetAddonName()

	//Create the values for the yamls
	config := struct {
		ManagedClusterName      string
		ManagedClusterNamespace string
		ServiceAccountName      string
		ManagedClusterAddonName string
		ClusterRoleName         string
	}{
		ManagedClusterName:      klusterletaddonconfig.Name + "-" + componentName,
		ManagedClusterNamespace: klusterletaddonconfig.Name,
		ServiceAccountName:      klusterletaddonconfig.Name + "-" + componentName,
		ManagedClusterAddonName: addon.GetManagedClusterAddOnName(),
		ClusterRoleName:         addons.GetAddonClusterRolePrefix() + addon.GetManagedClusterAddOnName(),
	}

	newApplier, err := applier.NewApplier(
		bindata.NewBindataReader(),
		&templateprocessor.Options{},
		r.client,
		klusterletaddonconfig,
		r.scheme,
		merger,
		&applier.Options{
			Backoff: &wait.Backoff{
				Steps:    1,
				Duration: 10 * time.Millisecond,
				Factor:   1.0,
			},
		},
	)
	if err != nil {
		return err
	}

	err = newApplier.CreateOrUpdateInPath(
		"resources/hub/common",
		nil,
		false,
		config,
	)
	if err != nil {
		return err
	}
	// delete old role & rolebindings created in previous releases
	if err := deleteOutDatedRoleRoleBinding(addon, klusterletaddonconfig, r.client); err != nil {
		log.Info("Failed to delete outdated role/rolebinding. Skipping.", "error message:", err)
	}

	return nil
}

// newCRManifestWork returns ManifestWork of a component CR
func newCRManifestWork(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client) (*manifestworkv1.ManifestWork, error) {
	var cr runtime.Object

	var err error
	cr, err = addon.NewAddonCR(klusterletaddonconfig, addonoperator.KlusterletAddonNamespace)

	if err != nil {
		return nil, err
	}

	// construct manifestwork
	var manifests []manifestworkv1.Manifest
	var manifest manifestworkv1.Manifest
	if addon.CheckHubKubeconfigRequired() {
		var secret runtime.Object
		secret, err = newHubKubeconfigSecret(
			klusterletaddonconfig,
			client,
			addon.GetAddonName(),
			addonoperator.KlusterletAddonNamespace,
		)
		if err != nil {
			return nil, err
		}
		manifest = manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret}}
		manifests = append(manifests, manifest)
	}

	manifest = manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: cr}}
	manifests = append(manifests, manifest)

	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
			Namespace: klusterletaddonconfig.Namespace,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}
	return manifestWork, nil
}

// syncManifestWorkCRs creates/updates/deletes all CR Manifestworks according to klusterletAddonConfig's configuration
// loops through all the components, and return the last error if there are errors, or return nil if succeeded
func syncManifestWorkCRs(klusterletaddonconfig *agentv1.KlusterletAddonConfig, r *ReconcileKlusterletAddon) error {
	var lastErr error
	lastErr = nil

	for _, addon := range addonsArray {
		addonName := addon.GetAddonName()
		// create sa/clusterrole/clusterrolebindig for each addon
		if addon.CheckHubKubeconfigRequired() {
			if err := createOrUpdateHubKubeConfigResources(klusterletaddonconfig, r, addon); err != nil {
				log.Error(err, fmt.Sprintf("Failed to create sa/clusterrole/clusterrolebindig for componnet %s", addonName))
				lastErr = err
				continue
			}
		}
		if addon.IsEnabled(klusterletaddonconfig) {
			// create Manifestwork if enabled
			if manifestWork, err := newCRManifestWork(addon, klusterletaddonconfig, r.client); err != nil {
				lastErr = err
			} else if err = utils.CreateOrUpdateManifestWork(
				manifestWork,
				r.client,
				klusterletaddonconfig,
				r.scheme,
			); err != nil {
				log.Error(err, "Failed to create manifest work for addon "+addonName)
				lastErr = err
			}
		} else {
			// delete Manifestwork if disabled
			if err := utils.DeleteManifestWork(
				addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
				klusterletaddonconfig.Namespace,
				r.client,
				false,
			); err != nil && !errors.IsNotFound(err) {
				log.Error(err, fmt.Sprintf("Failed to delete %s ManifestWork", addonName))
				lastErr = err
			}
		}
	}

	return lastErr
}

// syncManagedClusterAddonCRs creates/updates/deletes all CR ManagedClusterAddon according to klusterletAddonConfig's configuration
// loops through all the components, and return the last error if there are errors, or return nil if succeeded
func syncManagedClusterAddonCRs(klusterletaddonconfig *agentv1.KlusterletAddonConfig, r *ReconcileKlusterletAddon) error {
	var lastErr error
	lastErr = nil
	for _, addon := range addonsArray {
		if addon.IsEnabled(klusterletaddonconfig) {
			// create ManagedClusterAddon if enabled, and will not block if failed.
			// created ManagedClusterAddon should has controller reference points to the klusterletaddonconfig
			// and it should has the correct AddonRef in status
			if err := updateManagedClusterAddon(addon, klusterletaddonconfig, r.client, r.scheme); err != nil {
				log.Error(err, "Failed to create ManagedClusterAddon "+addon.GetAddonName())
				lastErr = err
			}
		}
	}
	return lastErr
}

// updateManagedClusterAddon updates ManagedClusterAddon to make sure it has correct reference in status
// if ManagedClusterAddon for an addon is not exist, will create the ManagedClusterAddon
// and will set controller reference to be the given klusterletaddonconfig
func updateManagedClusterAddon(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	scheme *runtime.Scheme,
) error {
	managedClusterAddon := &addonv1alpha1.ManagedClusterAddOn{}
	// check if it exists
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      addon.GetManagedClusterAddOnName(),
			Namespace: klusterletaddonconfig.Namespace,
		},
		managedClusterAddon,
	); err != nil && errors.IsNotFound(err) {
		// create new
		newManagedClusterAddon := &addonv1alpha1.ManagedClusterAddOn{
			TypeMeta: metav1.TypeMeta{
				APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
				Kind:       "ManagedClusterAddOn",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      addon.GetManagedClusterAddOnName(),
				Namespace: klusterletaddonconfig.Namespace,
			},
		}
		if err := controllerutil.SetControllerReference(klusterletaddonconfig, newManagedClusterAddon, scheme); err != nil {
			log.Error(err, "failed to set controller of ManagedClusterAddOn "+addon.GetManagedClusterAddOnName())
			return err
		}
		if err := client.Create(context.TODO(), newManagedClusterAddon); err != nil {
			log.Error(err, "")
			return err
		}
		managedClusterAddon = newManagedClusterAddon
	} else if err != nil {
		return err
	}
	ref := []addonv1alpha1.ObjectReference{
		addonv1alpha1.ObjectReference{
			Group:    agentv1.SchemeGroupVersion.Group,
			Resource: "klusterletaddonconfigs",
			Name:     klusterletaddonconfig.Name,
		},
	}
	addonMeta := addonv1alpha1.AddOnMeta{}
	addonConf := addonv1alpha1.ConfigCoordinates{}
	if addonMap, ok := clustermanagementaddon.ClusterManagementAddOnMap[addon.GetManagedClusterAddOnName()]; ok {
		addonMeta.Description = addonMap.Description
		addonMeta.DisplayName = addonMap.DisplayName
		addonConf.CRDName = addonMap.CRDName
		addonConf.CRName = klusterletaddonconfig.Name
	}

	if !reflect.DeepEqual(managedClusterAddon.Status.RelatedObjects, ref) ||
		!reflect.DeepEqual(managedClusterAddon.Status.AddOnMeta, addonMeta) ||
		!reflect.DeepEqual(managedClusterAddon.Status.AddOnConfiguration, addonConf) {
		managedClusterAddon.Status.RelatedObjects = ref
		managedClusterAddon.Status.AddOnMeta = addonMeta
		managedClusterAddon.Status.AddOnConfiguration = addonConf

		if err := client.Status().Update(context.TODO(), managedClusterAddon); err != nil {
			log.Error(err, fmt.Sprintf("Failed to update ManagedClusterAddon %s status", managedClusterAddon.Name))
			return err
		}
	}

	return nil
}

// deleteManifestWorkCRs deletes all CR Manifestworks
// returns true if deletion of all components is completed or component not found
func deleteManifestWorkCRs(
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	removeFinalizers bool) (bool, error) {
	allCompleted := true
	var lastErr error
	lastErr = nil
	for _, addon := range addonsArray {
		err := utils.DeleteManifestWork(
			addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
			klusterletaddonconfig.Namespace,
			client,
			removeFinalizers,
		)
		if err != nil && errors.IsNotFound(err) {
			continue
		}
		allCompleted = false
		if err != nil { // object still exist
			lastErr = err
		}
	}
	return allCompleted, lastErr
}

// getServiceAccountToken - retrieve service account token
func getServiceAccountToken(
	client client.Client,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	componentName string) (token []byte, cert []byte, retErr error) {
	// get service account created for component
	sa := &corev1.ServiceAccount{}

	if err := client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      klusterletaddonconfig.Name + "-" + componentName,
			Namespace: klusterletaddonconfig.Namespace,
		},
		sa,
	); err != nil {
		return nil, nil, err
	}

	saSecret := &corev1.Secret{}
	for _, secret := range sa.Secrets {
		secretNsN := types.NamespacedName{
			Name:      secret.Name,
			Namespace: sa.Namespace,
		}

		if err := client.Get(context.TODO(), secretNsN, saSecret); err != nil {
			continue
		}

		if saSecret.Type == corev1.SecretTypeServiceAccountToken {
			break
		}
	}

	token, ok := saSecret.Data["token"]
	if !ok {
		return nil, nil, fmt.Errorf("data of serviceaccount token secret does not contain token")
	}
	cert, ok = saSecret.Data["ca.crt"]
	if !ok {
		return token, nil, nil
	}

	return token, cert, nil
}

// getKubeAPIServerAddress - Get the API server address
func getKubeAPIServerAddress(client client.Client) (string, error) {
	infraConfig := &ocinfrav1.Infrastructure{}

	if err := client.Get(context.TODO(), types.NamespacedName{Name: infrastructureConfigName}, infraConfig); err != nil {
		return "", err
	}

	return infraConfig.Status.APIServerURL, nil
}

// getKubeAPIServerSecretName iterate through all namespacedCertificates
// returns the first one which has a name matches the given dnsName
func getKubeAPIServerSecretName(client client.Client, dnsName string) (string, error) {
	apiserver := &ocinfrav1.APIServer{}
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: apiserverConfigName},
		apiserver,
	); err != nil {
		if errors.IsNotFound(err) {
			log.Info("APIServer cluster not found")
			return "", nil
		}
		return "", err
	}
	// iterate through all namedcertificates
	for _, namedCert := range apiserver.Spec.ServingCerts.NamedCertificates {
		for _, name := range namedCert.Names {
			if strings.EqualFold(name, dnsName) {
				return namedCert.ServingCertificate.Name, nil
			}
		}
	}
	return "", nil
}

// getKubeAPIServerCertificate looks for secret in openshift-config namespace, and returns tls.crt
func getKubeAPIServerCertificate(client client.Client, secretName string) ([]byte, error) {
	secret := &corev1.Secret{}
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: secretName, Namespace: openshiftConfigNamespace},
		secret,
	); err != nil {
		log.Error(err, fmt.Sprintf("Failed to get secret %s/%s", openshiftConfigNamespace, secretName))
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if secret.Type != corev1.SecretTypeTLS {
		return nil, fmt.Errorf(
			"secret %s/%s should have type=kubernetes.io/tls",
			openshiftConfigNamespace,
			secretName,
		)
	}
	res, ok := secret.Data["tls.crt"]
	if !ok {
		return nil, fmt.Errorf(
			"failed to find data[tls.crt] in secret %s/%s",
			openshiftConfigNamespace,
			secretName,
		)
	}
	return res, nil
}

// checkIsIBMCloud detects if the current cloud vendor is ibm or not
// we know we are on OCP already, so if it's also ibm cloud, it's roks
func checkIsIBMCloud(client client.Client) (bool, error) {
	nodes := &corev1.NodeList{}
	err := client.List(context.TODO(), nodes)
	if err != nil {
		log.Error(err, "failed to get nodes list")
		return false, err
	}
	if len(nodes.Items) == 0 {
		log.Error(err, "failed to list any nodes")
		return false, nil
	}

	providerID := nodes.Items[0].Spec.ProviderID
	if strings.Contains(providerID, "ibm") {
		return true, nil
	}

	return false, nil
}

// getValidCertificatesFromURL dial to serverURL and get certificates
// only will return certificates signed by trusted ca and verified (with verifyOptions)
// if certificates are all signed by unauthorized party, will return nil
// rootCAs is for tls handshake verification
func getValidCertificatesFromURL(serverURL string, rootCAs *x509.CertPool) ([]*x509.Certificate, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		log.Error(err, "failed to parse url: "+serverURL)
		return nil, err
	}
	log.Info("getting certificate of " + u.Hostname() + ":" + u.Port())
	conf := &tls.Config{
		// server should support tls1.2
		MinVersion: tls.VersionTLS12,
		ServerName: u.Hostname(),
	}
	if rootCAs != nil {
		conf.RootCAs = rootCAs
	}

	conn, err := tls.Dial("tcp", u.Hostname()+":"+u.Port(), conf)

	if err != nil {
		log.Error(err, "failed to dial "+serverURL)
		// ignore certificate signed by unknown authority error
		if _, ok := err.(x509.UnknownAuthorityError); ok {
			return nil, nil
		}
		return nil, err
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	retCerts := []*x509.Certificate{}
	opt := x509.VerifyOptions{Roots: rootCAs}
	// check certificates
	for _, cert := range certs {
		if _, err := cert.Verify(opt); err == nil {
			log.V(2).Info("Adding a valid certificate")
			retCerts = append(retCerts, cert)
		} else {
			log.V(2).Info("Skipping an invalid certificate")
		}
	}
	return retCerts, nil
}

// newHubKubeconfigSecret -  creates a new hub-kubeconfig-secret
func newHubKubeconfigSecret(klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	componentName string,
	namespace string) (*corev1.Secret, error) {
	var certData []byte
	kubeAPIServer, err := getKubeAPIServerAddress(client)
	if err != nil {
		return nil, err
	}

	if u, err := url.Parse(kubeAPIServer); err == nil {
		apiServerCertSecretName, err := getKubeAPIServerSecretName(client, u.Hostname())
		if err != nil {
			return nil, err
		}
		if len(apiServerCertSecretName) > 0 {
			apiServerCert, err := getKubeAPIServerCertificate(client, apiServerCertSecretName)
			if err != nil {
				log.Error(err, "failed to get apiserver certificate, use default instead")
			} else if len(apiServerCert) > 0 {
				certData = apiServerCert
			}
		}
	}
	saToken, caCert, err := getServiceAccountToken(client, klusterletaddonconfig, componentName)
	if err != nil {
		return nil, err
	}
	if len(certData) == 0 {
		// fallback to service account token ca.crt
		if len(caCert) > 0 {
			certData = caCert
		}

		// check if it's roks
		// if it's ocp && it's on ibm cloud, we treat it as roks
		isROKS, err := checkIsIBMCloud(client)
		if err != nil {
			return nil, err
		}
		if isROKS {
			// ROKS should have a certificate that is signed by trusted CA
			if certs, err := getValidCertificatesFromURL(kubeAPIServer, nil); err != nil {
				// should retry if failed to connect to apiserver
				log.Error(err, fmt.Sprintf("failed to connect to %s", kubeAPIServer))
				return nil, err
			} else if len(certs) > 0 {
				// simply don't give any certs as the apiserver is using certs signed by known CAs
				certData = nil
			} else {
				log.Info("No additional valid certificate found for APIserver. Skipping.")
			}
		}
	}

	kubeConfig := clientcmdapi.Config{
		// Define a cluster stanza based on the bootstrap kubeconfig.
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   kubeAPIServer,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: certData,
		}},
		// Define auth based on the obtained client cert.
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			Token: string(saToken),
		}},
		// Define a context that connects the auth info and cluster, and set it as the default
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "default",
		}},
		CurrentContext: "default-context",
	}

	kubeConfigData, err := runtime.Encode(clientcmdlatest.Codec, &kubeConfig)
	if err != nil {
		return nil, err
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      componentName + "-hub-kubeconfig",
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"kubeconfig": kubeConfigData,
		},
	}, nil
}
