// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	"context"
	"fmt"
	"reflect"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/bindata"
	addons "github.com/open-cluster-management/endpoint-operator/pkg/components"
	addonoperator "github.com/open-cluster-management/endpoint-operator/pkg/components/addon-operator/v1"
	appmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/endpoint-operator/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/workmgr/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/controller/clustermanagementaddon"
	"github.com/open-cluster-management/endpoint-operator/pkg/utils"
	"github.com/open-cluster-management/library-go/pkg/applier"
	"github.com/open-cluster-management/library-go/pkg/templateprocessor"
	ocinfrav1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	infrastructureConfigName = "cluster"
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
	}{
		ManagedClusterName:      klusterletaddonconfig.Name + "-" + componentName,
		ManagedClusterNamespace: klusterletaddonconfig.Name,
		ServiceAccountName:      klusterletaddonconfig.Name + "-" + componentName,
		ManagedClusterAddonName: addon.GetManagedClusterAddOnName(),
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
		"resources/hub/roles/"+componentName,
		nil,
		false,
		config,
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
	componentName string) ([]byte, error) {
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
		return nil, err
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
		return nil, fmt.Errorf("data of serviceaccount token secret does not contain token")
	}

	return token, nil
}

// getKubeAPIServerAddress - Get the API server address
func getKubeAPIServerAddress(client client.Client) (string, error) {
	infraConfig := &ocinfrav1.Infrastructure{}

	if err := client.Get(context.TODO(), types.NamespacedName{Name: infrastructureConfigName}, infraConfig); err != nil {
		return "", err
	}

	return infraConfig.Status.APIServerURL, nil
}

// newHubKubeconfigSecret -  creates a new hub-kubeconfig-secret
func newHubKubeconfigSecret(klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	componentName string,
	namespace string) (*corev1.Secret, error) {

	saToken, err := getServiceAccountToken(client, klusterletaddonconfig, componentName)
	if err != nil {
		return nil, err
	}

	kubeAPIServer, err := getKubeAPIServerAddress(client)
	if err != nil {
		return nil, err
	}

	kubeConfig := clientcmdapi.Config{
		// Define a cluster stanza based on the bootstrap kubeconfig.
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                kubeAPIServer,
			InsecureSkipTLSVerify: true,
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
