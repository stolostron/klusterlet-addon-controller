// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package clustermanagementaddon

import (
	"context"
	"fmt"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clustermanagementaddon")

// Add creates a new ClusterManagementAddOn Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileClusterManagementAddOn{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clustermanagementaddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterManagementAddon
	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ClusterManagementAddOn{}}, &handler.EnqueueRequestForObject{},
		klusterletAddonPredicate())
	if err != nil {
		return err
	}

	return nil
}
func klusterletAddonPredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			return agentv1.KlusterletAddons[e.Meta.GetName()]
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			return agentv1.KlusterletAddons[e.Meta.GetName()]
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld == nil || e.MetaNew == nil ||
				e.ObjectOld == nil || e.ObjectNew == nil {
				klog.Error(nil, "Update event is invalid", "event", e)
				return false
			}
			return agentv1.KlusterletAddons[e.MetaNew.GetName()]
		},
	})
}

// blank assignment to verify that ReconcileClusterManagementAddOn implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClusterManagementAddOn{}

// ReconcileClusterManagementAddOn reconciles a ClusterManagementAddOn object
type ReconcileClusterManagementAddOn struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManagedClusterAddOn object and makes changes based on the state read
// and status in the ManagedClusterAddOn
func (r *ReconcileClusterManagementAddOn) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ManagedClusterAddOn instance
	clusterManagementAddOn := &addonv1alpha1.ClusterManagementAddOn{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, clusterManagementAddOn); err != nil {
		if errors.IsNotFound(err) {
			clusterManagementAddonMeta := ClusterManagementAddOnMap[request.Name]
			clusterManagementAddon := newClusterManagementAddon(request.Name, clusterManagementAddonMeta)
			if err := r.client.Create(context.TODO(), clusterManagementAddon); err != nil {
				log.Error(err, fmt.Sprintf("Failed to create %s clustermanagementaddon ", request.Name))
				return reconcile.Result{}, err
			}
			log.Info(fmt.Sprintf("Create %s clustermanagementaddon", request.Name))
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if clusterManagementAddOn != nil {
		if err := updateClusterManagementAddOn(r.client, request.Name, clusterManagementAddOn); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
