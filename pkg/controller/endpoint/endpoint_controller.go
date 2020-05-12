// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package endpoint contain the controller and the main reconcile function for the endpoints.multicloud.ibm.com
package endpoint

import (
	"context"
	"time"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
	appmgr "github.com/open-cluster-management/endpoint-operator/pkg/appmgr/v1beta1"
	certpolicycontroller "github.com/open-cluster-management/endpoint-operator/pkg/certpolicycontroller/v1beta1"
	ciscontroller "github.com/open-cluster-management/endpoint-operator/pkg/ciscontroller/v1beta1"
	component "github.com/open-cluster-management/endpoint-operator/pkg/component/v1beta1"
	connmgr "github.com/open-cluster-management/endpoint-operator/pkg/connmgr/v1beta1"
	iampolicycontroller "github.com/open-cluster-management/endpoint-operator/pkg/iampolicycontroller/v1beta1"
	policyctrl "github.com/open-cluster-management/endpoint-operator/pkg/policyctrl/v1beta1"
	searchcollector "github.com/open-cluster-management/endpoint-operator/pkg/searchcollector/v1beta1"
	workmgr "github.com/open-cluster-management/endpoint-operator/pkg/workmgr/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.CertPolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CertPolicyController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.CISController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CISController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &multicloudv1beta1.IAMPolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &multicloudv1beta1.Endpoint{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for IAMPolicyController to controller")
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
		reqLogger.Error(err, "Unable to reconcile component", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}

	tempRequeue, err = connmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile connmgr", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = workmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile workmgr", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = searchcollector.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile searchcollector", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = certpolicycontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile certpolicycontroller", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = ciscontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile ciscontroller", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = iampolicycontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile iampolicycontroller", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = policyctrl.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile policyctrl", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	_, err = appmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile appmgr", "endpointName", instance.GetName(), "endpointNamespace", instance.GetNamespace())
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
