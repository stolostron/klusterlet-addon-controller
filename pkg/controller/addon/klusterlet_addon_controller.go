package addon

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	imageregistryv1alpha1 "github.com/stolostron/cluster-lifecycle-api/imageregistry/v1alpha1"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	klusterletAddonConfigAnnotationPause = "klusterletaddonconfig-pause"

	// annotationNodeSelector is key name of nodeSelector annotation on ManagedCluster
	annotationNodeSelector = "open-cluster-management/nodeSelector"

	// annotationValues is key name of tolerations annotation on ManagedCluster
	annotationTolerations = "open-cluster-management/tolerations"

	// annotationValues is the key name of values annotation on managedClusterAddon
	annotationValues = "addon.open-cluster-management.io/values"
)

// globalValues is the values can be overridden by klusterletAddon-controller
type globalValues struct {
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	Global      global              `json:"global,omitempty"`
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
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      obj.GetName(),
						Namespace: obj.GetName(),
					},
				},
			}
		}),
	)
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ManagedClusterAddOn{}},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      obj.GetNamespace(),
						Namespace: obj.GetNamespace(),
					},
				},
			}
		}),

		klusterletAddonPredicate())

	return err
}

func klusterletAddonPredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			_, existed := agentv1.KlusterletAddons[e.Object.GetName()]
			return existed
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			_, existed := agentv1.KlusterletAddons[e.Object.GetName()]
			return existed
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				klog.Error(nil, "Update event is invalid", "event", e)
				return false
			}
			_, existed := agentv1.KlusterletAddons[e.ObjectOld.GetName()]
			return existed
		},
	})
}

type ReconcileKlusterletAddOn struct {
	client client.Client
}

func (r *ReconcileKlusterletAddOn) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// Fetch the managedCluster instance
	managedCluster := &managedclusterv1.ManagedCluster{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: request.Namespace}, managedCluster); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, r.deleteAllManagedClusterAddon(ctx, request.Name)
		}
		return reconcile.Result{}, err
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, r.deleteAllManagedClusterAddon(ctx, managedCluster.GetName())
	}

	// Fetch the klusterletAddonConfig instance
	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := r.client.Get(ctx, request.NamespacedName, klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
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

	tolerations, err := getTolerations(managedCluster)
	if err != nil {
		return reconcile.Result{}, err
	}

	var aggregatedErrs []error
	for addonName, needUpdate := range agentv1.KlusterletAddons {
		if !addonIsEnabled(addonName, klusterletAddonConfig) {
			if err := r.deleteManagedClusterAddon(ctx, addonName, managedCluster.GetName()); err != nil {
				aggregatedErrs = append(aggregatedErrs, err)
			}
			continue
		}

		// work-manger addon handles by itself, does not need to update here.
		if !needUpdate {
			continue
		}

		imageOverrides, err := getImageOverrides(managedCluster, addonName)
		if err != nil {
			return reconcile.Result{}, err
		}
		gv := getGlobalValues(tolerations, nodeSelector, imageOverrides, addonName, klusterletAddonConfig)

		if err := r.updateManagedClusterAddon(ctx, gv, addonName, managedCluster.GetName()); err != nil {
			aggregatedErrs = append(aggregatedErrs, err)
		}
	}
	if len(aggregatedErrs) != 0 {
		return reconcile.Result{}, fmt.Errorf("failed create/update addon %v", aggregatedErrs)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileKlusterletAddOn) deleteAllManagedClusterAddon(ctx context.Context, clusterName string) error {
	var aggregatedErrs []error
	for addonName := range agentv1.KlusterletAddons {
		err := r.deleteManagedClusterAddon(ctx, addonName, clusterName)
		if err != nil {
			aggregatedErrs = append(aggregatedErrs, err)
		}
	}
	if len(aggregatedErrs) != 0 {
		return fmt.Errorf("failed to delelte all addons %v", aggregatedErrs)
	}
	return nil
}

func (r *ReconcileKlusterletAddOn) deleteManagedClusterAddon(ctx context.Context, addonName, clusterName string) error {
	addon := &addonv1alpha1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: clusterName,
		},
	}

	err := r.client.Delete(ctx, addon)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (r *ReconcileKlusterletAddOn) updateManagedClusterAddon(ctx context.Context, gv globalValues, addonName, clusterName string) error {
	valuesString, err := marshalGlobalValues(gv)
	if err != nil {
		return err
	}
	addon := &addonv1alpha1.ManagedClusterAddOn{}
	err = r.client.Get(ctx, types.NamespacedName{Name: addonName, Namespace: clusterName}, addon)
	if errors.IsNotFound(err) {
		if !agentv1.KlusterletAddons[addonName] {
			return nil
		}

		newAddon := newManagedClusterAddon(addonName, clusterName)
		if len(valuesString) != 0 {
			newAddon.SetAnnotations(map[string]string{annotationValues: valuesString})
		}

		return r.client.Create(ctx, newAddon)
	}
	if err != nil {
		return err
	}

	update := false
	addon = addon.DeepCopy()

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

	err = r.client.Update(ctx, addon)
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
	nodeSelector := map[string]string{}

	nodeSelectorString, ok := managedCluster.Annotations[annotationNodeSelector]
	if !ok {
		return nodeSelector, nil
	}

	if err := json.Unmarshal([]byte(nodeSelectorString), &nodeSelector); err != nil {
		return nil, fmt.Errorf("invalid nodeSelector annotation of cluster %s, %v", managedCluster.Name, err)
	}

	if err := validateNodeSelector(nodeSelector); err != nil {
		return nil, fmt.Errorf("invalid nodeSelector annotation of cluster %s, %v", managedCluster.Name, err)
	}

	return nodeSelector, nil
}

// refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go#L3498
func validateNodeSelector(nodeSelector map[string]string) error {
	errs := []error{}
	for key, val := range nodeSelector {
		if errMsgs := validation.IsQualifiedName(key); len(errMsgs) != 0 {
			errs = append(errs, fmt.Errorf(strings.Join(errMsgs, ";")))
		}
		if errMsgs := validation.IsValidLabelValue(val); len(errMsgs) != 0 {
			errs = append(errs, fmt.Errorf(strings.Join(errMsgs, ";")))
		}
	}
	return utilerrors.NewAggregate(errs)
}

func getTolerations(cluster *managedclusterv1.ManagedCluster) ([]corev1.Toleration, error) {
	tolerations := []corev1.Toleration{}

	tolerationsString, ok := cluster.Annotations[annotationTolerations]
	if !ok {
		return tolerations, nil
	}

	if err := json.Unmarshal([]byte(tolerationsString), &tolerations); err != nil {
		return nil, fmt.Errorf("invalid tolerations annotation of cluster %s, %v", cluster.Name, err)
	}

	if err := validateTolerations(tolerations); err != nil {
		return nil, fmt.Errorf("invalid tolerations annotation of cluster %s, %v", cluster.Name, err)
	}

	return tolerations, nil
}

// refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/validation/validation.go#L3330
func validateTolerations(tolerations []corev1.Toleration) error {
	errs := []error{}
	for _, toleration := range tolerations {
		// validate the toleration key
		if len(toleration.Key) > 0 {
			if errMsgs := validation.IsQualifiedName(toleration.Key); len(errMsgs) != 0 {
				errs = append(errs, fmt.Errorf(strings.Join(errMsgs, ";")))
			}
		}

		// empty toleration key with Exists operator and empty value means match all taints
		if len(toleration.Key) == 0 && toleration.Operator != corev1.TolerationOpExists {
			if len(toleration.Operator) == 0 {
				errs = append(errs, fmt.Errorf(
					"operator must be Exists when `key` is empty, which means \"match all values and all keys\""))
			}
		}

		if toleration.TolerationSeconds != nil && toleration.Effect != corev1.TaintEffectNoExecute {
			errs = append(errs, fmt.Errorf("effect must be 'NoExecute' when `tolerationSeconds` is set"))
		}

		// validate toleration operator and value
		switch toleration.Operator {
		// empty operator means Equal
		case corev1.TolerationOpEqual, "":
			if errMsgs := validation.IsValidLabelValue(toleration.Value); len(errMsgs) != 0 {
				errs = append(errs, fmt.Errorf(strings.Join(errMsgs, ";")))
			}
		case corev1.TolerationOpExists:
			if len(toleration.Value) > 0 {
				errs = append(errs, fmt.Errorf("value must be empty when `operator` is 'Exists'"))
			}
		default:
			errs = append(errs, fmt.Errorf("the operator %q is not supported", toleration.Operator))
		}

		// validate toleration effect, empty toleration effect means match all taint effects
		if len(toleration.Effect) > 0 {
			switch toleration.Effect {
			case corev1.TaintEffectNoSchedule, corev1.TaintEffectPreferNoSchedule, corev1.TaintEffectNoExecute:
				// allowed values are NoSchedule, PreferNoSchedule and NoExecute
			default:
				errs = append(errs, fmt.Errorf("the effect %q is not supported", toleration.Effect))
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

func getImageOverrides(managedCluster *managedclusterv1.ManagedCluster, addonName string) (map[string]string, error) {
	imageOverrides := map[string]string{}
	if len(managedCluster.Annotations) == 0 {
		return imageOverrides, nil
	}

	if _, ok := managedCluster.Annotations[imageregistryv1alpha1.ClusterImageRegistriesAnnotation]; !ok {
		return imageOverrides, nil
	}

	for _, imageKey := range agentv1.KlusterletAddonImageNames[addonName] {
		image, err := agentv1.GetImage(managedCluster, imageKey)
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

func getGlobalValues(tolerations []corev1.Toleration,
	nodeSelector map[string]string,
	imageOverrides map[string]string,
	addonName string,
	config *agentv1.KlusterletAddonConfig) globalValues {
	return globalValues{
		Tolerations: tolerations,
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
