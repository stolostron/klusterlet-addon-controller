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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcv1 "open-cluster-management.io/api/cluster/v1"

	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/common"
)

const (
	provisionerAnnotation = "cluster.open-cluster-management.io/provisioner"
	// disableAddonAutomaticInstallationAnnotationKey is the annotation key for disabling the functionality of
	// installing addon automatically
	disableAddonAutomaticInstallationAnnotationKey = "addon.open-cluster-management.io/disable-automatic-installation"
)

var log = logf.Log.WithName("managedcluster-controller")

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	if !hostedAddOnEnabled(managedCluster) && !hypershiftCluster(managedCluster) && !clusterClaimCluster(managedCluster) && !hasAnnotationCreateWithDefaultKAC(managedCluster) {
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
		var kacNew *kacv1.KlusterletAddonConfig
		switch {
		case hostedAddOnEnabled(cluster):
			kacNew = hostedKAC(name)
		case clusterClaimCluster(cluster.GetObjectMeta()):
			kacNew = clusterClaimKAC(name)
		case hypershiftCluster(cluster.GetObjectMeta()):
			kacNew = hypershiftKAC(name)
		case hasAnnotationCreateWithDefaultKAC(cluster.GetObjectMeta()):
			kacNew = defaultKAC(name)
		default:
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

func hostedAddOnEnabled(meta metav1.Object) bool {
	switch {
	case meta == nil:
		return false
	case meta.GetAnnotations()[common.AnnotationKlusterletDeployMode] != "Hosted":
		return false
	case len(meta.GetAnnotations()[common.AnnotationKlusterletHostingClusterName]) == 0:
		return false
	case !strings.EqualFold(meta.GetAnnotations()[common.AnnotationEnableHostedModeAddons], "true"):
		return false
	default:
		return true
	}
}

func hypershiftCluster(meta metav1.Object) bool {
	return strings.Contains(meta.GetAnnotations()[provisionerAnnotation], "HypershiftDeployment.cluster.open-cluster-management.io")
}

func clusterClaimCluster(meta metav1.Object) bool {
	return strings.Contains(meta.GetAnnotations()[provisionerAnnotation], "ClusterClaim.hive.openshift.io")
}

func hasAnnotationCreateWithDefaultKAC(meta metav1.Object) bool {
	return strings.EqualFold(meta.GetAnnotations()[common.AnnotationCreateWithDefaultKlusterletAddonConfig], "true")
}

func hostedKAC(clusterName string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterName,
			Name:      clusterName,
		},
		Spec: kacv1.KlusterletAddonConfigSpec{
			ClusterName:                clusterName,
			ClusterNamespace:           clusterName,
			ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
			CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
			PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
		},
	}
}

func clusterClaimKAC(clusterName string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterName,
			Name:      clusterName,
		},
		Spec: kacv1.KlusterletAddonConfigSpec{
			ClusterName:                clusterName,
			ClusterNamespace:           clusterName,
			ClusterLabels:              map[string]string{"vendor": "OpenShift"}, // Required for object to be created
			ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
		},
	}
}

func hypershiftKAC(clusterName string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterName,
			Name:      clusterName,
		},
		Spec: kacv1.KlusterletAddonConfigSpec{
			ClusterName:                clusterName,
			ClusterNamespace:           clusterName,
			ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
			CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
		},
	}
}

func defaultKAC(clusterName string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: clusterName,
			Name:      clusterName,
		},
		Spec: kacv1.KlusterletAddonConfigSpec{
			ClusterName:                clusterName,
			ClusterNamespace:           clusterName,
			ApplicationManagerConfig:   kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			CertPolicyControllerConfig: kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			PolicyController:           kacv1.KlusterletAddonAgentConfigSpec{Enabled: true},
			SearchCollectorConfig:      kacv1.KlusterletAddonAgentConfigSpec{Enabled: false},
		},
	}
}
