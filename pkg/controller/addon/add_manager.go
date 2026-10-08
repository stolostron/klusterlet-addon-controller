package addon

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
	if err != nil {
		return err
	}

	// Watch merged-collector-config CollectorConfig changes on the hub so the search-collector
	// ManagedClusterAddOn's values annotation (see collectorconfig.go) is refreshed promptly when
	// the hub's search collection rules change, instead of only picking it up on the next
	// KlusterletAddonConfig/ManagedCluster/ManagedClusterAddOn event from the three watches above.
	//
	// The search component (and its CollectorConfig CRD) may not be installed on every hub, so a
	// failure to set up this specific watch is logged and skipped rather than failing controller
	// startup — every other addon, and search-collector's own annotation refresh via the other
	// three triggers, must keep working regardless.
	collectorConfigObj := &unstructured.Unstructured{}
	collectorConfigObj.SetGroupVersionKind(collectorConfigGVK)
	if err := c.Watch(source.Kind(mgr.GetCache(), collectorConfigObj,
		handler.TypedEnqueueRequestsFromMapFunc[*unstructured.Unstructured](
			func(ctx context.Context, obj *unstructured.Unstructured) []reconcile.Request {
				// There can be several CollectorConfig CRs on the hub (per-team, user-authored,
				// merged) — only a change to the merged one is relevant to what gets distributed.
				if obj.GetName() != mergedCollectorConfigName {
					return nil
				}
				return enqueueAllManagedClusters(ctx, mgr.GetClient())
			}),
	)); err != nil {
		klog.Warningf(
			"Could not watch CollectorConfig for instant search-collector config propagation "+
				"(the search component may not be installed on this hub): %v", err)
	}

	return nil
}

// enqueueAllManagedClusters lists every ManagedCluster and returns one reconcile.Request per
// cluster, using the same Name==Namespace==cluster-name convention the ManagedCluster watch above
// uses. A merged-collector-config change is hub-scoped, not per-cluster, so it conceptually
// affects every cluster running the search-collector addon — re-enqueuing all of them is simpler
// and safer than trying to determine which subset actually has the addon enabled.
func enqueueAllManagedClusters(ctx context.Context, c client.Client) []reconcile.Request {
	clusters := &managedclusterv1.ManagedClusterList{}
	if err := c.List(ctx, clusters); err != nil {
		klog.Errorf("Could not list ManagedClusters to re-enqueue after a CollectorConfig change: %v", err)
		return nil
	}
	requests := make([]reconcile.Request, 0, len(clusters.Items))
	for i := range clusters.Items {
		name := clusters.Items[i].GetName()
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{Name: name, Namespace: name},
		})
	}
	return requests
}
