// Copyright Contributors to the Open Cluster Management project

package globalproxy

import (
	"context"
	"reflect"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/helpers"
)

//nolint:unparam // clusterName parameter is part of test helper signature
func newKlusterletAddonConfig(clusterName string, proxyConfig agentv1.ProxyConfig,
	appProxyPolicy agentv1.ProxyPolicy, conditions []metav1.Condition) *agentv1.KlusterletAddonConfig {
	return &agentv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterName,
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
				Enabled:     true,
				ProxyPolicy: appProxyPolicy,
			},
		},
		Status: agentv1.KlusterletAddonConfigStatus{
			OCPGlobalProxy: proxyConfig,
			Conditions:     conditions,
		},
	}
}

func Test_GlobalProxyReconciler_Reconcile(t *testing.T) {
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})

	var testCases = []struct {
		name                          string
		existingKlusterletAddonConfig *agentv1.KlusterletAddonConfig
		existingConfigSecret          *v1.Secret
		request                       ctrl.Request
		expectedKlusterletAddonConfig *agentv1.KlusterletAddonConfig
		expectedResult                ctrl.Result
		expectedErr                   error
	}{
		{
			name:                          "update klusterletAddonConfig status correctly",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, "", []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{
					HTTPProxy:  "https://username:password@proxy.example.com:123/",
					HTTPSProxy: "https://username:password@proxy.example.com:123/",
					NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
				},
				"", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionTrue,
						Reason:  agentv1.ReasonOCPGlobalProxyDetected,
						Message: "Detected the cluster-wide proxy config in install config.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name:                          "update baremetal klusterletAddonConfig status correctly",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, "", []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigBareMetalYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{
					HTTPProxy:  "https://username:password@proxy.example.com:123/",
					HTTPSProxy: "https://username:password@proxy.example.com:123/",
					NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
				},
				"", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionTrue,
						Reason:  agentv1.ReasonOCPGlobalProxyDetected,
						Message: "Detected the cluster-wide proxy config in install config.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name: "update klusterletAddonConfig proxyPolicy correctly",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{
					HTTPProxy:  "https://username:password@proxy.example.com:123/",
					HTTPSProxy: "https://username:password@proxy.example.com:123/",
					NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
				},
				"", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionTrue,
						Reason:  agentv1.ReasonOCPGlobalProxyDetected,
						Message: "Detected the cluster-wide proxy config in install config.",
					},
				}),
			existingConfigSecret: helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{
					HTTPProxy:  "https://username:password@proxy.example.com:123/",
					HTTPSProxy: "https://username:password@proxy.example.com:123/",
					NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
				},
				agentv1.ProxyPolicyOCPGlobalProxy, []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionTrue,
						Reason:  agentv1.ReasonOCPGlobalProxyDetected,
						Message: "Detected the cluster-wide proxy config in install config.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name:                          "no install config secret",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, "", []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-test", "cluster1", helpers.InstallConfigYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{}, "", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionFalse,
						Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
						Message: "The cluster is not provisioned by ACM.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name:                          "no install-config.yaml in secret ",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, "", []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", []byte{}),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{}, "", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionFalse,
						Reason:  agentv1.ReasonOCPGlobalProxyDetectedFail,
						Message: "miss install-config.yaml in install config secret cluster1-install-config",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name:                          "no proxy config in install-config.yaml ",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, "", []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigNoProxyYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{}, "", []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionFalse,
						Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
						Message: "There is no cluster-wide proxy config in install config.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
		{
			name:                 "no klusterletAddonConfig",
			existingConfigSecret: helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: nil,
			expectedResult:                reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second},
			expectedErr:                   nil,
		},
		{
			name:                          "appProxyPolicy is not empty",
			existingKlusterletAddonConfig: newKlusterletAddonConfig("cluster1", agentv1.ProxyConfig{}, agentv1.ProxyPolicyCustomProxy, []metav1.Condition{}),
			existingConfigSecret:          helpers.NewInstallConfigSecret("cluster1-install-config", "cluster1", helpers.InstallConfigYaml),
			request: ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "cluster1",
					Namespace: "cluster1",
				},
			},
			expectedKlusterletAddonConfig: newKlusterletAddonConfig("cluster1",
				agentv1.ProxyConfig{
					HTTPProxy:  "https://username:password@proxy.example.com:123/",
					HTTPSProxy: "https://username:password@proxy.example.com:123/",
					NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
				},
				agentv1.ProxyPolicyCustomProxy, []metav1.Condition{
					{
						Type:    agentv1.OCPGlobalProxyDetected,
						Status:  metav1.ConditionTrue,
						Reason:  agentv1.ReasonOCPGlobalProxyDetected,
						Message: "Detected the cluster-wide proxy config in install config.",
					},
				}),
			expectedResult: reconcile.Result{},
			expectedErr:    nil,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			r := &GlobalProxyReconciler{
				kubeClient: kubefake.NewSimpleClientset(c.existingConfigSecret),
				scheme:     testscheme,
			}

			if c.existingKlusterletAddonConfig != nil {
				r.runtimeClient = fake.NewClientBuilder().WithScheme(testscheme).
					WithRuntimeObjects(c.existingKlusterletAddonConfig).
					WithStatusSubresource(c.existingKlusterletAddonConfig).Build()
			} else {
				r.runtimeClient = fake.NewClientBuilder().WithScheme(testscheme).Build()
			}

			result, err := r.Reconcile(context.TODO(), c.request)
			if err != nil && c.expectedErr == nil {
				t.Errorf("expected no err but got %v", err)
			}
			if err == nil && c.expectedErr != nil {
				t.Errorf("expected err %v, but got nil", c.expectedErr)
			}

			if !reflect.DeepEqual(result, c.expectedResult) {
				t.Errorf("expected globalProxyResult %v,but got %v", c.expectedResult, result)
			}

			if c.expectedKlusterletAddonConfig != nil {
				addonAgentConfig := &agentv1.KlusterletAddonConfig{}
				if err := r.runtimeClient.Get(context.TODO(),
					types.NamespacedName{Name: c.expectedKlusterletAddonConfig.Name,
						Namespace: c.expectedKlusterletAddonConfig.Namespace},
					addonAgentConfig); err != nil {
					t.Errorf("expected KlusterletAddonConfig %v, but got err %v", c.expectedKlusterletAddonConfig, err)
				}
				if !reflect.DeepEqual(addonAgentConfig.Spec, c.expectedKlusterletAddonConfig.Spec) {
					t.Errorf("expected KlusterletAddonConfig spec %v, but got %v",
						c.expectedKlusterletAddonConfig.Spec, addonAgentConfig.Spec)
				}
				if !reflect.DeepEqual(addonAgentConfig.Status.OCPGlobalProxy, c.expectedKlusterletAddonConfig.Status.OCPGlobalProxy) {
					t.Errorf("expected KlusterletAddonConfig status OCPGlobalProxy %v, but got %v",
						c.expectedKlusterletAddonConfig.Status.OCPGlobalProxy, addonAgentConfig.Status.OCPGlobalProxy)
				}

				if len(addonAgentConfig.Status.Conditions) != len(c.expectedKlusterletAddonConfig.Status.Conditions) {
					t.Errorf("expected the condition %v, but got %v",
						c.expectedKlusterletAddonConfig.Status.Conditions, addonAgentConfig.Status.Conditions)
				}
			}
		})
	}
}

func Test_getGlobalProxyInInstallConfig(t *testing.T) {
	var testCases = []struct {
		name                string
		installConfig       []byte
		expectedProxyConfig agentv1.ProxyConfig
		expectedErr         error
	}{
		{
			name:          "get correct proxyConfig",
			installConfig: helpers.InstallConfigYaml,
			expectedProxyConfig: agentv1.ProxyConfig{
				HTTPProxy:  "https://username:password@proxy.example.com:123/",
				HTTPSProxy: "https://username:password@proxy.example.com:123/",
				NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
			},
			expectedErr: nil,
		},
		{
			name:          "no proxy in install config",
			installConfig: helpers.InstallConfigNoProxyYaml,
			expectedProxyConfig: agentv1.ProxyConfig{
				HTTPProxy:  "",
				HTTPSProxy: "",
				NoProxy:    "",
			},
			expectedErr: nil,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			proxyConfig, err := getGlobalProxyInInstallConfig(c.installConfig)
			if err != nil && c.expectedErr == nil {
				t.Errorf("expected no err but got %v", err)
			}
			if err == nil && c.expectedErr != nil {
				t.Errorf("expected err %v, but got nil", c.expectedErr)
			}
			if !reflect.DeepEqual(proxyConfig, c.expectedProxyConfig) {
				t.Errorf("expected proxyConfig %v, but got %v", c.expectedProxyConfig, proxyConfig)
			}
		})
	}
}
