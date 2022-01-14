// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package klusterletaddon

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
)

func Test_createManifestWorkCRD(t *testing.T) {
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
		KlusterletAddonConfig: testKlusterletAddonConfig,
		ClusterName:           "test-managedcluster",
	}

	type args struct {
		r                  *ReconcileKlusterletAddon
		klusterletaddoncfg *agentv1.AddonAgentConfig
		kubeversion        string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create manifestwork for crds with kubeversion 0.17.0",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testAddonAgentConfig,
				kubeversion:        "1.17.0",
			},
			wantErr: false,
		},
		{
			name: "create manifestwork for crds with kubeversion 0.11.0",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testAddonAgentConfig,
				kubeversion:        "1.11.0",
			},
			wantErr: false,
		},
		{
			name: "create manifestwork for crds with kubeversion 0.15.0",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testAddonAgentConfig,
				kubeversion:        "1.15.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := createManifestWorkCRD(tt.args.klusterletaddoncfg, tt.args.kubeversion, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncManifestWorkCRs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
