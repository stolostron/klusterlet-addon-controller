package managedcluster

import (
	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/client-go/kubernetes"
	mcv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new ManagedCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newReconciler(mgr))
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("managedcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind(mgr.GetCache(), &mcv1.ManagedCluster{},
		&handler.TypedEnqueueRequestForObject[*mcv1.ManagedCluster]{},
		predicate.TypedFuncs[*mcv1.ManagedCluster]{
			GenericFunc: func(e event.TypedGenericEvent[*mcv1.ManagedCluster]) bool { return false },
			CreateFunc: func(e event.TypedCreateEvent[*mcv1.ManagedCluster]) bool {
				if e.Object == nil {
					log.Error(nil, "Create event has no runtime object to create", "event", e)
					return false
				}

				return hostedAddOnEnabled(e.Object) || hypershiftCluster(e.Object) || clusterClaimCluster(e.Object) || hasAnnotationCreateWithDefaultKAC(e.Object)
			},
			DeleteFunc: func(e event.TypedDeleteEvent[*mcv1.ManagedCluster]) bool {
				return false
			},
			UpdateFunc: func(e event.TypedUpdateEvent[*mcv1.ManagedCluster]) bool {
				if e.ObjectOld == nil || e.ObjectNew == nil {
					log.Error(nil, "Update event is invalid", "event", e)
					return false
				}

				return hostedAddOnEnabled(e.ObjectNew) || hypershiftCluster(e.ObjectOld) || clusterClaimCluster(e.ObjectNew) || hasAnnotationCreateWithDefaultKAC(e.ObjectNew)
			},
		}))
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &kacv1.KlusterletAddonConfig{},
		&handler.TypedEnqueueRequestForObject[*kacv1.KlusterletAddonConfig]{},
		predicate.TypedFuncs[*kacv1.KlusterletAddonConfig]{
			GenericFunc: func(e event.TypedGenericEvent[*kacv1.KlusterletAddonConfig]) bool { return false },
			CreateFunc:  func(e event.TypedCreateEvent[*kacv1.KlusterletAddonConfig]) bool { return false },
			// If the klusterletAddonConfig is deleted, we will recreate it.
			DeleteFunc: func(e event.TypedDeleteEvent[*kacv1.KlusterletAddonConfig]) bool { return true },
			UpdateFunc: func(e event.TypedUpdateEvent[*kacv1.KlusterletAddonConfig]) bool { return false },
		}))
	if err != nil {
		return err
	}

	return nil
}
