package upgrade

import (
	"context"
	"os"
	"testing"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newMgmtAddon(addonName, crdName string, annotations, labels map[string]string) *addonv1alpha1.ClusterManagementAddOn {
	return &addonv1alpha1.ClusterManagementAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:        addonName,
			Annotations: annotations,
			Labels:      labels,
		},
		Spec: addonv1alpha1.ClusterManagementAddOnSpec{
			AddOnConfiguration: addonv1alpha1.ConfigCoordinates{
				CRDName: crdName,
			},
		},
	}
}

func newSubscription(subName, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps.open-cluster-management.io/v1",
			"kind":       "Subscription",
			"metadata": map[string]interface{}{
				"name":      subName,
				"namespace": namespace,
				"uid":       "34582483-0cb1-4409-b0c1-b3dbbeac85bf",
			},
		},
	}
}

func TestReconcileUpgradeMgmtAddon(t *testing.T) {
	testscheme := scheme.Scheme
	_ = addonv1alpha1.AddToScheme(testscheme)

	tests := []struct {
		name                     string
		addonName                string
		installNamespace         string
		mgmtAddon                *addonv1alpha1.ClusterManagementAddOn
		subscription             runtime.Object
		expectedErr              bool
		expectedReleaseName      string
		expectedReleaseNamespace string
	}{
		{
			name:             "search-collector addon",
			addonName:        agentv1.SearchAddonName,
			installNamespace: "open-cluster-management",
			mgmtAddon: newMgmtAddon(agentv1.SearchAddonName,
				"klusterletaddonconfigs.agent.open-cluster-management.io", nil, nil),
			subscription:             newSubscription("search-prod-sub", "open-cluster-management"),
			expectedErr:              false,
			expectedReleaseName:      "search-prod-34582",
			expectedReleaseNamespace: "open-cluster-management",
		},
		{
			name:             "cert-policy-controller addon",
			addonName:        agentv1.CertPolicyAddonName,
			installNamespace: "open-cluster-management",
			mgmtAddon: newMgmtAddon(agentv1.CertPolicyAddonName,
				"klusterletaddonconfigs.agent.open-cluster-management.io", nil, nil),
			subscription:             newSubscription("grc-sub", "open-cluster-management"),
			expectedErr:              false,
			expectedReleaseName:      "grc-34582",
			expectedReleaseNamespace: "open-cluster-management",
		},
		{
			name:             "work-manager addon",
			addonName:        agentv1.WorkManagerAddonName,
			installNamespace: "open-cluster-management",
			mgmtAddon: newMgmtAddon(agentv1.WorkManagerAddonName,
				"klusterletaddonconfigs.agent.open-cluster-management.io", nil, nil),
			subscription:             newSubscription("grc-sub", "open-cluster-management"),
			expectedErr:              false,
			expectedReleaseName:      "",
			expectedReleaseNamespace: "",
		},
		{
			name:             "cert-policy-controller addon with releaseName",
			addonName:        agentv1.CertPolicyAddonName,
			installNamespace: "open-cluster-management",
			mgmtAddon: newMgmtAddon(agentv1.CertPolicyAddonName,
				"klusterletaddonconfigs.agent.open-cluster-management.io",
				map[string]string{"meta.helm.sh/release-name": "grc-12345",
					"meta.helm.sh/release-namespace": "open-cluster-management"},
				map[string]string{"app.kubernetes.io/managed-by": "Helm"}),
			subscription:             newSubscription("grc-sub", "open-cluster-management"),
			expectedErr:              false,
			expectedReleaseName:      "grc-12345",
			expectedReleaseNamespace: "open-cluster-management",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.addonName,
					Namespace: tt.addonName,
				},
			}

			os.Setenv("POD_NAMESPACE", tt.installNamespace)
			initObjs := []runtime.Object{tt.mgmtAddon}

			reconciler := &ReconcileUpgradeMgmtAddon{
				client:        fake.NewFakeClientWithScheme(testscheme, initObjs...),
				dynamicClient: dynamicfake.NewSimpleDynamicClient(testscheme, tt.subscription),
			}

			_, err := reconciler.Reconcile(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.expectedErr && err == nil {
				t.Errorf("expected error ,but got nil")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("expected no error ,but got %v", err)
			}

			expecedMgmtAddon := &addonv1alpha1.ClusterManagementAddOn{}
			if err = reconciler.client.Get(context.TODO(),
				types.NamespacedName{Name: request.Name}, expecedMgmtAddon); err != nil {
				if errors.IsNotFound(err) {
					t.Errorf("expected got clusterManagementAddon ,but got nil")
				}

				t.Errorf("expected got clusterManagementAddon ,but got err %v", err)
			}
			if expecedMgmtAddon.Spec.AddOnConfiguration.CRDName == "klusterletaddonconfigs.agent.open-cluster-management.io" {
				t.Errorf("expected no crd name ,but got klusterletaddonconfigs")
			}

			if expecedMgmtAddon.Name == agentv1.SearchAddonName || expecedMgmtAddon.Name == agentv1.CertPolicyAddonName {
				if len(expecedMgmtAddon.Annotations) == 0 {
					t.Errorf("expected annotations ,but got 0")
				}

				releaseName := expecedMgmtAddon.Annotations["meta.helm.sh/release-name"]
				if releaseName != tt.expectedReleaseName {
					t.Errorf("expected release name%v , but got %v", tt.expectedReleaseName, releaseName)
				}
				releaseNamespace := expecedMgmtAddon.Annotations["meta.helm.sh/release-namespace"]
				if releaseNamespace != tt.expectedReleaseNamespace {
					t.Errorf("expected release namespace %v ,but got %v", tt.expectedReleaseNamespace, releaseName)
				}

				if len(expecedMgmtAddon.Labels) == 0 {
					t.Errorf("expected labels, but got 0")
				}
				if expecedMgmtAddon.Labels["app.kubernetes.io/managed-by"] != "Helm" {
					t.Errorf("expected labels managed-by Helm, but got %v", expecedMgmtAddon.Labels["app.kubernetes.io/managed-by"])
				}
			}

		})
	}
}
