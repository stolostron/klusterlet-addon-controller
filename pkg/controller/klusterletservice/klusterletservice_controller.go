/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package klusterletservice

import (
	"context"
	"time"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/certmgr"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/component"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/connmgr"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/searchcollector"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/topology"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/workmgr"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_klusterletservice")

/**
 * USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
 * business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KlusterletService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKlusterletService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("klusterletservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KlusterletService
	err = c.Watch(&source.Kind{Type: &klusterletv1alpha1.KlusterletService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KlusterletService
	err = c.Watch(&source.Kind{Type: &klusterletv1alpha1.CertManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1alpha1.KlusterletService{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CertManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1alpha1.Tiller{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1alpha1.KlusterletService{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for Tiller to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1alpha1.ConnectionManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1alpha1.KlusterletService{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ConnectionManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &extensionsv1beta1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1alpha1.KlusterletService{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for Deployment to controller")
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKlusterletService{}

// ReconcileKlusterletService reconciles a KlusterletService object
type ReconcileKlusterletService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KlusterletService object and makes changes based on the state read
// and what is in the KlusterletService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKlusterletService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KlusterletService")

	// Fetch the KlusterletService instance
	instance := &klusterletv1alpha1.KlusterletService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	err = component.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = certmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = tiller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = connmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = workmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = topology.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = searchcollector.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.client.Update(context.TODO(), instance)
	if err != nil && errors.IsConflict(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		log.Error(err, "Fail to UPDATE instance")
		return reconcile.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}
