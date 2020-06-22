// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package klusterletaddon

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
)

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
