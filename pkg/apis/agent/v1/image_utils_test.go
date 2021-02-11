package v1

import (
	"testing"

	"github.com/open-cluster-management/klusterlet-addon-controller/version"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetImageWithManifest(t *testing.T) {
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
				"ocm-release-version": version.Version,
			},
		},
		Data: map[string]string{
			"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":      "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
		},
	}

	client := fake.NewFakeClient([]runtime.Object{
		testConfigMap,
	}...)
	err := LoadConfigmaps(client)
	if err != nil {
		return
	}
	type args struct {
		klusterletaddonconfig *KlusterletAddonConfig
		component             string
	}

	tests := []struct {
		name    string
		args    args
		want    GlobalValues
		wantErr bool
	}{
		{
			name: "Use Component Sha in " + version.Version,
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{
					Spec: KlusterletAddonConfigSpec{
						ImageRegistry: "sample-registry/uniquePath",
					},
				},
				component: "endpoint_component_operator",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{},
				component:             "notExistsComponent",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
		{
			name: "Image not in manifest.json",
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{},
				component:             "fakeKey",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imgRepository, err := tt.args.klusterletaddonconfig.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly. Error:%s", err)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.ImageOverrides[tt.args.component], imgRepository, "repository should match")
			}
		})
	}
}

func TestGetImageWithManyConfigmapManifest(t *testing.T) {
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
				"ocm-release-version": version.Version,
			},
		},
		Data: map[string]string{
			"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":      "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
		},
	}

	testConfigMap1 := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2.2.1",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": "2.2.1",
			},
		},
		Data: map[string]string{
			"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":      "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
		},
	}

	testConfigMapInvalidVersion := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2.2.1.12",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": "2.2.1.12",
			},
		},
		Data: map[string]string{
			"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":      "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
		},
	}

	client := fake.NewFakeClient([]runtime.Object{
		testConfigMap, testConfigMap1, testConfigMapInvalidVersion,
	}...)
	err := LoadConfigmaps(client)
	if err != nil {
		return
	}
	type args struct {
		klusterletaddonconfig *KlusterletAddonConfig
		component             string
	}

	tests := []struct {
		name    string
		args    args
		want    GlobalValues
		wantErr bool
	}{
		{
			name: "Use Component Sha in " + version.Version,
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{
					Spec: KlusterletAddonConfigSpec{
						ImageRegistry: "sample-registry/uniquePath",
					},
				},
				component: "endpoint_component_operator",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"endpoint_component_operator": "sample-registry/uniquePath/endpoint-component-operator@sha256:fake-sha256-2-1-0",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{},
				component:             "notExistsComponent",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
		{
			name: "Image not in manifest.json",
			args: args{
				klusterletaddonconfig: &KlusterletAddonConfig{},
				component:             "fakeKey",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imgRepository, err := tt.args.klusterletaddonconfig.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly. Error:%s", err)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.ImageOverrides[tt.args.component], imgRepository, "repository should match")
			}
		})
	}
}
