package addon

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	v1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"open-cluster-management.io/api/addon/v1alpha1"
	mcv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func validateValues(values, expectedValues string) error {
	if values == expectedValues {
		return nil
	}

	v := map[string]interface{}{}
	err := json.Unmarshal([]byte(values), &v)
	if err != nil {
		return fmt.Errorf("failed to unmarshal values %v", err)
	}
	ev := map[string]interface{}{}
	err = json.Unmarshal([]byte(expectedValues), &ev)
	if err != nil {
		return fmt.Errorf("failed to unmarshal expectedValues %v", err)
	}

	if !reflect.DeepEqual(v, ev) {
		return fmt.Errorf("the values and expected values are different")
	}
	return nil
}

func Test_updateAnnotationValues(t *testing.T) {
	cases := []struct {
		name             string
		gv               globalValues
		annotationValues string
		expectedValues   string
		expectedErr      bool
	}{
		{
			name: "annotation no global",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "true"},
				ProxyConfig:    nil,
			}},
			annotationValues: `{"logLevel":1}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"true"}}}`,
		},
		{
			name: "annotation no value",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "true"},
				ProxyConfig:    nil,
			}},
			annotationValues: "",
			expectedErr:      false,
			expectedValues:   `{"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"true"}}}`,
		},
		{
			name: "annotation global image override",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "false"},
				ProxyConfig:    map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"true"}}}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
		},
		{
			name: "annotation global no image override",
			gv: globalValues{Global: global{
				NodeSelector: map[string]string{"infraNode": "false"},
				ProxyConfig:  map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"pullPolicy":"Always","imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"true"}}}`,
			expectedErr:      false,
			expectedValues:   `{"logLevel":1,"global":{"pullPolicy":"Always","imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.4"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
		},
		{
			name: "annotation no change",
			gv: globalValues{Global: global{
				ImageOverrides: map[string]string{"multicloud_manager": "myquay.io/multicloud_manager:2.5"},
				NodeSelector:   map[string]string{"infraNode": "false"},
				ProxyConfig:    map[string]string{"HTTP_PROXY": "1.1.1.1", "HTTPS_PROXY": "2.2.2.2", "NO_PROXY": "3.3.3.3"},
			}},
			annotationValues: `{"logLevel":1,"global":{"imageOverrides":{"multicloud_manager":"myquay.io/multicloud_manager:2.5"},"nodeSelector":{"infraNode":"false"},"proxyConfig":{"HTTPS_PROXY":"2.2.2.2","HTTP_PROXY":"1.1.1.1","NO_PROXY":"3.3.3.3"}}}`,
			expectedErr:      false,
			expectedValues:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			values, err := updateAnnotationValues(c.gv, c.annotationValues)
			if !c.expectedErr && err != nil {
				t.Errorf("expected no error but got %v", err)
			}
			if c.expectedErr && err == nil {
				t.Errorf("expected error but got nil")
			}
			if err := validateValues(values, c.expectedValues); err != nil {
				t.Errorf("expected values %v, but got %v. error:%v", c.expectedValues, values, err)
			}
		})
	}
}

func newKlusterletAddonConfig(clusterName string) *v1.KlusterletAddonConfig {
	return &v1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterName,
		},
		Spec: v1.KlusterletAddonConfigSpec{
			ProxyConfig:                v1.ProxyConfig{},
			SearchCollectorConfig:      v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			PolicyController:           v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			ApplicationManagerConfig:   v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			CertPolicyControllerConfig: v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			IAMPolicyControllerConfig:  v1.KlusterletAddonAgentConfigSpec{Enabled: true},
		},
	}
}

func newKlusterletAddonConfigWithProxy(clusterName string) *v1.KlusterletAddonConfig {
	return &v1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterName,
		},
		Spec: v1.KlusterletAddonConfigSpec{
			ProxyConfig:                v1.ProxyConfig{},
			SearchCollectorConfig:      v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			PolicyController:           v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			ApplicationManagerConfig:   v1.KlusterletAddonAgentConfigSpec{Enabled: true, ProxyPolicy: v1.ProxyPolicyOCPGlobalProxy},
			CertPolicyControllerConfig: v1.KlusterletAddonAgentConfigSpec{Enabled: true},
			IAMPolicyControllerConfig:  v1.KlusterletAddonAgentConfigSpec{Enabled: true},
		},
		Status: v1.KlusterletAddonConfigStatus{
			OCPGlobalProxy: v1.ProxyConfig{
				HTTPProxy:  "1.1.1.1",
				HTTPSProxy: "2.2.2.2",
				NoProxy:    "localhost",
			},
		},
	}
}

func newDeletingManagedCluster(name string) *mcv1.ManagedCluster {
	now := metav1.Now()
	return &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			DeletionTimestamp: &now,
		},
	}
}

func newManagedCluster(name string, annotations map[string]string) *mcv1.ManagedCluster {
	return &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
	}
}

func Test_Reconcile(t *testing.T) {
	testscheme := scheme.Scheme
	_ = mcv1.AddToScheme(testscheme)
	_ = v1alpha1.AddToScheme(testscheme)
	_ = apis.AddToScheme(testscheme)

	tests := []struct {
		name                  string
		clusterName           string
		managedCluster        *mcv1.ManagedCluster
		klusterletAddonConfig *v1.KlusterletAddonConfig
		managedClusterAddons  []runtime.Object
		want                  reconcile.Result
		validateFunc          func(t *testing.T, client client.Client)
	}{
		{
			name:                  "cluster is deleted, delete all addons",
			clusterName:           "cluster1",
			klusterletAddonConfig: newKlusterletAddonConfig("cluster1"),
			managedClusterAddons: []runtime.Object{
				newManagedClusterAddon(v1.ApplicationAddonName, "cluster1", ""),
				newManagedClusterAddon(v1.SearchAddonName, "cluster1", ""),
			},
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "notcluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 0 {
					t.Errorf("expected 0 addons, but got %v", len(addonList.Items))
				}
			},
		},
		{
			name:                  "cluster is deleting, delete all addons",
			clusterName:           "cluster1",
			managedCluster:        newDeletingManagedCluster("cluster1"),
			klusterletAddonConfig: newKlusterletAddonConfig("cluster1"),
			managedClusterAddons: []runtime.Object{
				newManagedClusterAddon(v1.ApplicationAddonName, "cluster1", ""),
				newManagedClusterAddon(v1.SearchAddonName, "cluster1", ""),
			},
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "cluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 0 {
					t.Errorf("expected 0 addons, but got %v", len(addonList.Items))
				}
			},
		},
		{
			name:                  "cluster is created, create all addons",
			clusterName:           "cluster1",
			managedCluster:        newManagedCluster("cluster1", nil),
			klusterletAddonConfig: newKlusterletAddonConfig("cluster1"),
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "cluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 6 {
					t.Errorf("expected 6 addons, but got %v", len(addonList.Items))
				}
			},
		},
		{
			name:        "cluster is created in hosed mode with hosted add-on enabled",
			clusterName: "cluster1",
			managedCluster: newManagedCluster("cluster1", map[string]string{
				common.AnnotationKlusterletDeployMode:         "Hosted",
				common.AnnotationKlusterletHostingClusterName: "local-cluster",
				common.AnnotationEnableHostedModeAddons:       "true",
			}),
			klusterletAddonConfig: newKlusterletAddonConfig("cluster1"),
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "cluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 6 {
					t.Errorf("expected 6 addons, but got %v", len(addonList.Items))
				}

				for _, addon := range addonList.Items {
					if hostedAddOns.Has(addon.Name) {
						if value := addon.Annotations[common.AnnotationAddOnHostingClusterName]; value != "local-cluster" {
							t.Errorf("expected hosting cluster of addon %q is %q, but got %s", addon.Name, "local-cluster", value)
						}

						if addon.Spec.InstallNamespace != "klusterlet-cluster1" {
							t.Errorf("expected install namespace of addon %q is %q, but got %s", addon.Name, "klusterlet-cluster1", addon.Spec.InstallNamespace)
						}
					} else {
						if _, ok := addon.Annotations[common.AnnotationAddOnHostingClusterName]; ok {
							t.Errorf("expected addon %q is installed in default mode, but in hosted mode", addon.Name)
						}

						if addon.Spec.InstallNamespace != agentv1.KlusterletAddonNamespace {
							t.Errorf("expected install namespace of addon %q is %q, but got %s", addon.Name, agentv1.KlusterletAddonNamespace, addon.Spec.InstallNamespace)
						}
					}
				}
			},
		},
		{
			name:           "no klusterletaddonconfig",
			clusterName:    "cluster1",
			managedCluster: newManagedCluster("cluster1", nil),
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "cluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 0 {
					t.Errorf("expected 0 addons, but got %v", len(addonList.Items))
				}
			},
		},
		{
			name:        "local-cluster with annotations",
			clusterName: "local-cluster",
			managedCluster: newManagedCluster("local-cluster", map[string]string{
				annotationNodeSelector: `{"node":"infra"}`,
			}),
			klusterletAddonConfig: newKlusterletAddonConfig("local-cluster"),
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "local-cluster"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 6 {
					t.Errorf("expected 6 addons, but got %v", len(addonList.Items))
				}
				for _, addon := range addonList.Items {
					annotations := addon.GetAnnotations()
					values, ok := annotations[annotationValues]
					if !ok {
						t.Errorf("no values annotation")
					}
					gv := globalValues{}
					if err := json.Unmarshal([]byte(values), &gv); err != nil {
						t.Errorf("failed to Unmarshal gv annotation")
					}
					if len(gv.Global.NodeSelector) == 0 {
						t.Errorf("failed to get nodeSelector in gv")
					}
				}
			},
		},
		{
			name:                  "cluster with proxy",
			clusterName:           "cluster1",
			managedCluster:        newManagedCluster("cluster1", nil),
			klusterletAddonConfig: newKlusterletAddonConfigWithProxy("cluster1"),
			managedClusterAddons: []runtime.Object{
				newManagedClusterAddon(v1.ApplicationAddonName, "cluster1", ""),
			},
			validateFunc: func(t *testing.T, kubeClient client.Client) {
				addonList := &v1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: "cluster1"})
				if err != nil {
					t.Errorf("faild to list addons. %v", err)
				}
				if len(addonList.Items) != 6 {
					t.Errorf("expected 6 addons, but got %v", len(addonList.Items))
				}
				for _, addon := range addonList.Items {
					if addon.GetName() != v1.ApplicationAddonName {
						continue
					}
					annotations := addon.GetAnnotations()
					values, ok := annotations[annotationValues]
					if !ok {
						t.Errorf("no values annotation")
					}
					gv := globalValues{}
					if err := json.Unmarshal([]byte(values), &gv); err != nil {
						t.Errorf("failed to Unmarshal gv annotation")
					}
					if len(gv.Global.ProxyConfig) == 0 {
						t.Errorf("failed to get proxyConfig in gv")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			objs := []runtime.Object{}
			if tt.managedCluster != nil {
				objs = append(objs, tt.managedCluster)
			}
			if tt.klusterletAddonConfig != nil {
				objs = append(objs, tt.klusterletAddonConfig)
			}
			if len(tt.managedClusterAddons) != 0 {
				objs = append(objs, tt.managedClusterAddons...)
			}

			reconciler := &ReconcileKlusterletAddOn{
				client: fake.NewClientBuilder().WithScheme(testscheme).WithRuntimeObjects(objs...).Build(),
			}
			request := reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      tt.clusterName,
					Namespace: tt.clusterName,
				},
			}
			actual, err := reconciler.Reconcile(context.TODO(), request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("expected %v but got %v", tt.want, actual)
			}

			tt.validateFunc(t, reconciler.client)
		})
	}
}
