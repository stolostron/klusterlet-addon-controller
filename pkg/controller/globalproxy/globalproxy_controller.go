// Copyright Contributors to the Open Cluster Management project

package globalproxy

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

type Reconciler struct {
	runtimeClient client.Client
	kubeClient    kubernetes.Interface
	scheme        *runtime.Scheme
}

func newReconciler(mgr manager.Manager, kubeClient kubernetes.Interface) reconcile.Reconciler {
	return &Reconciler{
		runtimeClient: mgr.GetClient(),
		kubeClient:    kubeClient,
		scheme:        mgr.GetScheme(),
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := r.runtimeClient.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
		}
		return reconcile.Result{}, err
	}

	newStatus := klusterletAddonConfig.Status.DeepCopy()

	installConfigSecret, err := r.kubeClient.CoreV1().Secrets(req.Namespace).Get(ctx, fmt.Sprintf("%s-install-config", req.Name), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
			meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
				Type:    agentv1.OCPGlobalProxyDetected,
				Status:  metav1.ConditionFalse,
				Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
				Message: "The cluster is not provisioned by ACM.",
			})
			return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
		}
		return reconcile.Result{}, err
	}

	globalProxy, err := getGlobalProxyConfig(installConfigSecret)
	if err != nil {
		newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionFalse,
			Reason:  agentv1.ReasonOCPGlobalProxyDetectedFail,
			Message: err.Error(),
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	if globalProxy.NoProxy == "" && globalProxy.HTTPProxy == "" && globalProxy.HTTPSProxy == "" {
		newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionFalse,
			Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
			Message: "There is no cluster-wide proxy config in install config.",
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	if !reflect.DeepEqual(newStatus.OCPGlobalProxy, globalProxy) {
		newStatus.OCPGlobalProxy = globalProxy
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionTrue,
			Reason:  agentv1.ReasonOCPGlobalProxyDetected,
			Message: "Detected the cluster-wide proxy config in install config.",
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	// Set the ProxyPolicy of ApplicationManager in KlusterletAddonConfig to OCPGlobalProxy when
	// ApplicationManager is enabled and ProxyPolicy is not set by user.
	if klusterletAddonConfig.Spec.ApplicationManagerConfig.Enabled &&
		klusterletAddonConfig.Spec.ApplicationManagerConfig.ProxyPolicy == "" {
		klusterletAddonConfig.Spec.ApplicationManagerConfig.ProxyPolicy = agentv1.ProxyPolicyOCPGlobalProxy
		return ctrl.Result{}, r.runtimeClient.Update(context.TODO(), klusterletAddonConfig)
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) updateStatus(clusterName string, status *agentv1.KlusterletAddonConfigStatus) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
		err := r.runtimeClient.Get(context.TODO(), types.NamespacedName{Name: clusterName, Namespace: clusterName},
			klusterletAddonConfig)
		if err != nil {
			return err
		}

		newStatus := klusterletAddonConfig.Status.DeepCopy()
		newStatus.OCPGlobalProxy = status.OCPGlobalProxy
		for _, condition := range status.Conditions {
			meta.SetStatusCondition(&newStatus.Conditions, condition)
		}
		if !equality.Semantic.DeepEqual(klusterletAddonConfig.Status, newStatus) {
			klusterletAddonConfig.Status = *newStatus
			return r.runtimeClient.Status().Update(context.TODO(), klusterletAddonConfig)
		}
		return nil
	})
	return err
}

// getGlobalProxyConfig gets proxyConfig from install-config secret
func getGlobalProxyConfig(installConfigSecret *corev1.Secret) (agentv1.ProxyConfig, error) {
	proxyConfig := agentv1.ProxyConfig{}
	if len(installConfigSecret.Data) == 0 {
		return proxyConfig, fmt.Errorf("miss Data in install config secret %v", installConfigSecret.Name)
	}

	installConfigData, ok := installConfigSecret.Data["install-config.yaml"]
	if !ok {
		return proxyConfig, fmt.Errorf("miss install-config.yaml in install config secret %v", installConfigSecret.Name)
	}

	return getGlobalProxyInInstallConfig(installConfigData)
}

func getClusterNetworkCIDRs(proxyConfigRaw map[string]interface{}) ([]string, error) {
	clusterNetwork, _, err := unstructured.NestedSlice(proxyConfigRaw, "networking", "clusterNetwork")
	if err != nil {
		return []string{}, err
	}

	cidrs := make([]string, 0, len(clusterNetwork))
	for _, clusterNetworkEntry := range clusterNetwork {
		cidr, _, err := unstructured.NestedString(clusterNetworkEntry.(map[string]interface{}), "cidr")
		if err != nil {
			return []string{}, err
		}
		cidrs = append(cidrs, cidr)
	}

	return cidrs, nil
}

func getMachineNetworkCIDRs(proxyConfigRaw map[string]interface{}) ([]string, error) {
	machineNetwork, _, err := unstructured.NestedSlice(proxyConfigRaw, "networking", "machineNetwork")
	if err != nil {
		return []string{}, err
	}

	cidrs := make([]string, 0, len(machineNetwork))
	for _, machineNetworkEntry := range machineNetwork {
		cidr, _, err := unstructured.NestedString(machineNetworkEntry.(map[string]interface{}), "cidr")
		if err != nil {
			return []string{}, err
		}
		cidrs = append(cidrs, cidr)
	}

	return cidrs, nil
}

func getServiceNetworkCIDRs(proxyConfigRaw map[string]interface{}) ([]string, error) {
	cidrs, _, err := unstructured.NestedStringSlice(proxyConfigRaw, "networking", "serviceNetwork")
	if err != nil {
		return []string{}, err
	}

	return cidrs, nil
}

// For installations on Amazon Web Services (AWS), Google Cloud Platform (GCP), Microsoft Azure,
// and Red Hat OpenStack Platform (RHOSP), the Proxy object status.noProxy field is also populated
// with the instance metadata endpoint (169.254.169.254).
// ref: https://docs.openshift.com/container-platform/4.9/networking/enable-cluster-wide-proxy.html
func getMetaDataEndpoint(proxyConfigRaw map[string]interface{}) string {
	platforms := []string{"aws", "azure", "gcp"}
	for _, platform := range platforms {
		_, found, _ := unstructured.NestedString(proxyConfigRaw, "platform", platform, "region")
		if found {
			return "169.254.169.254"
		}
	}
	_, found, _ := unstructured.NestedString(proxyConfigRaw, "platform", "openstack", "externalNetwork")
	if found {
		return "169.254.169.254"
	}
	return ""
}

// refer: https://github.com/openshift/installer/blob/master/docs/design/baremetal/networking-infrastructure.md#internal-dns
func getHostName(proxyConfigRaw map[string]interface{}) (string, error) {
	clusterName, _, err := unstructured.NestedString(proxyConfigRaw, "metadata", "name")
	if err != nil {
		return "", err
	}

	baseDomain, _, err := unstructured.NestedString(proxyConfigRaw, "baseDomain")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("api-int.%s.%s", clusterName, baseDomain), nil
}

// getGlobalProxyInInstallConfig gets proxyConfig from install-config.yaml
// refer: https://github.com/openshift/installer/blob/master/docs/user/customization.md#proxy
func getGlobalProxyInInstallConfig(installConfig []byte) (agentv1.ProxyConfig, error) {
	proxyConfig := agentv1.ProxyConfig{}
	proxyConfigRaw := map[string]interface{}{}

	err := yaml.Unmarshal(installConfig, &proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}

	// proxy defined in https://github.com/openshift/installer/blob/master/pkg/types/installconfig.go
	proxyConfig.HTTPSProxy, _, err = unstructured.NestedString(proxyConfigRaw, "proxy", "httpsProxy")
	if err != nil {
		return proxyConfig, err
	}
	proxyConfig.HTTPProxy, _, err = unstructured.NestedString(proxyConfigRaw, "proxy", "httpProxy")
	if err != nil {
		return proxyConfig, err
	}

	noProxy, _, err := unstructured.NestedString(proxyConfigRaw, "proxy", "noProxy")
	if err != nil {
		return proxyConfig, err
	}

	if proxyConfig.HTTPProxy == "" && proxyConfig.HTTPSProxy == "" && noProxy == "" {
		return proxyConfig, nil
	}

	if noProxy == "*" {
		proxyConfig.NoProxy = noProxy
		return proxyConfig, nil
	}

	// The noProxy field needs to be populated with the values of some default DNS address, networking.machineNetwork[].cidr,
	// networking.clusterNetwork[].cidr, and networking.serviceNetwork[] fields from the installation configuration.
	// should be the same with the status.noProxy in proxy.config.openshift.io/cluster.
	noProxyList := sets.NewString(".cluster.local", ".svc", "localhost", "127.0.0.1")
	noProxyList.Insert(noProxy)

	clusterNetworkCIDRs, err := getClusterNetworkCIDRs(proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}
	noProxyList.Insert(clusterNetworkCIDRs...)

	machineNetworkCIDRs, err := getMachineNetworkCIDRs(proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}
	noProxyList.Insert(machineNetworkCIDRs...)

	serviceNetworkCIDRs, err := getServiceNetworkCIDRs(proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}
	noProxyList.Insert(serviceNetworkCIDRs...)

	hostName, err := getHostName(proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}
	noProxyList.Insert(hostName)

	noProxyList.Insert(getMetaDataEndpoint(proxyConfigRaw))

	noProxyList.Delete("")
	proxyConfig.NoProxy = strings.Join(noProxyList.List(), ",")
	return proxyConfig, nil
}
