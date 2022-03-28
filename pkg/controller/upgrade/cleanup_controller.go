package upgrade

import (
	"context"
	"fmt"
	"time"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func CleanupAdd(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return cleanupAdd(mgr, newCleanupReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newCleanupReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCleanup{client: mgr.GetClient()}
}

func cleanupAdd(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("cleanup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ManagedClusterAddOn{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetName(),
							Namespace: obj.Meta.GetNamespace(),
						},
					},
				}
			},
		)},
		upgradePredicate())

	return err
}

type ReconcileCleanup struct {
	client client.Client
}

func (r *ReconcileCleanup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the addon instance
	addon := &addonv1alpha1.ManagedClusterAddOn{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, addon)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	if err != nil {
		return reconcile.Result{}, err
	}

	if !addon.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	if err = r.cleanupAddonOwnerRef(addon); err != nil {
		return reconcile.Result{}, err
	}

	if _, ok := agentv1.DeprecatedAddonComponentNames[addon.GetName()]; !ok {
		return reconcile.Result{}, nil
	}

	upgradeCompleted, err := r.addonOperatorUpgradeCompleted(addon.GetNamespace())
	if err != nil {
		return reconcile.Result{}, err
	}
	if !upgradeCompleted {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	conditions := addon.Status.Conditions
	if addon.GetName() == agentv1.PolicyAddonName {
		availableCondition := meta.FindStatusCondition(conditions, "Available")
		if availableCondition == nil {
			return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}
		if availableCondition.Status == metav1.ConditionTrue {
			return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}

		sub := time.Since(availableCondition.LastTransitionTime.Time)
		if sub < 5*time.Minute {
			return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}

		deleted, err := r.cleanupDeprecatedResources(addon)
		if err != nil {
			return reconcile.Result{}, err
		}
		if !deleted {
			return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
		}
		return reconcile.Result{}, nil
	}

	if !meta.IsStatusConditionTrue(conditions, "Available") {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	if !meta.IsStatusConditionTrue(conditions, "ManifestApplied") {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	condition := meta.FindStatusCondition(conditions, "RegistrationApplied")
	if condition == nil {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}
	if condition.Status != metav1.ConditionTrue {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	sub := time.Since(condition.LastTransitionTime.Time)
	if sub < 5*time.Minute {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	deleted, err := r.cleanupDeprecatedResources(addon)
	if err != nil {
		return reconcile.Result{}, err
	}
	if !deleted {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileCleanup) addonOperatorUpgradeCompleted(clusterName string) (bool, error) {
	addonOperatorWork := &manifestworkv1.ManifestWork{}
	err := r.client.Get(context.TODO(),
		types.NamespacedName{Namespace: clusterName, Name: manifestWorkName(clusterName, klusterletAddonOperator)},
		addonOperatorWork)
	if errors.IsNotFound(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	labels := addonOperatorWork.GetLabels()
	if _, ok := labels[agentv1.UpgradeLabel]; !ok {
		return false, nil
	}

	for _, condition := range addonOperatorWork.Status.Conditions {
		if condition.Type == "Applied" || condition.Type == "Available" {
			if condition.Status != "True" {
				return false, nil
			}
		}
	}

	for _, manifest := range addonOperatorWork.Status.ResourceStatus.Manifests {
		if manifest.ResourceMeta.Resource == "deployments" {
			for _, condition := range manifest.Conditions {
				if condition.Type == "Applied" || condition.Type == "Available" || condition.Type == "StatusFeedbackSynced" {
					if condition.Status != "True" {
						return false, nil
					}
				}
			}
		}
	}
	return true, nil
}

func (r *ReconcileCleanup) cleanupDeprecatedResources(addon *addonv1alpha1.ManagedClusterAddOn) (bool, error) {
	errs := []error{}

	if err := r.cleanupDeprecatedAddon(addon); err != nil {
		errs = append(errs, err)
	}
	if err := r.cleanupAgentManifestWork(addon); err != nil {
		errs = append(errs, err)
	}
	if err := r.cleanupDeprecatedRoleBinding(addon); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		return false, fmt.Errorf("failed to clean up resources. %v", errs)
	}

	return r.cleanupOperatorManifestWorks(addon.Namespace)
}

func (r *ReconcileCleanup) cleanupDeprecatedAddon(addon *addonv1alpha1.ManagedClusterAddOn) error {
	if agentv1.PolicyAddonName != addon.GetName() {
		return nil
	}

	return r.client.Delete(context.TODO(), addon, &client.DeleteOptions{})
}

func (r *ReconcileCleanup) cleanupDeprecatedRoleBinding(addon *addonv1alpha1.ManagedClusterAddOn) error {
	componentName := agentv1.DeprecatedAddonComponentNames[addon.GetName()]
	if componentName == "" {
		return nil
	}
	clusterName := addon.GetNamespace()
	addonRoleBindingName := roleBindingName(clusterName, componentName)
	addonRoleBinding := &v1.RoleBinding{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: clusterName, Name: addonRoleBindingName}, addonRoleBinding)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		klog.Errorf("failed to get addon roleBinding %v. %v", addonRoleBindingName, err)
		return err
	}
	err = r.client.Delete(context.TODO(), addonRoleBinding, &client.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		klog.Errorf("failed to delete addon roleBinding %v. %v", addonRoleBindingName, err)
		return err
	}

	return nil
}

func (r *ReconcileCleanup) cleanupAgentManifestWork(addon *addonv1alpha1.ManagedClusterAddOn) error {
	componentName := agentv1.DeprecatedAddonComponentNames[addon.GetName()]
	if componentName == "" {
		return nil
	}

	return r.deleteManifestWork(addon.Namespace, agentManifestWorkName(addon.Namespace, componentName))
}

func (r *ReconcileCleanup) deleteManifestWork(namespace, name string) error {
	work := &manifestworkv1.ManifestWork{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, work)
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	return r.client.Delete(context.TODO(), work, &client.DeleteOptions{})
}

// check if the deprecated addon agent manifestWorks are deleted, if yes, delete operator and crds manifestWorks.
func (r *ReconcileCleanup) cleanupOperatorManifestWorks(namespace string) (bool, error) {
	for _, agentWorkName := range agentv1.DeprecatedAgentManifestworks {
		work := &manifestworkv1.ManifestWork{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      manifestWorkName(namespace, agentWorkName),
		}, work)
		if errors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return false, err
		}

		return false, nil
	}

	if err := r.deleteManifestWork(namespace, manifestWorkName(namespace, klusterletAddonCRDs)); err != nil {
		return false, err
	}
	if err := r.deleteManifestWork(namespace, manifestWorkName(namespace, klusterletAddonOperator)); err != nil {
		return false, err
	}
	return true, nil
}

func (r *ReconcileCleanup) cleanupAddonOwnerRef(addon *addonv1alpha1.ManagedClusterAddOn) error {
	if len(addon.OwnerReferences) == 0 {
		return nil
	}

	newAddon := addon.DeepCopy()
	newOwnerReferences := []metav1.OwnerReference{}
	needUpdate := false
	for _, owner := range addon.OwnerReferences {
		if owner.Kind == "KlusterletAddonConfig" {
			needUpdate = true
			continue
		}
		newOwnerReferences = append(newOwnerReferences, owner)
	}

	if !needUpdate {
		return nil
	}

	newAddon.SetOwnerReferences(newOwnerReferences)
	return r.client.Update(context.TODO(), newAddon, &client.UpdateOptions{})
}
