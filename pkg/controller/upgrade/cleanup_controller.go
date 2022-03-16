package upgrade

import (
	"context"
	"fmt"
	"time"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/utils"
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
	return &ReconcileUpgrade{client: mgr.GetClient()}
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
		utils.KlusterletAddonPredicate())

	return err
}

type ReconcileCleanup struct {
	client client.Client
}

func (r *ReconcileCleanup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the addon instance
	addon := &addonv1alpha1.ManagedClusterAddOn{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, addon); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !addon.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	if _, ok := agentv1.KlusterletAddonComponentNames[addon.GetName()]; !ok {
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
		return reconcile.Result{}, r.cleanupDeprecatedResources(addon)
	}

	if !meta.IsStatusConditionTrue(conditions, "Available") {
		return reconcile.Result{Requeue: true, RequeueAfter: 10 * time.Second}, nil
	}

	if !meta.IsStatusConditionTrue(conditions, "AddonManifestApplied") {
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

	return reconcile.Result{}, r.cleanupDeprecatedAddon(addon)
}

func (r *ReconcileCleanup) addonOperatorUpgradeCompleted(clusterName string) (bool, error) {
	addonOperatorWork := &manifestworkv1.ManifestWork{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Namespace: clusterName, Name: manifestWorkName(clusterName, klusterletAddonOperator)},
		addonOperatorWork); err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}
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

func (r *ReconcileCleanup) cleanupDeprecatedResources(addon *addonv1alpha1.ManagedClusterAddOn) error {
	errs := []error{}

	if err := r.cleanupDeprecatedAddon(addon); err != nil {
		errs = append(errs, err)
	}
	if err := r.cleanupDeprecatedManifestWorks(addon); err != nil {
		errs = append(errs, err)
	}
	if err := r.cleanupDeprecatedRoleBinding(addon); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		return fmt.Errorf("failed to clean up resources. %v", errs)
	}

	return nil
}

func (r *ReconcileCleanup) cleanupDeprecatedAddon(addon *addonv1alpha1.ManagedClusterAddOn) error {
	for _, addonName := range agentv1.DeprecatedManagedClusterAddons {
		if addonName == addon.GetName() {
			if err := r.client.Delete(context.TODO(), addon, &client.DeleteOptions{}); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				klog.Errorf("failed to delete addon %v. %v", addon.GetName(), err)
				return err
			}
			return nil
		}
	}

	return nil
}

func (r *ReconcileCleanup) cleanupDeprecatedRoleBinding(addon *addonv1alpha1.ManagedClusterAddOn) error {
	componentName := agentv1.KlusterletAddonComponentNames[addon.GetName()]
	if componentName == "" {
		return nil
	}
	clusterName := addon.GetNamespace()
	addonRoleBindingName := roleBindingName(clusterName, componentName)
	addonRoleBinding := &v1.RoleBinding{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: clusterName, Name: addonRoleBindingName}, addonRoleBinding); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("failed to get addon roleBinding %v. %v", addonRoleBindingName, err)
		return err
	}
	if err := r.client.Delete(context.TODO(), addonRoleBinding, &client.DeleteOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("failed to delete addon roleBinding %v. %v", addonRoleBindingName, err)
		return err
	}

	return nil
}

func (r *ReconcileCleanup) cleanupDeprecatedManifestWorks(addon *addonv1alpha1.ManagedClusterAddOn) error {
	componentName := agentv1.KlusterletAddonComponentNames[addon.GetName()]
	if componentName == "" {
		return nil
	}
	clusterName := addon.GetClusterName()

	if err := r.deleteManifestWork(clusterName, agentManifestWorkName(clusterName, componentName)); err != nil {
		return err
	}

	for _, agentWorkName := range agentv1.DeprecatedAgentManifestworks {
		work := &manifestworkv1.ManifestWork{}
		err := r.client.Get(context.TODO(), types.NamespacedName{
			Namespace: clusterName,
			Name:      manifestWorkName(clusterName, agentWorkName),
		}, work)
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			return err
		}
		return nil
	}

	if err := r.deleteManifestWork(clusterName, manifestWorkName(clusterName, klusterletAddonCRDs)); err != nil {
		return err
	}
	if err := r.deleteManifestWork(clusterName, manifestWorkName(clusterName, klusterletAddonOperator)); err != nil {
		return err
	}
	return nil

}

func (r *ReconcileCleanup) deleteManifestWork(clusterName, name string) error {
	work := &manifestworkv1.ManifestWork{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: clusterName,
		Name:      name,
	}, work)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return r.client.Delete(context.TODO(), work, &client.DeleteOptions{})

}
