// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedcluster

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mcv1 "open-cluster-management.io/api/cluster/v1"

	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

const (
	provisionerAnnotation = "cluster.open-cluster-management.io/provisioner"
	// disableAddonAutomaticInstallationAnnotationKey is the annotation key for disabling the functionality of
	// installing addon automatically
	disableAddonAutomaticInstallationAnnotationKey = "addon.open-cluster-management.io/disable-automatic-installation"
)

var log = logf.Log.WithName("managedcluster-controller")

// Add creates a new ManagedCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("managedcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &mcv1.ManagedCluster{}}, &handler.EnqueueRequestForObject{},
		predicate.Predicate(predicate.Funcs{
			GenericFunc: func(e event.GenericEvent) bool { return false },
			CreateFunc: func(e event.CreateEvent) bool {
				if e.Object == nil {
					log.Error(nil, "Create event has no runtime object to create", "event", e)
					return false
				}

				return hypershiftCluster(e.Object) || clusterClaimCluster(e.Object)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				if e.ObjectOld == nil || e.ObjectNew == nil {
					log.Error(nil, "Update event is invalid", "event", e)
					return false
				}

				return hypershiftCluster(e.ObjectOld) || clusterClaimCluster(e.ObjectNew)
			},
		}))
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &kacv1.KlusterletAddonConfig{}}, &handler.EnqueueRequestForObject{},
		predicate.Predicate(predicate.Funcs{
			GenericFunc: func(e event.GenericEvent) bool { return false },
			CreateFunc:  func(e event.CreateEvent) bool { return false },
			// If the klusterletAddonConfig is deleted, we will recreate it.
			DeleteFunc: func(e event.DeleteEvent) bool { return true },
			UpdateFunc: func(e event.UpdateEvent) bool { return false },
		}))
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManagedCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManagedCluster{}

// ReconcileManagedCluster reconciles a ManagedCluster object
type ReconcileManagedCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads managed cluster created by hive or hypershift, and create the default
// klusterlet addon config for them
func (r *ReconcileManagedCluster) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Name", request.Name)
	reqLogger.Info("Reconciling ManagedCluster")

	// Fetch the managedCluster instance
	managedCluster := &mcv1.ManagedCluster{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: request.Name}, managedCluster); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	if !hypershiftCluster(managedCluster) && !clusterClaimCluster(managedCluster) {
		return reconcile.Result{}, nil
	}

	if value, ok := managedCluster.Annotations[disableAddonAutomaticInstallationAnnotationKey]; ok &&
		strings.EqualFold(value, "true") {

		reqLogger.Info("Cluster has disable addon automatic installation annotation, skip addon deploy")
		return reconcile.Result{}, nil
	}

	// Create the klusterletAddonConfig if it does not exist
	return reconcile.Result{}, createKlusterletAddonConfig(r.client, managedCluster)
}

func createKlusterletAddonConfig(client client.Client, cluster *mcv1.ManagedCluster) error {
	ctx := context.Background()
	name := cluster.Name

	var kac kacv1.KlusterletAddonConfig
	err := client.Get(ctx, types.NamespacedName{Namespace: name, Name: name}, &kac)
	if errors.IsNotFound(err) {
		log.Info(fmt.Sprintf("Create a new KlusterletAddonConfig resource %s", name))
		kacNew := newKlusterletAddonConfig(clusterType(cluster), name)
		if kacNew == nil {
			return fmt.Errorf("new KlusterletAddonConfig %s", name)
		}
		if err = client.Create(ctx, kacNew); err != nil {
			return fmt.Errorf("create KlusterletAddonConfig %s error: %v", name, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("retreive KlusterletAddonConfig %s error: %v", name, err)
	}

	return nil
}

func hypershiftCluster(meta metav1.Object) bool {
	return strings.Contains(meta.GetAnnotations()[provisionerAnnotation], "HypershiftDeployment.cluster.open-cluster-management.io")
}

func clusterClaimCluster(meta metav1.Object) bool {
	return strings.Contains(meta.GetAnnotations()[provisionerAnnotation], "ClusterClaim.hive.openshift.io")
}

func clusterType(cluster *mcv1.ManagedCluster) string {
	if clusterClaimCluster(cluster.GetObjectMeta()) {
		return clusterTypeClusterClaim
	}

	if hypershiftCluster(cluster.GetObjectMeta()) {
		return clusterTypeHypershift
	}

	return "Unknown"
}

func newKlusterletAddonConfig(clusterType string, name string) *kacv1.KlusterletAddonConfig {
	switch clusterType {
	case clusterTypeClusterClaim:
		return &kacv1.KlusterletAddonConfig{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name,
				Name:      name,
			},
			Spec: kacv1.KlusterletAddonConfigSpec{
				ClusterName:                name,
				ClusterNamespace:           name,
				ClusterLabels:              map[string]string{"vendor": "OpenShift"}, // Required for object to be created
				ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				IAMPolicyControllerConfig:  kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			},
		}
	case clusterTypeHypershift:
		return &kacv1.KlusterletAddonConfig{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name,
				Name:      name,
			},
			Spec: kacv1.KlusterletAddonConfigSpec{
				ClusterName:                name,
				ClusterNamespace:           name,
				ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
				CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				IAMPolicyControllerConfig:  kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
				SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
			},
		}
	default:
		return nil
	}
}

const (
	clusterTypeHypershift   string = "Hypershift"
	clusterTypeClusterClaim string = "ClusterClaim"
)
