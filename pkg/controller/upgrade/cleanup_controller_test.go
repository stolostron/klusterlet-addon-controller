package upgrade

import (
	"context"
	"reflect"
	"testing"
	"time"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"open-cluster-management.io/api/addon/v1alpha1"
	workv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
)

func newManagedClusterAddon(name, namespace string, conditions []metav1.Condition) *v1alpha1.ManagedClusterAddOn {
	return &v1alpha1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ManagedClusterAddOnSpec{
			InstallNamespace: "open-cluster-management-agent-addon",
		},
		Status: v1alpha1.ManagedClusterAddOnStatus{
			Conditions: conditions,
		},
	}
}

func newManifestWork(name, namespace string) *workv1.ManifestWork {
	return &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      manifestWorkName(namespace, name),
			Namespace: namespace,
		},
	}
}

func newOperatorManifestWork(namespace string) *workv1.ManifestWork {
	return &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      manifestWorkName(namespace, klusterletAddonOperator),
			Namespace: namespace,
			Labels:    map[string]string{agentv1.UpgradeLabel: ""},
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				{
					Type:   "Available",
					Status: "True",
				},
				{
					Type:   "Applied",
					Status: "True",
				},
			},
			ResourceStatus: workv1.ManifestResourceStatus{
				Manifests: []workv1.ManifestCondition{
					{
						ResourceMeta: workv1.ManifestResourceMeta{
							Resource: "deployments",
						},
						StatusFeedbacks: workv1.StatusFeedbackResult{},
						Conditions: []metav1.Condition{
							{
								Type:   "Applied",
								Status: "True",
							},
							{
								Type:   "Available",
								Status: "True",
							},
							{
								Type:   "StatusFeedbackSynced",
								Status: "True",
							},
						},
					},
				},
			},
		},
	}
}

func newRoleBinding(name, namespace string) *v1.RoleBinding {
	return &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName(namespace, name),
			Namespace: namespace,
		},
	}
}

func TestReconcileManagedCluster(t *testing.T) {
	testscheme := scheme.Scheme
	_ = workv1.AddToScheme(testscheme)
	_ = apis.AddToScheme(testscheme)
	_ = v1alpha1.AddToScheme(testscheme)

	tests := []struct {
		name                       string
		addonName, clusterName     string
		addon                      runtime.Object
		manifestWorks              []runtime.Object
		roleBindings               []runtime.Object
		expectedResult             reconcile.Result
		expectedManifestWorksCount int
		expectedRoleBindingCount   int
		expectedAddonCount         int
	}{
		{
			name:        "cleanup policy addon",
			addonName:   agentv1.PolicyAddonName,
			clusterName: "cluster1",
			addon: newManagedClusterAddon(agentv1.PolicyAddonName, "cluster1",
				[]metav1.Condition{{Type: "Available", Status: "False",
					LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute))}}),
			manifestWorks: []runtime.Object{newManifestWork("klusterlet-addon-policyctrl", "cluster1"),
				newManifestWork(klusterletAddonCRDs, "cluster1"),
				newOperatorManifestWork("cluster1")},
			roleBindings:               []runtime.Object{newRoleBinding("policyctrl", "cluster1")},
			expectedResult:             reconcile.Result{Requeue: false},
			expectedAddonCount:         0,
			expectedManifestWorksCount: 0,
			expectedRoleBindingCount:   0,
		},
		{
			name:        "cleanup work-manager addon",
			addonName:   agentv1.WorkManagerAddonName,
			clusterName: "cluster1",
			addon: newManagedClusterAddon(agentv1.WorkManagerAddonName, "cluster1",
				[]metav1.Condition{
					{Type: "RegistrationApplied", Status: "True",
						LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute))},
					{Type: "Available", Status: "True"}, {Type: "ManifestApplied", Status: "True"},
				}),
			manifestWorks: []runtime.Object{newManifestWork("klusterlet-addon-workmgr", "cluster1"),
				newManifestWork(klusterletAddonCRDs, "cluster1"),
				newOperatorManifestWork("cluster1")},
			roleBindings:               []runtime.Object{newRoleBinding("workmgr", "cluster1")},
			expectedResult:             reconcile.Result{Requeue: false},
			expectedAddonCount:         1,
			expectedManifestWorksCount: 0,
			expectedRoleBindingCount:   0,
		},
		{
			name:        "cleanup certpolicyctrl addon",
			addonName:   agentv1.CertPolicyAddonName,
			clusterName: "cluster1",
			addon: newManagedClusterAddon(agentv1.CertPolicyAddonName, "cluster1",
				[]metav1.Condition{
					{Type: "RegistrationApplied", Status: "True",
						LastTransitionTime: metav1.NewTime(time.Now().Add(-10 * time.Minute))},
					{Type: "Available", Status: "True"}, {Type: "ManifestApplied", Status: "True"},
				}),
			manifestWorks: []runtime.Object{newManifestWork("klusterlet-addon-certpolicyctrl", "cluster1"),
				newManifestWork("klusterlet-addon-iampolicyctrl", "cluster1"),
				newManifestWork(klusterletAddonCRDs, "cluster1"),
				newOperatorManifestWork("cluster1")},
			roleBindings: []runtime.Object{newRoleBinding("certpolicyctrl", "cluster1"),
				newRoleBinding("iampolicyctrl", "cluster1")},
			expectedResult:             reconcile.Result{Requeue: false},
			expectedAddonCount:         1,
			expectedManifestWorksCount: 3,
			expectedRoleBindingCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.addonName,
					Namespace: tt.clusterName,
				},
			}

			initObjs := []runtime.Object{tt.addon}
			initObjs = append(initObjs, tt.manifestWorks...)
			initObjs = append(initObjs, tt.roleBindings...)

			reconciler := &ReconcileCleanup{
				client: fake.NewFakeClientWithScheme(testscheme, initObjs...),
			}

			actual, err := reconciler.Reconcile(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, tt.expectedResult) {
				t.Errorf("expected %v but got %v", tt.expectedResult, actual)
			}

			addonList := &v1alpha1.ManagedClusterAddOnList{}
			err = reconciler.client.List(context.TODO(), addonList, &client.ListOptions{})
			if err != nil {
				t.Errorf("failed to list addon %v", err)
			}
			if len(addonList.Items) != tt.expectedAddonCount {
				t.Errorf("expected addon count %v, but got %v", tt.expectedAddonCount, len(addonList.Items))
			}

			manifestWorkList := &workv1.ManifestWorkList{}
			err = reconciler.client.List(context.TODO(), manifestWorkList, &client.ListOptions{})
			if err != nil {
				t.Errorf("failed to list manfiestworks %v", err)
			}
			if len(manifestWorkList.Items) != tt.expectedManifestWorksCount {
				t.Errorf("expected work count %v, but got %v", tt.expectedManifestWorksCount, len(manifestWorkList.Items))
			}

			rolbeBindingList := &v1.RoleBindingList{}
			err = reconciler.client.List(context.TODO(), rolbeBindingList, &client.ListOptions{})
			if err != nil {
				t.Errorf("failed to list rolbeBinding %v", err)
			}
			if len(rolbeBindingList.Items) != tt.expectedRoleBindingCount {
				t.Errorf("expected rolebinding count %v, but got %v", tt.expectedRoleBindingCount, len(rolbeBindingList.Items))
			}
		})
	}
}
