package addon

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/utils"
	"github.com/stolostron/multicloud-operators-foundation/pkg/apis/imageregistry/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	klusterletAddonConfigAnnotationPause = "klusterletaddonconfig-pause"

	clusterImageRegistryLabel = "open-cluster-management.io/image-registry"

	// annotationNodeSelector is key name of nodeSelector annotation synced from mch
	annotationNodeSelector = "open-cluster-management/nodeSelector"

	// annotationValues is the key name of values annotation on managedClusterAddon
	annotationValues = "addon.open-cluster-management.io/values"
)

// globalValues is the values can be overridden by klusterletAddon-controller
type globalValues struct {
	Global global `json:"global,omitempty"`
}

type global struct {
	ImageOverrides map[string]string `json:"imageOverrides,omitempty"`
	NodeSelector   map[string]string `json:"nodeSelector,omitempty"`
	ProxyConfig    map[string]string `json:"proxyConfig,omitempty"`
}

func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKlusterletAddOn{client: mgr.GetClient()}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("klusterletAddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &agentv1.KlusterletAddonConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &managedclusterv1.ManagedCluster{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetName(),
							Namespace: obj.Meta.GetName(),
						},
					},
				}
			},
		)})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ManagedClusterAddOn{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetNamespace(),
							Namespace: obj.Meta.GetNamespace(),
						},
					},
				}
			},
		)},
		utils.KlusterletAddonPredicate())

	return err
}

type ReconcileKlusterletAddOn struct {
	client client.Client
}

func (r *ReconcileKlusterletAddOn) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the managedCluster instance
	managedCluster := &managedclusterv1.ManagedCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Namespace}, managedCluster); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, r.deleteAllManagedClusterAddon(request.Name)
		}
		return reconcile.Result{}, err
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.deleteAllManagedClusterAddon(managedCluster.GetName())
	}

	// Fetch the klusterletAddonConfig instance
	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, r.deleteAllManagedClusterAddon(managedCluster.GetName())
		}
		return reconcile.Result{}, err
	}

	if isPaused(klusterletAddonConfig) {
		return reconcile.Result{}, nil
	}

	nodeSelector, err := getNodeSelector(managedCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	var aggregatedErrs []error
	for addonName := range agentv1.KlusterletAddons {
		if !addonIsEnabled(addonName, klusterletAddonConfig) {
			if err := r.deleteManagedClusterAddon(addonName, managedCluster.GetName()); err != nil {
				aggregatedErrs = append(aggregatedErrs, err)
			}
			continue
		}

		imageOverrides, err := getImageOverrides(r.client, managedCluster, addonName)
		if err != nil {
			return reconcile.Result{}, err
		}
		gv := getGlobalValues(nodeSelector, imageOverrides, addonName, klusterletAddonConfig)

		if err := r.updateManagedClusterAddon(gv, addonName, managedCluster.GetName()); err != nil {
			aggregatedErrs = append(aggregatedErrs, err)
		}
	}
	if len(aggregatedErrs) != 0 {
		return reconcile.Result{}, fmt.Errorf("failed create/update addon %v", aggregatedErrs)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileKlusterletAddOn) deleteAllManagedClusterAddon(clusterName string) error {
	var aggregatedErrs []error
	for addonName, canBeDeleted := range agentv1.KlusterletAddons {
		if !canBeDeleted {
			continue
		}
		err := r.deleteManagedClusterAddon(addonName, clusterName)
		if err != nil {
			aggregatedErrs = append(aggregatedErrs, err)
		}
	}
	if len(aggregatedErrs) != 0 {
		return fmt.Errorf("failed to delelte all addons %v", aggregatedErrs)
	}
	return nil
}

func (r *ReconcileKlusterletAddOn) deleteManagedClusterAddon(addonName, clusterName string) error {
	if !agentv1.KlusterletAddons[addonName] {
		return nil
	}

	addon := &addonv1alpha1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: clusterName,
		},
	}

	err := r.client.Delete(context.TODO(), addon)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *ReconcileKlusterletAddOn) updateManagedClusterAddon(gv globalValues, addonName, clusterName string) error {
	valuesString, err := marshalGlobalValues(gv)
	if err != nil {
		return err
	}
	addon := &addonv1alpha1.ManagedClusterAddOn{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: addonName, Namespace: clusterName}, addon)
	if errors.IsNotFound(err) {
		if !agentv1.KlusterletAddons[addonName] {
			return nil
		}

		newAddon := newManagedClusterAddon(addonName, clusterName)
		if len(valuesString) != 0 {
			newAddon.SetAnnotations(map[string]string{annotationValues: valuesString})
		}

		return r.client.Create(context.TODO(), newAddon)
	}
	if err != nil {
		return err
	}

	update := false
	addon = addon.DeepCopy()
	if addon.Spec.InstallNamespace != agentv1.KlusterletAddonNamespace {
		addon.Spec.InstallNamespace = agentv1.KlusterletAddonNamespace
		update = true
	}

	if len(addon.Annotations) == 0 && len(valuesString) != 0 {
		addon.SetAnnotations(map[string]string{annotationValues: valuesString})
		update = true
	}
	if len(addon.Annotations) != 0 {
		values, existed := addon.Annotations[annotationValues]
		if !existed && len(valuesString) != 0 {
			addon.Annotations[annotationValues] = valuesString
			update = true
		}
		if existed && !reflect.DeepEqual(values, valuesString) {
			if len(valuesString) != 0 {
				addon.Annotations[annotationValues] = valuesString
			} else {
				delete(addon.Annotations, annotationValues)
			}
			update = true
		}
	}

	if !update {
		return nil
	}

	err = r.client.Update(context.TODO(), addon)
	if err != nil {
		return err
	}
	return nil
}

// isPaused returns true if the KlusterletAddonConfig instance is labeled as paused, and false otherwise
func isPaused(instance *agentv1.KlusterletAddonConfig) bool {
	a := instance.GetAnnotations()
	if len(a) == 0 {
		return false
	}

	if a[klusterletAddonConfigAnnotationPause] != "" &&
		strings.EqualFold(a[klusterletAddonConfigAnnotationPause], "true") {
		return true
	}

	return false
}

func getNodeSelector(managedCluster *managedclusterv1.ManagedCluster) (map[string]string, error) {
	var nodeSelector map[string]string
	if managedCluster.GetName() == "local-cluster" {
		annotations := managedCluster.GetAnnotations()
		if nodeSelectorString, ok := annotations[annotationNodeSelector]; ok {
			if err := json.Unmarshal([]byte(nodeSelectorString), &nodeSelector); err != nil {
				klog.Error(err, "failed to unmarshal nodeSelector annotation of cluster %v", managedCluster.GetName())
				return nodeSelector, err
			}
		}
	}

	return nodeSelector, nil
}

func getImageOverrides(client client.Client, managedCluster *managedclusterv1.ManagedCluster, addonName string) (map[string]string, error) {
	imageOverrides := map[string]string{}
	if len(managedCluster.Labels) == 0 {
		return imageOverrides, nil
	}
	imageRegistryLabelValue := managedCluster.Labels[clusterImageRegistryLabel]
	if imageRegistryLabelValue == "" {
		return imageOverrides, nil
	}

	segments := strings.Split(imageRegistryLabelValue, ".")
	if len(segments) != 2 {
		klog.Errorf("invalid format of image registry label value %v", imageRegistryLabelValue)
		return imageOverrides, fmt.Errorf("invalid format of image registry label value %v", imageRegistryLabelValue)
	}
	namespace := segments[0]
	imageRegistryName := segments[1]
	imageRegistry := &v1alpha1.ManagedClusterImageRegistry{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: imageRegistryName, Namespace: namespace}, imageRegistry)
	if err != nil {
		klog.Errorf("failed to get imageregistry %v/%v", namespace, imageRegistryName)
		return imageOverrides, err
	}
	registry := imageRegistry.Spec.Registry

	if registry == "" {
		return imageOverrides, nil
	}

	for _, imageKey := range agentv1.KlusterletAddonImageNames[addonName] {
		image, err := agentv1.GetImage(registry, imageKey)
		if err != nil {
			return imageOverrides, err
		}
		imageOverrides[imageKey] = image
	}

	return imageOverrides, nil
}

func getProxyConfig(addonName string, config *agentv1.KlusterletAddonConfig) map[string]string {
	var proxyPolicy agentv1.ProxyPolicy
	switch addonName {
	case agentv1.ApplicationAddonName:
		if !config.Spec.ApplicationManagerConfig.Enabled {
			return nil
		}
		proxyPolicy = config.Spec.ApplicationManagerConfig.ProxyPolicy
	case agentv1.CertPolicyAddonName:
		if !config.Spec.CertPolicyControllerConfig.Enabled {
			return nil
		}
		proxyPolicy = config.Spec.CertPolicyControllerConfig.ProxyPolicy
	case agentv1.IamPolicyAddonName:
		if !config.Spec.IAMPolicyControllerConfig.Enabled {
			return nil
		}
		proxyPolicy = config.Spec.IAMPolicyControllerConfig.ProxyPolicy
	case agentv1.ConfigPolicyAddonName, agentv1.PolicyFrameworkAddonName:
		if !config.Spec.PolicyController.Enabled {
			return nil
		}
		proxyPolicy = config.Spec.PolicyController.ProxyPolicy
	case agentv1.SearchAddonName:
		if !config.Spec.SearchCollectorConfig.Enabled {
			return nil
		}
		proxyPolicy = config.Spec.SearchCollectorConfig.ProxyPolicy
	}

	var proxyConfig map[string]string
	switch proxyPolicy {
	case agentv1.ProxyPolicyOCPGlobalProxy:
		proxyConfig = map[string]string{
			agentv1.HTTPProxy:  config.Status.OCPGlobalProxy.HTTPProxy,
			agentv1.HTTPSProxy: config.Status.OCPGlobalProxy.HTTPSProxy,
			agentv1.NoProxy:    config.Status.OCPGlobalProxy.NoProxy,
		}
	case agentv1.ProxyPolicyCustomProxy:
		proxyConfig = map[string]string{
			agentv1.HTTPProxy:  config.Spec.ProxyConfig.HTTPProxy,
			agentv1.HTTPSProxy: config.Spec.ProxyConfig.HTTPSProxy,
			agentv1.NoProxy:    config.Spec.ProxyConfig.NoProxy,
		}
	}
	return proxyConfig
}

func getGlobalValues(nodeSelector map[string]string,
	imageOverrides map[string]string,
	addonName string,
	config *agentv1.KlusterletAddonConfig) globalValues {
	return globalValues{
		Global: global{
			ImageOverrides: imageOverrides,
			NodeSelector:   nodeSelector,
			ProxyConfig:    getProxyConfig(addonName, config),
		},
	}
}

func newManagedClusterAddon(addonName, namespace string) *addonv1alpha1.ManagedClusterAddOn {
	return &addonv1alpha1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: namespace,
		},
		Spec: addonv1alpha1.ManagedClusterAddOnSpec{
			InstallNamespace: agentv1.KlusterletAddonNamespace,
		},
	}
}

func marshalGlobalValues(values globalValues) (string, error) {
	if len(values.Global.NodeSelector) == 0 &&
		len(values.Global.ProxyConfig) == 0 &&
		len(values.Global.ImageOverrides) == 0 {
		return "", nil
	}

	gvRaw, err := json.Marshal(values)
	if err != nil {
		return "", nil
	}
	return string(gvRaw), nil
}

func updateAnnotationValues(gv globalValues, annotationValues string) (string, error) {
	gvStr, err := marshalGlobalValues(gv)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gv. err:%v", err)
	}
	if len(gvStr) == 0 {
		return "", nil
	}
	if len(annotationValues) == 0 {
		return gvStr, nil
	}

	values := map[string]interface{}{}
	err = json.Unmarshal([]byte(annotationValues), &values)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal annotation values. err:%v", err)
	}

	oldGvRaw, ok := values["global"]
	if !ok {
		values["global"] = gv.Global
		v, err := json.Marshal(values)
		if err != nil {
			return "", fmt.Errorf("failed to marshal global values. err:%v", err)
		}

		return string(v), nil
	}

	newValues := map[string]interface{}{}
	err = json.Unmarshal([]byte(gvStr), &newValues)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal global values. err:%v", err)
	}
	newGvRaw, ok := newValues["global"]
	if !ok {
		return "", nil
	}

	mergedGv := mergeValues(oldGvRaw.(map[string]interface{}), newGvRaw.(map[string]interface{}))
	if !reflect.DeepEqual(oldGvRaw.(map[string]interface{}), mergedGv) {
		values["global"] = mergedGv
		v, err := json.Marshal(values)
		if err != nil {
			return "", fmt.Errorf("failed to marshal merged global values. err:%v", err)
		}

		return string(v), nil
	}

	return "", nil
}

// MergeValues merges the 2 given Values to a Values.
// the values of b will override that in a for the same fields.
func mergeValues(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeValues(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func addonIsEnabled(addonName string, config *agentv1.KlusterletAddonConfig) bool {
	switch addonName {
	case agentv1.ApplicationAddonName:
		return config.Spec.ApplicationManagerConfig.Enabled
	case agentv1.ConfigPolicyAddonName:
		return config.Spec.PolicyController.Enabled
	case agentv1.CertPolicyAddonName:
		return config.Spec.CertPolicyControllerConfig.Enabled
	case agentv1.IamPolicyAddonName:
		return config.Spec.IAMPolicyControllerConfig.Enabled
	case agentv1.PolicyAddonName:
		return false //  has been deprecated
	case agentv1.PolicyFrameworkAddonName:
		return config.Spec.PolicyController.Enabled
	case agentv1.SearchAddonName:
		return config.Spec.SearchCollectorConfig.Enabled
	case agentv1.WorkManagerAddonName:
		return true
	}
	return true
}
