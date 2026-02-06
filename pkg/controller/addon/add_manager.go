package addon

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
)

func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newReconciler(mgr))
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("klusterletAddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &agentv1.KlusterletAddonConfig{},
		&handler.TypedEnqueueRequestForObject[*agentv1.KlusterletAddonConfig]{}))
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &managedclusterv1.ManagedCluster{},
		handler.TypedEnqueueRequestsFromMapFunc[*managedclusterv1.ManagedCluster](
			func(ctx context.Context, cluster *managedclusterv1.ManagedCluster) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      cluster.GetName(),
							Namespace: cluster.GetName(),
						},
					},
				}
			}),
	))
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &addonv1alpha1.ManagedClusterAddOn{},
		handler.TypedEnqueueRequestsFromMapFunc[*addonv1alpha1.ManagedClusterAddOn](
			func(ctx context.Context, addon *addonv1alpha1.ManagedClusterAddOn) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      addon.GetNamespace(),
							Namespace: addon.GetNamespace(),
						},
					},
				}
			}),
		predicate.TypedFuncs[*addonv1alpha1.ManagedClusterAddOn]{
			GenericFunc: func(e event.TypedGenericEvent[*addonv1alpha1.ManagedClusterAddOn]) bool { return false },
			CreateFunc: func(e event.TypedCreateEvent[*addonv1alpha1.ManagedClusterAddOn]) bool {
				if e.Object == nil {
					klog.Error(nil, "Create event has no runtime object to create", "event", e)
					return false
				}
				_, existed := agentv1.KlusterletAddons[e.Object.GetName()]
				return existed
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*addonv1alpha1.ManagedClusterAddOn]) bool {
				if e.Object == nil {
					klog.Error(nil, "Delete event has no runtime object to delete", "event", e)
					return false
				}
				_, existed := agentv1.KlusterletAddons[e.Object.GetName()]
				return existed
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*addonv1alpha1.ManagedClusterAddOn]) bool {
				if e.ObjectOld == nil || e.ObjectNew == nil {
					klog.Error(nil, "Update event is invalid", "event", e)
					return false
				}
				_, existed := agentv1.KlusterletAddons[e.ObjectOld.GetName()]
				return existed
			},
		}))

	return err
}
