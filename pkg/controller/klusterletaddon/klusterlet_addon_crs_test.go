// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package klusterletaddon

import (
	"context"
	"reflect"
	"testing"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	appmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/iampolicycontroller/v1"
	ocinfrav1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_syncManifestWorkCRs(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{}, &ocinfrav1.APIServer{})
	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})

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

	testConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap-2.3.0",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"ocm-configmap-type":  "image-manifest",
				"ocm-release-version": "2.3.0",
			},
		},
		Data: map[string]string{
			"klusterlet_addon_operator": "sample-registry/uniquePath/klusterlet-addon-operator@sha256:fake-sha256-2-1-0",
			"cert_policy_controller":    "sample-registry/uniquePath/cert-policy-controller@sha256:fake-sha256-2-1-0",
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
			Version: "2.0.0",
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
						testKlusterletAddonConfig, testServiceAccountAppmgr, testServiceAccountWorkmgr,
						infrastructConfig, testSecret, testConfigMap,
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

func Test_syncManagedClusterAddonCRs(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})

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
			SearchCollectorConfig: agentv1.KlusterletAddonConfigSearchCollectorSpec{
				Enabled: true,
			},
			Version: "2.3.0",
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
			name: "create ManagedClusterAddons for all components crs in klusterletaddonconfig",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig,
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
			err := syncManagedClusterAddonCRs(tt.args.klusterletaddoncfg, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("syncManagedClusterAddonCRs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
func Test_newCRManifestWork(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{}, &ocinfrav1.APIServer{})

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
			Version: "2.0.0",
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
	testServiceAccountCert := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-certpolicyctrl",
			Namespace: "test-managedcluster",
		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "test-managedcluster",
			},
		},
	}
	testServiceAccountIAM := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-iampolicyctrl",
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
		addon              addons.KlusterletAddon
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create manifestwork for cert policy controller",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testServiceAccountCert, testSecret,
						infrastructConfig,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
				addon:              certpolicyctrl.AddonCertPolicyCtrl{},
			},
			wantErr: false,
		},
		{
			name: "create manifestwork for iam policy controller",
			args: args{
				r: &ReconcileKlusterletAddon{
					client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
						testKlusterletAddonConfig, testServiceAccountAppmgr, testSecret, infrastructConfig,
						testServiceAccountIAM,
					}...),
					scheme: testscheme,
				},
				klusterletaddoncfg: testKlusterletAddonConfig,
				addon:              iampolicyctrl.AddonIAMPolicyCtrl{},
			},
			wantErr: false,
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
				addon:              appmgr.AddonAppMgr{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw, err := newCRManifestWork(tt.args.addon, tt.args.klusterletaddoncfg, tt.args.r.client)
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
			Version: "2.0.0",
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

func Test_updateManagedClusterAddon(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})

	testKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-name",
			Namespace: "test-managedcluster-namespace",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonConfigApplicationManagerSpec{
				Enabled: true,
			},
			Version: "2.0.0",
		},
	}

	addon1 := certpolicyctrl.AddonCertPolicyCtrl{}
	addon2 := appmgr.AddonAppMgr{}
	addonResource := addonv1alpha1.ObjectReference{
		Name:     "test-managedcluster-name",
		Group:    "agent.open-cluster-management.io",
		Resource: "klusterletaddonconfigs",
	}
	mca1 := &addonv1alpha1.ManagedClusterAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ManagedClusterAddon",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon1.GetManagedClusterAddOnName(),
			Namespace: "test-managedcluster-namespace",
		},
	}
	mca2 := &addonv1alpha1.ManagedClusterAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ManagedClusterAddon",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon2.GetManagedClusterAddOnName(),
			Namespace: "test-managedcluster-namespace",
		},
		Status: addonv1alpha1.ManagedClusterAddOnStatus{
			RelatedObjects: []addonv1alpha1.ObjectReference{addonResource},
		},
	}
	// if not exist will create one with correct name & ref

	// if exist but not right ref, will add ref
	type args struct {
		client             client.Client
		klusterletaddoncfg *agentv1.KlusterletAddonConfig
		addon              addons.KlusterletAddon
		scheme             *runtime.Scheme
	}
	tests := []struct {
		name              string
		args              args
		wantAddonResource []addonv1alpha1.ObjectReference
		wantErr           bool
	}{
		{
			name: "create when not created",
			args: args{
				client:             fake.NewFakeClientWithScheme(testscheme, []runtime.Object{}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				addon:              addon1,
				scheme:             testscheme,
			},
			wantAddonResource: []addonv1alpha1.ObjectReference{addonResource},
			wantErr:           false,
		},
		{
			name: "update when not complete",
			args: args{
				client:             fake.NewFakeClientWithScheme(testscheme, []runtime.Object{mca1}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				addon:              addon1,
				scheme:             testscheme,
			},
			wantAddonResource: []addonv1alpha1.ObjectReference{addonResource},
			wantErr:           false,
		},
		{
			name: "do nothing when same",
			args: args{
				client:             fake.NewFakeClientWithScheme(testscheme, []runtime.Object{mca2}...),
				klusterletaddoncfg: testKlusterletAddonConfig,
				addon:              addon2,
				scheme:             testscheme,
			},
			wantAddonResource: []addonv1alpha1.ObjectReference{addonResource},
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updateManagedClusterAddon(tt.args.addon, tt.args.klusterletaddoncfg,
				tt.args.client, tt.args.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateManagedClusterAddon() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil && !tt.wantErr {
				//get it should succeed
				getMca := &addonv1alpha1.ManagedClusterAddOn{}
				getErr := tt.args.client.Get(context.TODO(),
					types.NamespacedName{
						Name:      tt.args.addon.GetManagedClusterAddOnName(),
						Namespace: "test-managedcluster-namespace",
					},
					getMca,
				)
				if getErr != nil {
					t.Errorf("failed to get ManagedClusterAddon")
					return
				}
				if !reflect.DeepEqual(tt.wantAddonResource, getMca.Status.RelatedObjects) {
					t.Errorf("wrong addonResource in ManagedClusterAddon, want %v got %v",
						tt.wantAddonResource, getMca.Status.RelatedObjects)
					return
				}
			}
		})
	}

}
