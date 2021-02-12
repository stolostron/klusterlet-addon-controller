package klusterletaddon

import (
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
)

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		os.Exit(999)
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() error {
	testConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2.2.0",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": "2.2.0",
			},
		},
		Data: map[string]string{
			"endpoint_component_operator":         "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":              "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
			"iam_policy_controller":               "sample-registry/uniquePath/iam-policy-controller@sha256:fake-sha256-2-1-0",
			"config_policy_controller":            "sample-registry/uniquePath/config-policy-controller@sha256:fake-sha256-2-1-0",
			"governance_policy_spec_sync":         "sample-registry/uniquePath/governance-policy-spec-sync@sha256:fake-sha256-2-1-0",
			"governance_policy_status_sync":       "sample-registry/uniquePath/governance-policy-status-sync@sha256:fake-sha256-2-1-0",
			"governance_policy_template_sync":     "sample-registry/uniquePath/governance-policy-template-sync@sha256:fake-sha256-2-1-0",
			"search_collector":                    "sample-registry/uniquePath/search-collector@sha256:fake-sha256-2-1-0",
			"multicloud_manager":                  "sample-registry/uniquePath/multicloud-manager@sha256:fake-sha256-2-1-0",
			"multicluster_operators_subscription": "sample-registry/uniquePath/multicluster-operators-subscription@sha256:fake-sha256-2-1-0",
		},
	}

	client := fake.NewFakeClient([]runtime.Object{
		testConfigMap,
	}...)

	return agentv1.LoadConfigmaps(client)

}

func teardown() {
}

func Test_createManifestWorkComponentOperator(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})

	testKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonConfigApplicationManagerSpec{
				Enabled: true,
			},
			ImagePullSecret: "test-managedcluster",
		},
	}

	testSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte("fake-token"),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	testWrongSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte("fake-token"),
		},
	}

	type args struct {
		r                  *ReconcileKlusterletAddon
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testSecret,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
			},
			wantErr: false,
		},
		{
			name: "wrong secret",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testWrongSecret,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createManifestWorkComponentOperator(tt.args.klusterletaddoncfg, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("createManifestWorkComponentOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
