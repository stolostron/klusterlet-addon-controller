package upgrade

import (
	"context"
	"fmt"
	"os"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	annotationReleaseName      = "meta.helm.sh/release-name"
	annotationReleaseNamespace = "meta.helm.sh/release-namespace"
	labelManagedBy             = "app.kubernetes.io/managed-by"
)

func UpgradeMgmtAddonAdd(mgr manager.Manager, dynamicClient dynamic.Interface) error {
	return upgradeMgmtAddonAdd(mgr, newUpgradeMgmtAddonReconciler(mgr, dynamicClient))
}

// newUpgradeMgmtAddonReconciler returns a new reconcile.Reconciler
func newUpgradeMgmtAddonReconciler(mgr manager.Manager, dynamicClient dynamic.Interface) reconcile.Reconciler {
	return &ReconcileUpgradeMgmtAddon{client: mgr.GetClient(), scheme: mgr.GetScheme(), dynamicClient: dynamicClient}
}

func upgradeMgmtAddonAdd(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("upgrade-mgmtAddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ClusterManagementAddOn{}},
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
		)},
		upgradePredicate())

	return err
}

type ReconcileUpgradeMgmtAddon struct {
	client        client.Client
	scheme        *runtime.Scheme
	dynamicClient dynamic.Interface
}

func (r *ReconcileUpgradeMgmtAddon) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	clusterManagementAddon := &addonv1alpha1.ClusterManagementAddOn{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name}, clusterManagementAddon); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	// policy-controller addon is refactored to config-policy-controller and governance-policy-framework addons.
	// need delete the old clusterManagementAddon cr.
	if clusterManagementAddon.Name == agentv1.PolicyAddonName {
		return reconcile.Result{}, r.client.Delete(context.TODO(), clusterManagementAddon, &client.DeleteOptions{})
	}

	needUpdate := false
	newAddon := clusterManagementAddon.DeepCopy()

	// only update release annotations and helm label for search and grc addon because they are installed by
	// subscription. helmRelease cannot update the resources that are not installed by itself.
	if newAddon.Name == agentv1.SearchAddonName ||
		newAddon.Name == agentv1.CertPolicyAddonName ||
		newAddon.Name == agentv1.IamPolicyAddonName {
		releaseName, err := r.GetReleaseName(newAddon.Name)
		if err != nil {
			return reconcile.Result{}, err
		}
		annotations := newAddon.Annotations
		if len(annotations) == 0 {
			newAddon.SetAnnotations(map[string]string{
				annotationReleaseName:      releaseName,
				annotationReleaseNamespace: os.Getenv("POD_NAMESPACE"),
			})
			needUpdate = true
		} else if annotations[annotationReleaseName] == "" || annotations[annotationReleaseNamespace] == "" {
			annotations[annotationReleaseName] = releaseName
			annotations[annotationReleaseNamespace] = os.Getenv("POD_NAMESPACE")
			newAddon.SetAnnotations(annotations)
			needUpdate = true
		}

		labels := newAddon.Labels
		if len(labels) == 0 {
			newAddon.SetLabels(map[string]string{labelManagedBy: "Helm"})
			needUpdate = true
		} else if labels[labelManagedBy] == "" {
			labels[labelManagedBy] = "Helm"
			newAddon.SetLabels(labels)
			needUpdate = true
		}
	}

	if newAddon.Spec.AddOnConfiguration.CRDName == "klusterletaddonconfigs.agent.open-cluster-management.io" {
		newAddon.Spec.AddOnConfiguration.CRDName = ""
		needUpdate = true
	}

	if !needUpdate {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, r.client.Update(context.TODO(), newAddon, &client.UpdateOptions{})
}

var subscriptionsGVR = schema.GroupVersionResource{
	Group:    "apps.open-cluster-management.io",
	Version:  "v1",
	Resource: "subscriptions",
}

func (r *ReconcileUpgradeMgmtAddon) GetReleaseName(addonName string) (string, error) {
	subscriptionName := ""
	releaseNamePrefix := ""
	switch addonName {
	case agentv1.SearchAddonName:
		subscriptionName = "search-prod-sub"
		releaseNamePrefix = "search-prod"
	case agentv1.CertPolicyAddonName, agentv1.IamPolicyAddonName:
		subscriptionName = "grc-sub"
		releaseNamePrefix = "grc"
	default:
		return "", fmt.Errorf("the addon %v is not needed to handle", addonName)
	}

	sub, err := r.dynamicClient.Resource(subscriptionsGVR).Namespace(os.Getenv("POD_NAMESPACE")).
		Get(context.TODO(), subscriptionName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", releaseNamePrefix, getShortSubUID(sub.GetUID())), nil
}

func getShortSubUID(subUID types.UID) string {
	shortUID := subUID

	if len(subUID) >= 5 {
		shortUID = subUID[:5]
	}

	return string(shortUID)
}
