package addon

import (
	"context"
	"os"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	mchov1 "github.com/stolostron/multiclusterhub-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	namespace := os.Getenv("POD_NAMESPACE")
	if namespace == "" {
		klog.Info("POD_NAMESPACE not set, using 'open-cluster-management'")
		namespace = "open-cluster-management"
	}

	return add(mgr, newReconciler(mgr, namespace))
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
	if err != nil {
		return err
	}

	// Need to reconcile all KlusterletAddonConfigs when the grc InternalHubComponent is changed
	err = c.Watch(source.Kind(mgr.GetCache(), &mchov1.InternalHubComponent{},
		handler.TypedEnqueueRequestsFromMapFunc[*mchov1.InternalHubComponent](
			func(ctx context.Context, ihc *mchov1.InternalHubComponent) []reconcile.Request {
				var configList agentv1.KlusterletAddonConfigList
				if err := mgr.GetClient().List(ctx, &configList); err != nil {
					return nil
				}

				requests := make([]reconcile.Request, len(configList.Items))

				for i, config := range configList.Items {
					requests[i] = reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      config.Name,
							Namespace: config.Namespace,
						},
					}
				}

				return requests
			}),
		predicate.TypedFuncs[*mchov1.InternalHubComponent]{
			GenericFunc: func(e event.TypedGenericEvent[*mchov1.InternalHubComponent]) bool { return false },
			CreateFunc: func(e event.TypedCreateEvent[*mchov1.InternalHubComponent]) bool {
				if e.Object == nil {
					klog.Error(nil, "Create event has no runtime object to create", "event", e)
					return false
				}
				return e.Object.GetName() == "grc"
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*mchov1.InternalHubComponent]) bool {
				if e.Object == nil {
					klog.Error(nil, "Delete event has no runtime object to delete", "event", e)
					return false
				}
				return e.Object.GetName() == "grc"
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*mchov1.InternalHubComponent]) bool {
				if e.ObjectOld == nil || e.ObjectNew == nil {
					klog.Error(nil, "Update event is invalid", "event", e)
					return false
				}
				return e.ObjectOld.GetName() == "grc"
			},
		},
	))

	return err
}
