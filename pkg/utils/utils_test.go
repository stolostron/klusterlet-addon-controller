// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

//Package utils contains common utility functions that gets call by many differerent packages
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
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
