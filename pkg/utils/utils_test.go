// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//Package utils contains common utility functions that gets call by many differerent packages
package utils

import (
	"testing"
	"time"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestUniqueStringSlice(t *testing.T) {
	logf.SetLogger(zap.Logger(true))

	testCases := []struct {
		Input          []string
		ExpectedOutput []string
	}{
		{[]string{"foo", "bar"}, []string{"foo", "bar"}},
		{[]string{"foo", "bar", "bar"}, []string{"foo", "bar"}},
		{[]string{"foo", "foo", "bar", "bar"}, []string{"foo", "bar"}},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.ExpectedOutput, UniqueStringSlice(testCase.Input))
	}
}

func TestAddFinalizer(t *testing.T) {
	testManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
				"klusterlet.controller",
			},
		},
	}

	testManagedCluster1 := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
				"klusterlet.controller",
			},
		},
	}

	ExpectedtestManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
				"klusterlet.controller",
			},
		},
	}
	tests := []struct {
		name      string
		cluster   *managedclusterv1.ManagedCluster
		finalizer string
		Expected  *managedclusterv1.ManagedCluster
	}{
		{"add", testManagedCluster, "klusterlet.controller", ExpectedtestManagedCluster},
		{"don't add", testManagedCluster1, "klusterlet.controller", ExpectedtestManagedCluster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddFinalizer(tt.cluster, tt.finalizer)
			assert.Equal(t, tt.cluster, tt.Expected)
		})
	}
}

func TestRemoveFinalizer(t *testing.T) {
	testManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
			},
		},
	}
	testManagedCluster1 := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
				"klusterlet.controller",
			},
		},
	}
	ExpectedtestManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
			},
		},
	}
	tests := []struct {
		name      string
		cluster   *managedclusterv1.ManagedCluster
		finalizer string
		Expected  *managedclusterv1.ManagedCluster
	}{
		{"don't remove", testManagedCluster, "klusterlet.controller", ExpectedtestManagedCluster},
		{"remove", testManagedCluster1, "klusterlet.controller", ExpectedtestManagedCluster},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveFinalizer(tt.cluster, tt.finalizer)
			assert.Equal(t, tt.cluster, tt.Expected)
		})
	}
}

func TestHasFinalizer(t *testing.T) {
	testManagedCluster := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
			},
		},
	}

	testManagedCluster1 := &managedclusterv1.ManagedCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: managedclusterv1.SchemeGroupVersion.String(),
			Kind:       "ManagedCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"cluster.open-cluster-management.io/api-resource-cleanup",
				"klusterlet.controller",
			},
		},
	}

	tests := []struct {
		name      string
		cluster   *managedclusterv1.ManagedCluster
		finalizer string
		Expected  bool
	}{
		{"don't have", testManagedCluster, "klusterlet.controller", false},
		{"have", testManagedCluster1, "klusterlet.controller", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasFinalizer(tt.cluster, tt.finalizer)
			assert.Equal(t, got, tt.Expected)
		})
	}
}

func TestDeleteManifestWork(t *testing.T) {
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})

	manifestWork := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "work",
			Namespace: "test-managedcluster",
		},
	}
	manifestWorkFinalizer := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workfinalizer",
			Namespace: "test-managedcluster",
			Finalizers: []string{
				"work.finalizer",
			},
		},
	}
	type args struct {
		name             string
		namespace        string
		client           client.Client
		removeFinalizers bool
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "remove with no finalizers",
			args: args{
				name:      "work",
				namespace: "test-managedcluster",
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
					manifestWork,
				}...),
				removeFinalizers: false,
			},
			wantErr: false,
		},
		{
			name: "remove with finalizers",
			args: args{
				name:      "workfinalizer",
				namespace: "test-managedcluster",
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
					manifestWorkFinalizer,
				}...),
				removeFinalizers: true,
			},
			wantErr: false,
		},
		{
			name: "remove not found resouce, should return true",
			args: args{
				name:      "uniqueObjectName",
				namespace: "test-managedcluster",
				client: fake.NewFakeClientWithScheme(testscheme, []runtime.Object{
					manifestWorkFinalizer,
				}...),
				removeFinalizers: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteManifestWork(tt.args.name, tt.args.namespace, tt.args.client, tt.args.removeFinalizers)
			if tt.wantErr != (err != nil) {
				t.Errorf("DeleteManifestWork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateOrUpdateManifestWork(t *testing.T) {
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})

	manifestWork := &manifestworkv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			APIVersion: manifestworkv1.SchemeGroupVersion.String(),
			Kind:       "ManifestWork",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "work",
			Namespace: "test-managedcluster",
		},
	}

	type args struct {
		manifestwork *manifestworkv1.ManifestWork
		client       client.Client
		owner        metav1.Object
		scheme       *runtime.Scheme
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create manifestwork",
			args: args{
				manifestwork: manifestWork,
				client:       fake.NewFakeClientWithScheme(testscheme, []runtime.Object{}...),
				owner:        manifestWork,
				scheme:       testscheme,
			},
			wantErr: false,
		},
		{
			name: "update manifestwork",
			args: args{
				manifestwork: manifestWork,
				client:       fake.NewFakeClientWithScheme(testscheme, []runtime.Object{manifestWork}...),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateOrUpdateManifestWork(tt.args.manifestwork, tt.args.client, tt.args.owner, tt.args.scheme)
			if tt.wantErr != (err != nil) {
				t.Errorf("CreateOrUpdateManifestWork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_compareManifests(t *testing.T) {
	testSecret1 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			"token": []byte("fake-token1"),
		},
	}
	testSecret2 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			"token": []byte("fake-token2"),
		},
	}
	type args struct {
		r1 *runtime.RawExtension
		r2 *runtime.RawExtension
	}
	obj1 := runtime.RawExtension{Object: testSecret1}
	obj1copy := runtime.RawExtension{Object: testSecret1.DeepCopy()}
	data1, err := obj1.MarshalJSON()
	if err != nil {
		t.Errorf("failed to marshal object %v", err)
	}
	data1copy, err := obj1copy.MarshalJSON()
	if err != nil {
		t.Errorf("failed to marshal object %v", err)
	}
	raw1 := runtime.RawExtension{Raw: data1, Object: nil}
	raw1copy := runtime.RawExtension{Raw: data1copy, Object: nil}
	obj2 := runtime.RawExtension{Object: testSecret2}
	data2, err := obj2.MarshalJSON()
	if err != nil {
		t.Errorf("failed to marshal object %v", err)
	}
	raw2 := runtime.RawExtension{Raw: data2, Object: nil}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both empty",
			args: args{&runtime.RawExtension{}, &runtime.RawExtension{}},
			want: true,
		},
		{
			name: "one empty",
			args: args{&runtime.RawExtension{}, &obj1},
			want: false,
		},
		{
			name: "same object",
			args: args{&obj1, &obj1copy},
			want: true,
		},
		{
			name: "same raw",
			args: args{&raw1, &raw1copy},
			want: true,
		},
		{
			name: "same raw & obj",
			args: args{&obj1, &raw1},
			want: true,
		},
		{
			name: "different objects",
			args: args{&obj1, &obj2},
			want: false,
		},
		{
			name: "different raw & object",
			args: args{&obj1, &raw2},
			want: false,
		},
		{
			name: "different raws",
			args: args{&raw1, &raw2},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareManifests(tt.args.r1, tt.args.r2)
			if got != tt.want {
				t.Errorf("compareManifests() error = %v, wantErr %v", got, tt.want)
			}
		})
	}
}

func Test_compareManifestWorks(t *testing.T) {
	secret1 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			"token": []byte("fake-token1"),
		},
	}
	secret2 := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Data: map[string][]byte{
			"token": []byte("fake-token2"),
		},
	}
	secret1old := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "test-managedcluster",
			Namespace:         "test-managedcluster",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-200 * time.Second)),
		},
		Data: map[string][]byte{
			"token": []byte("fake-token1"),
		},
	}

	type args struct {
		mw1 *manifestworkv1.ManifestWork
		mw2 *manifestworkv1.ManifestWork
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "exact same",
			args: args{
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
				},
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "value same, timestamp diff",
			args: args{
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							},
						},
					},
				},
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1old}},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "value diff",
			args: args{
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							},
						},
					},
				},
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "out of order",
			args: args{
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
				},
				&manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareManifestWorks(tt.args.mw1, tt.args.mw2)
			if got != tt.want {
				t.Errorf("compareManifestWorks() error = %v, wantErr %v", got, tt.want)
			}
		})
	}

}
