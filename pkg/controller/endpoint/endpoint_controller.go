// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package endpoint contain the controller and the main reconcile function for the endpoints.multicloud.ibm.com
package endpoint

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	appmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/appmgr/v1beta1"
	certmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/certmgr/v1beta1"
	component "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/component/v1beta1"
	connmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/connmgr/v1beta1"
	monitoring "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/monitoring/v1beta1"
	policyctrl "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/policyctrl/v1beta1"
	searchcollector "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/searchcollector/v1beta1"
	serviceregistry "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/serviceregistry/v1beta1"
	topology "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/topology/v1beta1"
	workmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/workmgr/v1beta1"
)

var log = logf.Log.WithName("controller_endpoint")

// Add creates a new Endpoint Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileEndpoint{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("endpoint-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Endpoint
	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.Endpoint{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources and requeue the owner Endpoint
	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.ApplicationManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ApplicationManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.CertManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CertManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.ConnectionManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ConnectionManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.PolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for PolicyController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.SearchCollector{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for SearchCollector to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.ServiceRegistry{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ServiceRegistry to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.TopologyCollector{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for TopologyCollector to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.WorkManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for WorkManager to controller")
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileEndpoint implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEndpoint{}

// ReconcileEndpoint reconciles a Endpoint object
type ReconcileEndpoint struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Endpoint object and makes changes based on the state read
// and what is in the Endpoint.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileEndpoint) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Endpoint")

	// Fetch the Endpoint instance
	instance := &multicloudv1beta1.Endpoint{}
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

	var requeue, tempRequeue bool

	err = component.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	tempRequeue, err = certmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = connmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = workmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = searchcollector.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = policyctrl.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = serviceregistry.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = topology.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = monitoring.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	_, err = appmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.client.Update(context.TODO(), instance)
	if err != nil && errors.IsConflict(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		reqLogger.Error(err, "Fail to UPDATE instance")
		return reconcile.Result{}, err
	}

	if instance.GetDeletionTimestamp() != nil || requeue {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Minute}, nil
}
