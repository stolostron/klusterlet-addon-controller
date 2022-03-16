package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonoperator "github.com/stolostron/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/utils"
	"github.com/stolostron/multicloud-operators-foundation/pkg/apis/imageregistry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func UpgradeAdd(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return upgradeAdd(mgr, newUpgradeReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newUpgradeReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileUpgrade{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

func upgradeAdd(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("upgrade-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &managedclusterv1.ManagedCluster{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetName(),
							Namespace: obj.Meta.GetName(),
						},
					},
				}
			},
		)})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ManagedClusterAddOn{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetNamespace(),
							Namespace: obj.Meta.GetNamespace(),
						},
					},
				}
			},
		)},
		utils.KlusterletAddonPredicate())

	return err
}

type ReconcileUpgrade struct {
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileUpgrade) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	clusterName := request.Name
	// Fetch the managedCluster instance
	managedCluster := &managedclusterv1.ManagedCluster{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: clusterName}, managedCluster); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if !managedCluster.DeletionTimestamp.IsZero() {
		return reconcile.Result{}, nil
	}

	isUpgraded, err := r.addonOperatorIsUpgraded(clusterName)
	if err != nil {
		return reconcile.Result{}, err
	}
	if isUpgraded {
		return reconcile.Result{}, nil
	}

	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: clusterName, Name: clusterName}, klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if len(klusterletAddonConfig.GetFinalizers()) != 0 {
		addonConfig := klusterletAddonConfig.DeepCopy()
		addonConfig.Finalizers = []string{}
		if err = r.client.Update(context.TODO(), addonConfig, &client.UpdateOptions{}); err != nil {
			return reconcile.Result{}, err
		}
	}

	addonAgentConfig, err := r.prepareAddonAgentConfig(managedCluster, klusterletAddonConfig)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.upgradeAddonOperatorManifestWork(addonAgentConfig)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileUpgrade) addonOperatorIsUpgraded(clusterName string) (bool, error) {
	addonOperatorWork := &manifestworkv1.ManifestWork{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Namespace: clusterName, Name: manifestWorkName(clusterName, klusterletAddonOperator)},
		addonOperatorWork); err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	labels := addonOperatorWork.GetLabels()
	if _, ok := labels[agentv1.UpgradeLabel]; !ok {
		return false, nil
	}

	return true, nil
}

func (r *ReconcileUpgrade) prepareAddonAgentConfig(
	cluster *managedclusterv1.ManagedCluster,
	klusterletAddonConfig *agentv1.KlusterletAddonConfig) (*agentv1.AddonAgentConfig, error) {
	addonAgentConfig := &agentv1.AddonAgentConfig{
		KlusterletAddonConfig:    klusterletAddonConfig,
		ClusterName:              cluster.GetName(),
		NodeSelector:             map[string]string{},
		Registry:                 os.Getenv("DEFAULT_IMAGE_REGISTRY"),
		ImagePullSecret:          os.Getenv("DEFAULT_IMAGE_PULL_SECRET"),
		ImagePullSecretNamespace: os.Getenv("POD_NAMESPACE"),
		ImagePullPolicy:          corev1.PullIfNotPresent,
	}
	// if the managedCluster is labelled by imageRegistry, then use it.
	if cluster.Labels[clusterImageRegistryLabel] != "" {
		registry, pullSecretNamespace, pullSecret, err := getImageRegistryAndPullSecret(r.client,
			cluster.Labels[clusterImageRegistryLabel])
		if err != nil {
			klog.Error(err, "failed to get custom registry and pull secret. %v", err)
			return nil, err
		} else {
			addonAgentConfig.ImagePullSecret = pullSecret
			addonAgentConfig.ImagePullSecretNamespace = pullSecretNamespace
			addonAgentConfig.Registry = registry
		}
	}

	if cluster.GetName() == "local-cluster" {
		annotations := cluster.GetAnnotations()
		if nodeSelectorString, ok := annotations[AnnotationNodeSelector]; ok {
			if err := json.Unmarshal([]byte(nodeSelectorString), &addonAgentConfig.NodeSelector); err != nil {
				klog.Errorf("failed to unmarshal nodeSelector annotation of cluster %v. %v", cluster.GetName(), err)
				return nil, err
			}
		}
	}
	return addonAgentConfig, nil
}

func (r *ReconcileUpgrade) upgradeAddonOperatorManifestWork(addonAgentConfig *agentv1.AddonAgentConfig) error {
	var manifests []manifestworkv1.Manifest

	// create namespace
	klusterletaddonNamespace := addonoperator.NewNamespace()

	// Create addon Operator clusterRole
	clusterRole := addonoperator.NewClusterRole(addonAgentConfig)

	// create clusterRoleBinding
	clusterRoleBinding := addonoperator.NewClusterRoleBinding(addonAgentConfig)

	// create service account
	serviceAccount := addonoperator.NewServiceAccount(addonAgentConfig, addonoperator.KlusterletAddonNamespace)

	// create imagePullSecret
	imagePullSecret, err := addonoperator.NewImagePullSecret(addonAgentConfig.ImagePullSecretNamespace,
		addonAgentConfig.ImagePullSecret, r.client)
	if err != nil {
		klog.Error(err, "Fail to create imagePullSecret")
		return err
	}

	// create deployment for klusterlet addon operator
	deployment, err := addonoperator.NewDeployment(addonAgentConfig, addonoperator.KlusterletAddonNamespace)
	if err != nil {
		klog.Error(err, "Fail to create desired klusterlet addon operator deployment")
		return err
	}
	// add namespace, clusterrole, clusterrolebinding, serviceaccount
	nsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: klusterletaddonNamespace}}
	crManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRole}}
	crbManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRoleBinding}}
	saManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: serviceAccount}}
	manifests = append(manifests, nsManifest, crManifest, crbManifest, saManifest)
	// add imagePullSecret
	if imagePullSecret != nil {
		ipsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: imagePullSecret}}
		manifests = append(manifests, ipsManifest)
	}
	// add deployment
	dplManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: deployment}}
	manifests = append(manifests, dplManifest)

	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      manifestWorkName(addonAgentConfig.ClusterName, klusterletAddonOperator),
			Namespace: addonAgentConfig.ClusterName,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			DeleteOption: &manifestworkv1.DeleteOption{
				PropagationPolicy: manifestworkv1.DeletePropagationPolicyTypeSelectivelyOrphan,
				SelectivelyOrphan: &manifestworkv1.SelectivelyOrphan{
					OrphaningRules: []manifestworkv1.OrphaningRule{
						{
							Resource: "namespaces",
							Name:     agentv1.KlusterletAddonNamespace,
						},
					},
				},
			},
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}

	if err := utils.CreateOrUpdateManifestWork(manifestWork, r.client, addonAgentConfig.KlusterletAddonConfig, r.scheme); err != nil {
		klog.Error(err, "Failed to create manifest work for component")
		return err
	}

	return nil
}

// getImageRegistryAndPullSecret gets registry and pullSecret from imageRegistryLabelValue.
// imageRegistryLabelValue format is namespace.imageRegistry
func getImageRegistryAndPullSecret(client client.Client,
	imageRegistryLabelValue string) (registry, namespace, pullSecret string, err error) {
	segments := strings.Split(imageRegistryLabelValue, ".")
	if len(segments) != 2 {
		klog.Errorf("invalid format of image registry label value %v", imageRegistryLabelValue)
		return "", "", "", fmt.Errorf("invalid format of image registry label value %v", imageRegistryLabelValue)
	}
	namespace = segments[0]
	imageRegistryName := segments[1]
	imageRegistry := &v1alpha1.ManagedClusterImageRegistry{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: imageRegistryName, Namespace: namespace}, imageRegistry)
	if err != nil {
		klog.Errorf("failed to get imageregistry %v/%v", namespace, imageRegistryName)
		return "", "", "", err
	}
	registry = imageRegistry.Spec.Registry
	pullSecret = imageRegistry.Spec.PullSecret.Name
	return registry, namespace, pullSecret, nil
}
