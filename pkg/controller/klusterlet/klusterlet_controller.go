// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package klusterlet contain the controller and the main reconcile function for the klusterlets.agent.open-cluster-management.io
package klusterlet

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

	klusterletv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1beta1"
	appmgr "github.com/open-cluster-management/endpoint-operator/pkg/appmgr/v1beta1"
	certpolicycontroller "github.com/open-cluster-management/endpoint-operator/pkg/certpolicycontroller/v1beta1"
	ciscontroller "github.com/open-cluster-management/endpoint-operator/pkg/ciscontroller/v1beta1"
	component "github.com/open-cluster-management/endpoint-operator/pkg/component/v1beta1"
	connmgr "github.com/open-cluster-management/endpoint-operator/pkg/connmgr/v1beta1"
	iampolicycontroller "github.com/open-cluster-management/endpoint-operator/pkg/iampolicycontroller/v1beta1"
	policyctrl "github.com/open-cluster-management/endpoint-operator/pkg/policyctrl/v1beta1"
	searchcollector "github.com/open-cluster-management/endpoint-operator/pkg/searchcollector/v1beta1"
	workmgr "github.com/open-cluster-management/endpoint-operator/pkg/workmgr/v1beta1"
)

var log = logf.Log.WithName("controller_klusterlet")

// Add creates a new Klusterlet Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKlusterlet{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("klusterlet-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Klusterlet
	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.Klusterlet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources and requeue the owner Klusterlet
	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.ApplicationManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ApplicationManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.CertPolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CertPolicyController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.CISController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for CISController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.IAMPolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for IAMPolicyController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.ConnectionManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for ConnectionManager to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.PolicyController{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for PolicyController to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.SearchCollector{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for SearchCollector to controller")
		return err
	}

	err = c.Watch(&source.Kind{Type: &klusterletv1beta1.WorkManager{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1beta1.Klusterlet{},
	})
	if err != nil {
		log.Error(err, "Fail to add Watch for WorkManager to controller")
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKlusterlet implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKlusterlet{}

// ReconcileKlusterlet reconciles a Klusterlet object
type ReconcileKlusterlet struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Klusterlet object and makes changes based on the state read
// and what is in the Klusterlet.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKlusterlet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Klusterlet")

	// Fetch the Klusterlet instance
	instance := &klusterletv1beta1.Klusterlet{}
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
		reqLogger.Error(err, "Unable to reconcile component", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}

	tempRequeue, err = connmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile connmgr", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = workmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile workmgr", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = searchcollector.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile searchcollector", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = certpolicycontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile certpolicycontroller", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = ciscontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile ciscontroller", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = iampolicycontroller.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile iampolicycontroller", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	tempRequeue, err = policyctrl.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile policyctrl", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
		return reconcile.Result{}, err
	}
	requeue = requeue || tempRequeue

	_, err = appmgr.Reconcile(instance, r.client, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Unable to reconcile appmgr", "klusterletName", instance.GetName(), "klusterletNamespace", instance.GetNamespace())
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
