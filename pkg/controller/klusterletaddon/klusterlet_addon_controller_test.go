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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/imageregistry/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	ocinfrav1 "github.com/openshift/api/config/v1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
)

func TestReconcileKlusterletAddon_Reconcile(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(managedclusterv1.SchemeGroupVersion, &managedclusterv1.ManagedCluster{})
	testscheme.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.Infrastructure{}, &ocinfrav1.APIServer{})
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
			ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
				Enabled: true,
			},
			Version: "2.0.0",
		},
	}

	terminatingKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-managedcluster",
			Namespace:         "test-managedcluster",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
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
		Type: corev1.SecretTypeServiceAccountToken,
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
	testManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-managedcluster",
		},
		Status: managedclusterv1.ManagedClusterStatus{
			Conditions: []metav1.Condition{
				metav1.Condition{
					Type:    managedclusterv1.ManagedClusterConditionAvailable,
					Status:  metav1.ConditionTrue,
					Reason:  "ManagedClusterAvailable",
					Message: "Managed cluster is available",
				},
			},
		},
	}
	terminatingManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-managedcluster",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
	}

	testManifestWorkCRD := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster" + KlusterletAddonCRDsPostfix,
			Namespace: "test-managedcluster",
		},
		Status: manifestworkv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				metav1.Condition{
					Type:    "Available",
					Status:  metav1.ConditionTrue,
					Reason:  "AppliedManifestWorkComplete",
					Message: "Apply manifest work complete",
				},
			},
		},
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
	}

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}

	type args struct {
		request reconcile.Request
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "klusterletaddonconfig do not exist",
			fields: fields{
				client: fake.NewFakeClient(),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue: false,
			},
			wantErr: false,
		},
		{
			name: "terminating klusterletaddonConfig",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme, testManagedCluster, terminatingKlusterletAddonConfig),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue: false,
			},
			wantErr: false,
		},
		{
			name: "terminating managed cluster",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme, terminatingManagedCluster),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue: false,
			},
			wantErr: false,
		},
		{
			name: "waiting for CRD manifestwork to update status",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme,
					testKlusterletAddonConfig,
					testManagedCluster,
					testSecret,
					infrastructConfig,
					testServiceAccountAppmgr,
					testServiceAccountWorkmgr),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue:      true,
				RequeueAfter: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "success",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme,
					testKlusterletAddonConfig,
					testManagedCluster,
					testSecret,
					infrastructConfig,
					testManifestWorkCRD,
					testServiceAccountAppmgr,
					testServiceAccountWorkmgr),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Minute,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconcileKlusterletAddon{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}

			got, err := r.Reconcile(tt.args.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileKlusterletAddon.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileKlusterletAddon.Reconcile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPaused(t *testing.T) {
	tests := []struct {
		name string
		arg  *agentv1.KlusterletAddonConfig
		want bool
	}{
		{
			name: "Unpaused (No pause annotation)",
			arg:  &agentv1.KlusterletAddonConfig{},
			want: false,
		},
		{
			name: "Paused",
			arg: &agentv1.KlusterletAddonConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: agentv1.SchemeGroupVersion.String(),
					Kind:       "KlusterletAddonConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{KlusterletAddonConfigAnnotationPause: "true"},
				},
			},
			want: true,
		},
		{
			name: "Unpaused (False annotation)",
			arg: &agentv1.KlusterletAddonConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: agentv1.SchemeGroupVersion.String(),
					Kind:       "KlusterletAddonConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{KlusterletAddonConfigAnnotationPause: "false"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPaused(tt.arg); got != tt.want {
				t.Errorf("isPaused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newCustomClient(t *testing.T) {
	secretA := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"data": []byte("fake-data-a"),
		},
		Type: corev1.SecretTypeOpaque,
	}
	secretB := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"data": []byte("fake-data-b"),
		},
		Type: corev1.SecretTypeOpaque,
	}
	configmapA := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"data": "fake-cm-data-a",
		},
	}
	configmapB := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-configmap",
			Namespace: "test-namespace",
		},
		Data: map[string]string{
			"data": "fake-cm-data-b",
		},
	}
	fakeClientA := fake.NewFakeClient(secretA, configmapA)
	fakeClientB := fake.NewFakeClient(secretB, configmapB)
	testClient := newCustomClient(fakeClientA, fakeClientB)

	t.Run("get secret should use apireader", func(t *testing.T) {
		gotSecret := &corev1.Secret{}
		if err := testClient.Get(context.TODO(), types.NamespacedName{
			Name:      "test-secret",
			Namespace: "test-namespace",
		}, gotSecret); err != nil {
			t.Errorf("custom client Get() got %v but wanted nil", err)
		} else if !reflect.DeepEqual(gotSecret.Data["data"], []byte("fake-data-b")) {
			t.Errorf("custom client Get() got %v but wanted %v", gotSecret.Data["data"], []byte("fake-data-b"))
		}
	})
	t.Run("get configmap should use default client", func(t *testing.T) {
		gotConfigmap := &corev1.ConfigMap{}
		if err := testClient.Get(context.TODO(), types.NamespacedName{
			Name:      "test-configmap",
			Namespace: "test-namespace",
		}, gotConfigmap); err != nil {
			t.Errorf("custom client Get() got %v but wanted nil", err)
		} else if !reflect.DeepEqual(gotConfigmap.Data["data"], "fake-cm-data-a") {
			t.Errorf("custom client Get() got %v but wanted %v", gotConfigmap.Data["data"], []byte("fake-cm-data-a"))
		}
	})
	t.Run("can still delete (with default client)", func(t *testing.T) {
		gotSecret := &corev1.Secret{}
		if err := testClient.Delete(context.TODO(), secretA); err != nil {
			t.Errorf("custom client Delete() got %v but wanted nil", err)
		}
		if err := fakeClientA.Get(context.TODO(), types.NamespacedName{
			Name:      "test-secret",
			Namespace: "test-namespace",
		}, gotSecret); !errors.IsNotFound(err) {
			t.Errorf("default client Get() got %v but wanted not found", err)
		}
	})

}

func newImageRegistry(namespace, name, registry, pullSecret string) *v1alpha1.ManagedClusterImageRegistry {
	return &v1alpha1.ManagedClusterImageRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ImageRegistrySpec{
			Registry:   registry,
			PullSecret: corev1.LocalObjectReference{Name: pullSecret},
		},
	}
}

func Test_getImageRegistryAndPullSecret(t *testing.T) {
	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.GroupVersion, &v1alpha1.ManagedClusterImageRegistry{})
	tests := []struct {
		name                    string
		imageRegistryLabelValue string
		client                  func() client.Client
		expectedErr             error
		expectedRegistry        string
		expectedNamespace       string
		expectedPullSecret      string
	}{
		{
			name:                    "get registry and pullSecret successfully",
			imageRegistryLabelValue: "myNamespace.myImageRegistry",
			client: func() client.Client {
				return fake.NewFakeClientWithScheme(s, newImageRegistry("myNamespace", "myImageRegistry", "myRegistry", "mySecret"))
			},
			expectedErr:        nil,
			expectedRegistry:   "myRegistry",
			expectedNamespace:  "myNamespace",
			expectedPullSecret: "mySecret",
		},
		{
			name:                    "invalid imageRegistryLabelValue",
			imageRegistryLabelValue: "myImageRegistry",
			client: func() client.Client {
				return fake.NewFakeClientWithScheme(s, newImageRegistry("myNamespace", "myImageRegistry", "myRegistry", "mySecret"))
			},
			expectedErr:        fmt.Errorf("invalid format of image registry label value myImageRegistry"),
			expectedRegistry:   "",
			expectedNamespace:  "",
			expectedPullSecret: "",
		},
		{
			name:                    "imageRegistry not found",
			imageRegistryLabelValue: "myNamespace.myImageRegistry",
			client: func() client.Client {
				return fake.NewFakeClientWithScheme(s, newImageRegistry("ns1", "ir1", "myRegistry", "mySecret"))
			},
			expectedErr:        fmt.Errorf("managedclusterimageregistries.imageregistry.open-cluster-management.io \"myImageRegistry\" not found"),
			expectedRegistry:   "",
			expectedNamespace:  "",
			expectedPullSecret: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testClient := test.client()
			registry, namespace, pullSecret, err := getImageRegistryAndPullSecret(testClient, test.imageRegistryLabelValue)
			if err == nil && test.expectedErr != nil {
				t.Errorf("should get error %v, but get nil", test.expectedErr)
			}

			if err != nil && test.expectedErr == nil {
				t.Errorf("should get no error, but get %v", err)
			}

			if err != nil && test.expectedErr != nil {
				if !reflect.DeepEqual(err.Error(), test.expectedErr.Error()) {
					t.Errorf("should get error %#v, but get error %#v", test.expectedErr.Error(), err.Error())
				}
			}

			if !reflect.DeepEqual(registry, test.expectedRegistry) {
				t.Errorf("should get registry %v, but get %v", test.expectedRegistry, registry)
			}
			if !reflect.DeepEqual(namespace, test.expectedNamespace) {
				t.Errorf("should get namesapce %v, but get %v", test.expectedNamespace, namespace)
			}
			if !reflect.DeepEqual(pullSecret, test.expectedPullSecret) {
				t.Errorf("should get pullsecret %v, but get %v", test.expectedPullSecret, pullSecret)
			}
		})
	}
}
