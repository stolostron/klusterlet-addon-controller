// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	"context"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/utils"
)

var log = logf.Log.WithName("controller_klusterletaddon")

const (
	KlusterletAddonConfigAnnotationPause = "klusterletaddonconfig-pause"
)

// Add creates a new KlusterletAddon Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKlusterletAddon{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("klusterletaddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Klusterlet
	err = c.Watch(&source.Kind{Type: &agentv1.KlusterletAddonConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch for changes to secondary resource Pods and requeue the owner ClusterDeployment
	err = c.Watch(
		&source.Kind{Type: &managedclusterv1.ManagedCluster{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetName(), // only handle klusterlet with name/namespaxe same as managedCluster's name
							Namespace: obj.Meta.GetName(),
						},
					},
				}
			},
		)},
	)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKlusterletAddon implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKlusterletAddon{}

// ReconcileKlusterletAddon reconciles a Klusterlet object
type ReconcileKlusterletAddon struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KlusterletAddonConfig object
// and makes changes based on the state read and what is in the KlusterletAddonConfig.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKlusterletAddon) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KlusterletAddonConfig")

	// Fetch the ManagedCluster instance
	managedCluster := &managedclusterv1.ManagedCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Namespace}, managedCluster); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fetch the Klusterletaddonconfig instance
	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			// remove finalizer on ManagedCluster if klusterlet not found
			if utils.HasFinalizer(managedCluster, KlusterletAddonFinalizer) {
				utils.RemoveFinalizer(managedCluster, KlusterletAddonFinalizer)
				if err := r.client.Update(context.TODO(), managedCluster); err != nil {
					return reconcile.Result{}, err
				}
			}
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if klusterletAddonConfig.DeletionTimestamp != nil {
		// if ManagedCluster not online, force delete all manifestwork
		removeFinalizers := !IsManagedClusterOnline(managedCluster)

		// delete & wait all CRs
		if isCompleted, err := deleteManifestWorkCRs(klusterletAddonConfig, r.client, removeFinalizers); err != nil {
			reqLogger.Error(err, "Fail to delete all ManifestWorks for Addon CRs")
			return reconcile.Result{}, err
		} else if !isCompleted {
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}

		// delete & wait component Operator
		if isCompleted, err := deleteManifestWorkHelper(
			klusterletAddonConfig.Name+KlusterletAddonOperatorPostfix,
			klusterletAddonConfig.Namespace,
			r.client,
			removeFinalizers,
		); err != nil {
			reqLogger.Error(err, "Fail to delete ManifestWork of Klusterlet Addon Operator")
			return reconcile.Result{}, err
		} else if !isCompleted {
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}

		// delete & wait CRDs
		if isCompleted, err := deleteManifestWorkHelper(
			klusterletAddonConfig.Name+KlusterletAddonCRDsPostfix,
			klusterletAddonConfig.Namespace,
			r.client,
			removeFinalizers,
		); err != nil {
			reqLogger.Error(err, "Fail to delete ManifestWork of CRDs")
			return reconcile.Result{}, err
		} else if !isCompleted {
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, err
		}

		utils.RemoveFinalizer(klusterletAddonConfig, KlusterletAddonFinalizer)
		if err := r.client.Update(context.TODO(), klusterletAddonConfig); err != nil {
			return reconcile.Result{}, err
		}
		// remove finalizer on managedCluster when all things are removed
		utils.RemoveFinalizer(managedCluster, KlusterletAddonFinalizer)
		if err := r.client.Update(context.TODO(), managedCluster); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	utils.AddFinalizer(managedCluster, KlusterletAddonFinalizer)
	if err := r.client.Update(context.TODO(), managedCluster); err != nil && errors.IsConflict(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		reqLogger.Error(err, "Fail to UPDATE managedCluster")
		return reconcile.Result{}, err
	}

	utils.AddFinalizer(klusterletAddonConfig, KlusterletAddonFinalizer)
	if err := r.client.Update(context.TODO(), klusterletAddonConfig); err != nil && errors.IsConflict(err) {
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	} else if err != nil {
		reqLogger.Error(err, "Fail to UPDATE KlusterletAddonConfig")
		return reconcile.Result{}, err
	}

	// Delete klusterletAddonConfig if ManagedCluster is in deletion
	if managedCluster.DeletionTimestamp != nil && klusterletAddonConfig.DeletionTimestamp == nil {
		if err := r.client.Delete(context.TODO(), klusterletAddonConfig); err != nil {
			reqLogger.Error(err, "Fail to trigger deletion of KlusterletAddonConfig")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if isPaused(klusterletAddonConfig) {
		reqLogger.Info("KlusterletAddonConfig reconciliation is paused. Nothing more to do.")
		return reconcile.Result{}, nil
	}

	// Fill with default imagePullSecret if empty
	if klusterletAddonConfig.Spec.ImagePullSecret == "" {
		klusterletAddonConfig.Spec.ImagePullSecret = os.Getenv("DEFAULT_IMAGE_PULL_SECRET")
	}

	// Create manifest work for crds
	if err := createManifestWorkCRD(klusterletAddonConfig, managedCluster.Status.Version.Kubernetes, r); err != nil {
		reqLogger.Error(err, "Fail to create manifest work for CRD")
		return reconcile.Result{}, err
	}

	// Create manifest work for Klusterlet Addon operator
	if err := createManifestWorkComponentOperator(klusterletAddonConfig, r); err != nil {
		reqLogger.Error(err, "Fail to create manifest work for klusterlet addon opearator")
		return reconcile.Result{}, err
	}

	// sync manifestWork for component crs according to klusterletAddonConfig enable/disable settings
	if err := syncManifestWorkCRs(klusterletAddonConfig, r); err != nil {
		reqLogger.Error(err, "Fail to create manifest work for CRs")
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Minute}, nil
}

// IsManagedClusterOnline - if cluster is online returns true otherwise returns false
func IsManagedClusterOnline(managedCluster *managedclusterv1.ManagedCluster) bool {
	for _, condition := range managedCluster.Status.Conditions {
		if condition.Type == managedclusterv1.ManagedClusterConditionAvailable { //not sure which condition is valid
			if condition.Status == "True" {
				return true
			}
		}
	}

	return false
}

// deleteManifestWorkHelper returns true if object is not found
func deleteManifestWorkHelper(name, namespace string, client client.Client, removeFinalizers bool) (bool, error) {
	err := utils.DeleteManifestWork(name, namespace, client, removeFinalizers)
	if err != nil && errors.IsNotFound(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return false, nil
}

// isPaused returns true if the KlusterletAddonConfig instance is labeled as paused, and false otherwise
func isPaused(instance *agentv1.KlusterletAddonConfig) bool {
	a := instance.GetAnnotations()
	if a == nil {
		return false
	}

	if a[KlusterletAddonConfigAnnotationPause] != "" &&
		strings.EqualFold(a[KlusterletAddonConfigAnnotationPause], "true") {
		return true
	}

	return false
}
