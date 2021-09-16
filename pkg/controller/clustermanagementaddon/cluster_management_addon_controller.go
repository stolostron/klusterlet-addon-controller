// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package clustermanagementaddon

import (
	"context"
	"fmt"

	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clustermanagementaddon")

// Add creates a new ClusterManagementAddOn Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
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
		addons.NewAddonNamePredicate())
	if err != nil {
		return err
	}

	return nil
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
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling ClusterManagementAddOn")

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
