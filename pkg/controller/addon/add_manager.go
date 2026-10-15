package addon

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
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

// collectorConfigCRDRecheckInterval is how often collectorConfigWatchRegistrar polls the
// RESTMapper for the CollectorConfig CRD when it wasn't available at controller startup.
const collectorConfigCRDRecheckInterval = 5 * time.Minute

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
	// The search component (and its CollectorConfig CRD) may not be installed on every hub.
	// Deliberately check CRD availability via the RESTMapper BEFORE calling c.Watch: a Watch()
	// call on a source.Kind only registers the informer for later — it doesn't contact the API
	// server synchronously, so a missing CRD would NOT surface as an error here. It would instead
	// surface ~2 minutes later (controller-runtime's default CacheSyncTimeout) inside
	// Controller.Start()'s WaitForSync call, which the manager treats as fatal — crashing this
	// controller's startup entirely, including the three pre-existing watches above, and with it
	// every other addon this controller manages. Checking first avoids ever registering that
	// broken informer in the shared cache.
	if ok, err := tryWatchCollectorConfig(mgr, c); err != nil {
		klog.Warningf(
			"Could not watch CollectorConfig for instant search-collector config propagation: %v", err)
	} else if !ok {
		// CRD not present yet (search not installed on this hub at startup) — rather than giving
		// up permanently (which would require restarting this controller after search gets
		// installed later), keep checking periodically in the background and register the watch
		// the first time it succeeds. c.Watch() is safe to call after the controller has already
		// started: per controller-runtime's Controller.Watch, once Started it just calls
		// src.Start(ctx, queue) directly — it does NOT re-run the blocking WaitForSync step that
		// caused the startup-crash bug above, so a still-missing CRD here only means "try again
		// next tick," never a crash.
		klog.Warningf(
			"search component's CollectorConfig CRD not found on this hub yet; will keep checking "+
				"every %s and register the instant config-change watch once it's installed "+
				"(search-collector's values annotation will still refresh via the other three "+
				"triggers above in the meantime)", collectorConfigCRDRecheckInterval)
		if err := mgr.Add(&collectorConfigWatchRegistrar{mgr: mgr, controller: c}); err != nil {
			return err
		}
	}

	return nil
}

// tryWatchCollectorConfig registers the CollectorConfig watch on c if its CRD is currently
// available on the hub (per mgr.GetRESTMapper()), returning (true, nil) on success. Returns
// (false, nil) — not an error — if the CRD simply isn't installed yet; the caller decides whether
// to retry. Any other error (e.g. the Watch call itself failing) is returned as-is.
func tryWatchCollectorConfig(mgr manager.Manager, c controller.Controller) (bool, error) {
	if _, err := mgr.GetRESTMapper().RESTMapping(
		collectorConfigGVK.GroupKind(), collectorConfigGVK.Version); err != nil {
		return false, nil
	}

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
		return false, err
	}
	return true, nil
}

// collectorConfigWatchRegistrar is a manager.Runnable that periodically retries
// tryWatchCollectorConfig until it succeeds, then stops. It exists for the case where the search
// component gets installed on this hub sometime after this controller already started — without
// it, the instant config-change watch would never be registered until this controller restarts
// (the other three triggers in add() still work in the meantime regardless; this only affects how
// promptly a hub config change propagates).
type collectorConfigWatchRegistrar struct {
	mgr        manager.Manager
	controller controller.Controller
}

// Start implements manager.Runnable.
func (r *collectorConfigWatchRegistrar) Start(ctx context.Context) error {
	return wait.PollUntilContextCancel(ctx, collectorConfigCRDRecheckInterval, false,
		func(ctx context.Context) (bool, error) {
			ok, err := tryWatchCollectorConfig(r.mgr, r.controller)
			if err != nil {
				klog.Warningf(
					"Could not watch CollectorConfig for instant search-collector config "+
						"propagation, will keep retrying: %v", err)
				return false, nil // keep polling — a transient Watch() failure shouldn't give up permanently.
			}
			if ok {
				klog.Infof(
					"search component's CollectorConfig CRD became available; registered the " +
						"instant config-change watch")
			}
			return ok, nil
		})
}

// NeedLeaderElection implements manager.LeaderElectionRunnable. This must only run on the leader:
// c.Watch() mutates the shared controller/cache state, which should not be done concurrently from
// multiple replicas.
func (r *collectorConfigWatchRegistrar) NeedLeaderElection() bool {
	return true
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
