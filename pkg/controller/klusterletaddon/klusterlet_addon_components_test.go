// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package klusterletaddon

import (
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
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
			Name:      "test-configmap-2.5.0",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": "2.5.0",
			},
		},
		Data: map[string]string{
			"klusterlet_addon_operator":           "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
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
			ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
				Enabled: true,
			},
		},
	}

	testAddonAgentConfig := &agentv1.AddonAgentConfig{
		KlusterletAddonConfig:    testKlusterletAddonConfig,
		ClusterName:              "test-managedcluster",
		ImagePullSecret:          "test-managedcluster",
		ImagePullSecretNamespace: "test-managedcluster",
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
		r                *ReconcileKlusterletAddon
		addonAgentConfig *agentv1.AddonAgentConfig
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
				addonAgentConfig: testAddonAgentConfig,
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
				addonAgentConfig: testAddonAgentConfig,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createManifestWorkComponentOperator(tt.args.addonAgentConfig, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("createManifestWorkComponentOperator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
