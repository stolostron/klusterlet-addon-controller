// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	"testing"

	imageregistryv1alpha1 "github.com/stolostron/cluster-lifecycle-api/imageregistry/v1alpha1"
	"github.com/stolostron/klusterlet-addon-controller/version"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetImageWithManifest(t *testing.T) {
	testConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2.6.2",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": version.Version,
			},
		},
		Data: map[string]string{
			"klusterlet_addon_operator": "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":    "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
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
		addonAgentConfig *AddonAgentConfig
		component        string
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
				addonAgentConfig: &AddonAgentConfig{
					ManagedCluster: &clusterv1.ManagedCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cluster1",
							Annotations: map[string]string{
								imageregistryv1alpha1.ClusterImageRegistriesAnnotation: `{"registries":[{"mirror":"quay.io/rhacm2","source":"sample-registry/uniquePath"}]}`,
							},
						},
					},
				},
				component: "klusterlet_addon_operator",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"klusterlet_addon_operator": "quay.io/rhacm2/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				addonAgentConfig: &AddonAgentConfig{},
				component:        "notExistsComponent",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
		{
			name: "Image not in manifest.json",
			args: args{
				addonAgentConfig: &AddonAgentConfig{},
				component:        "fakeKey",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imgRepository, err := tt.args.addonAgentConfig.GetImage(tt.args.component)
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
			Name:      "test-configmap-2.6.2",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": version.Version,
			},
		},
		Data: map[string]string{
			"klusterlet_addon_operator": "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":    "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
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
			"klusterlet_addon_operator": "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":    "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
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
			"klusterlet_addon_operator": "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":    "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
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
		addonAgentConfig *AddonAgentConfig
		component        string
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
				addonAgentConfig: &AddonAgentConfig{
					ManagedCluster: &clusterv1.ManagedCluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "cluster1",
							Annotations: map[string]string{
								imageregistryv1alpha1.ClusterImageRegistriesAnnotation: `{"registries":[{"mirror":"quay.io/rhacm2","source":"sample-registry/uniquePath"}]}`,
							},
						},
					},
				},
				component: "klusterlet_addon_operator",
			},
			want: GlobalValues{
				ImageOverrides: map[string]string{
					"klusterlet_addon_operator": "quay.io/rhacm2/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
				},
			},
			wantErr: false,
		},
		{
			name: "Not Exists Component",
			args: args{
				addonAgentConfig: &AddonAgentConfig{},
				component:        "notExistsComponent",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
		{
			name: "Image not in manifest.json",
			args: args{
				addonAgentConfig: &AddonAgentConfig{},
				component:        "fakeKey",
			},
			want:    GlobalValues{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running tests %s", tt.name)
			imgRepository, err := tt.args.addonAgentConfig.GetImage(tt.args.component)
			if tt.wantErr != (err != nil) {
				t.Errorf("Should return error correctly. Error:%s", err)
			} else if !tt.wantErr {
				assert.Equal(t, tt.want.ImageOverrides[tt.args.component], imgRepository, "repository should match")
			}
		})
	}
}
