// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package csr

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	csrclientv1 "k8s.io/client-go/kubernetes/typed/certificates/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	clusterNameLabel             = "open-cluster-management.io/cluster-name"
	managedClusterAddonNameLabel = "open-cluster-management.io/addon-name"
)

var log = logf.Log.WithName("controller_csr")

// Add creates a new csr Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	kubeClient, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, err
	}

	return &ReconcileCSR{
		client:    mgr.GetClient(),
		scheme:    mgr.GetScheme(),
		csrClient: kubeClient.CertificatesV1().CertificateSigningRequests(),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("csr-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CertificateSigningRequest
	if err := c.Watch(
		&source.Kind{Type: &certificatesv1.CertificateSigningRequest{}},
		&handler.EnqueueRequestForObject{},
		newCSRPredicate(),
	); err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCSR implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCSR{}

// ReconcileCSR reconciles a ClusterManagementAddOn object
type ReconcileCSR struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	csrClient csrclientv1.CertificateSigningRequestInterface
}

// Reconcile reads that state of the ManagedCluster and ManagedClusterAddOn object and approve the csr if it is
// created to request a client certificate for an addon.
func (r *ReconcileCSR) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling csr")

	// fetch csr instance
	csr := &certificatesv1.CertificateSigningRequest{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, csr); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// skip csr which has been processed
	if len(csr.Status.Conditions) > 0 {
		return reconcile.Result{}, nil
	}

	labels := csr.GetLabels()
	clusterName := labels[clusterNameLabel]
	managedClusterAddonName := labels[managedClusterAddonNameLabel]

	// skip csr which is not created by registration agent
	requestorPrefix := fmt.Sprintf("system:open-cluster-management:%s:", clusterName)
	if ok := strings.HasPrefix(csr.Spec.Username, requestorPrefix); !ok {
		return reconcile.Result{}, nil
	}

	// skip invalid addon registration csr
	if !isValidAddonCSR(csr, managedClusterAddonName, clusterName) {
		return reconcile.Result{}, nil
	}

	// check if ManagedClusterAddOn exists
	if err := r.client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      managedClusterAddonName,
			Namespace: clusterName,
		},
		&addonv1alpha1.ManagedClusterAddOn{},
	); err != nil && errors.IsNotFound(err) {
		// skip csr if ManagedClusterAddOn does not exist
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// fetch the ManagedCluster instance
	managedCluster := &managedclusterv1.ManagedCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{
		Name: clusterName,
	}, managedCluster); err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// check if managed cluster is accepted or not
	if !meta.IsStatusConditionTrue(managedCluster.Status.Conditions, managedclusterv1.ManagedClusterConditionHubAccepted) {
		reqLogger.Info("csr is not approved for Managedcluster has not been accepted yet", "csrName", csr.Name, "clusterName", clusterName)
		return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
	}

	// Auto approve the spoke cluster csr
	if err := r.approveCSR(csr); err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("csr is auto approved by csr controller", "csrName", csr.Name, "addonName", managedClusterAddonName, "clusterName", clusterName)
	return reconcile.Result{}, nil
}

// approveCSR approves the given csr
func (r *ReconcileCSR) approveCSR(csr *certificatesv1.CertificateSigningRequest) error {
	csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1.CertificateSigningRequestCondition{
		Type:    certificatesv1.CertificateApproved,
		Reason:  "AutoApprovedByHubCSRController",
		Message: "Auto approving addon certificate signing request.",
		Status:  corev1.ConditionTrue,
	})
	_, err := r.csrClient.UpdateApproval(context.TODO(), csr.Name, csr, metav1.UpdateOptions{})
	return err
}

func isValidAddonCSR(csr *certificatesv1.CertificateSigningRequest, managedClusterAddonName, clusterName string) bool {
	block, _ := pem.Decode(csr.Spec.Request)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return false
	}

	x509cr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return false
	}

	if len(x509cr.Subject.Organization) != 1 {
		return false
	}

	organization := x509cr.Subject.Organization[0]
	if organization != fmt.Sprintf("system:open-cluster-management:cluster:%s:addon:%s", clusterName, managedClusterAddonName) {
		return false
	}

	if !strings.HasPrefix(x509cr.Subject.CommonName, organization) {
		return false
	}

	return true
}
