// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package klusterletaddon

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/open-cluster-management/endpoint-operator/version"
	ocinfrav1 "github.com/openshift/api/config/v1"
)

var (
	manifestPath = filepath.Join("..", "..", "..", "image-manifests", version.Version+".json")
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
	return agentv1.LoadManifest(manifestPath)
}

func teardown() {
}

func Test_checkComponentIsEnabled(t *testing.T) {

	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
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
			PolicyController: agentv1.KlusterletAddonConfigPolicyControllerSpec{
				Enabled: false,
			},
		},
	}

	tests := []struct {
		name               string
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		componentName      string
		Expected           bool
		wantErr            bool
	}{
		{"enabled", testKlusterletAddonConfig, "appmgr", true, false},
		{"disable", testKlusterletAddonConfig, "policyctrl", false, false},
		{"not supported", testKlusterletAddonConfig, "fakecomponent", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := checkComponentIsEnabled(tt.componentName, tt.klusterletaddoncfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkComponentIsEnabled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, actual, tt.Expected)
			}
		})
	}
}

func Test_syncManifestWorkCRs(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{})

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
			"token": []byte("fake-token"),
		},
	}

	infrastructConfig := &ocinfrav1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocinfrav1.InfrastructureStatus{
			APIServerURL: "https://api.haos-new-playground.purple-chesterfield.com:6443",
		},
	}

	testServiceAccountAppmgr := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-managedcluster",
			},
		},
	}
	testServiceAccountWorkmgr := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-workmgr",
			Namespace: "test-managedcluster",
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-managedcluster",
			},
		},
	}
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
			name: "create manifestwork for all components crs",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testServiceAccountAppmgr, testServiceAccountWorkmgr, infrastructConfig, testSecret,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := syncManifestWorkCRs(tt.args.klusterletaddoncfg, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncManifestWorkCRs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_newCRManifestWork(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{})

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
			"token": []byte("fake-token"),
		},
	}

	infrastructConfig := &ocinfrav1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocinfrav1.InfrastructureStatus{
			APIServerURL: "https://test-hub-cluster.com:6443",
		},
	}

	testServiceAccountAppmgr := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-managedcluster",
			},
		},
	}

	type args struct {
		r                  *ReconcileKlusterletAddon
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		name               string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty name",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
				name:               "",
			},
			wantErr: true,
		},
		{
			name: "wrong name",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
				name:               "invalidname",
			},
			wantErr: true,
		},
		{
			name: "create manifestwork for application manager",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testServiceAccountAppmgr, testSecret, infrastructConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
				name:               "appmgr",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := newCRManifestWork(tt.args.name, tt.args.klusterletaddoncfg, tt.args.r.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCRManifestWork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// try to create manifestwork in k8s with generated manifestwork
			if !tt.wantErr {
				if err := tt.args.r.client.Create(context.TODO(), mw); err != nil {
					t.Errorf("newCRManifestWork() created Manifestwork cannot be created: %v", err)
					return
				}
			}
		})
	}
}

func Test_deleteManifestWorkCRs(t *testing.T) {
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
			PolicyController: agentv1.KlusterletAddonConfigPolicyControllerSpec{
				Enabled: true,
			},
			IAMPolicyControllerConfig: agentv1.KlusterletAddonConfigIAMPolicyControllerSpec{
				Enabled: true,
			},
		},
	}
	manifestWorkAppMgr := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-appmgr",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
	}
	manifestWorkPolicyController := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-policyctrl",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
	}
	manifestWorkIAMPolicyController := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-iampolicyctrl",
			Namespace: testKlusterletAddonConfig.Namespace,
			Finalizers: []string{
				"work.finalizer",
			},
		},
	}
	manifestWorkWorkMgr := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-workmgr",
			Namespace: testKlusterletAddonConfig.Namespace,
			Finalizers: []string{
				"work.finalizer",
			},
		},
	}

	type args struct {
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		client             client.Client
		removeFinalizers   bool
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "remove all",
			args: args{
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
					manifestWorkAppMgr,
					manifestWorkPolicyController,
				}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				removeFinalizers:   false,
			},
			wantErr: false,
		},
		{
			name: "remove all with finalizers",
			args: args{
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
					manifestWorkIAMPolicyController,
					manifestWorkWorkMgr,
				}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				removeFinalizers:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := deleteManifestWorkCRs(tt.args.klusterletaddoncfg, tt.args.client, tt.args.removeFinalizers)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCRManifestWork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}

}

func Test_getServiceAccountToken(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})

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
		},
	}

	testSA := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-appmgr",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-secret-name",
			},
		},
	}

	testSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret-name",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
		Data: map[string][]byte{
			"token": []byte("fake-token"),
		},
	}

	testSecret1 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret-name",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
	}
	type args struct {
		client             client.Client
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		componentName      string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "service account not found",
			args: args{
				client:             fake.NewFakeClient([]runtime.Object{}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "secret not found",
			args: args{
				client:             fake.NewFakeClient([]runtime.Object{testSA}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "token is empty",
			args: args{
				client:             fake.NewFakeClient([]runtime.Object{testSA, testSecret1}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "token is not empty",
			args: args{
				client:             fake.NewFakeClient([]runtime.Object{testSA, testSecret}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getServiceAccountToken(tt.args.client, tt.args.klusterletaddoncfg, tt.args.componentName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getServiceAccountToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_getKubeAPIServerAddress(t *testing.T) {
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{})

	infraConfig := &ocinfrav1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: ocinfrav1.InfrastructureSpec{},
		Status: ocinfrav1.InfrastructureStatus{
			APIServerURL: "http://127.0.0.1:6443",
		},
	}

	type args struct {
		client client.Client
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no cluster",
			args: args{
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{}...),
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "no error",
			args: args{
				client: fake.NewFakeClientWithScheme(testscheme, infraConfig),
			},
			want:    "http://127.0.0.1:6443",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getKubeAPIServerAddress(tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKubeAPIServerAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getKubeAPIServerAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newHubKubeconfigSecret(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{})

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
		},
	}

	testSA := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKlusterletAddonConfig.Name + "-appmgr",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-secret-name",
			},
		},
	}

	testSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret-name",
			Namespace: testKlusterletAddonConfig.Namespace,
		},
		Data: map[string][]byte{
			"token": []byte("fake-token"),
		},
	}

	testinfraConfig := &ocinfrav1.Infrastructure{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Spec: ocinfrav1.InfrastructureSpec{},
		Status: ocinfrav1.InfrastructureStatus{
			APIServerURL: "http://127.0.0.1:6443",
		},
	}

	ExpectedSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "appmgr-hub-kubeconfig",
			Namespace: "test-namespace",
		},
	}

	type args struct {
		client             client.Client
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		componentName      string
		namespace          string
	}
	tests := []struct {
		name    string
		args    args
		want    *corev1.Secret
		wantErr bool
	}{
		{
			name: "service account token is empty",
			args: args{
				client:             fake.NewFakeClientWithScheme(testscheme, []runtime.Object{testSA}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
				namespace:          "test-namespace",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "kubeAPIServer not found",
			args: args{
				client:             fake.NewFakeClientWithScheme(testscheme, []runtime.Object{testSA, testSecret}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
				namespace:          "test-namespace",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{testSA, testSecret,
					testinfraConfig}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				componentName:      "appmgr",
				namespace:          "test-namespace",
			},
			want:    ExpectedSecret,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newHubKubeconfigSecret(tt.args.klusterletaddoncfg, tt.args.client, tt.args.componentName,
				tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("newHubKubeconfigSecret() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				assert.Equal(t, got.Name, ExpectedSecret.Name)
				assert.Equal(t, got.Namespace, ExpectedSecret.Namespace)
			}
		})
	}
}
