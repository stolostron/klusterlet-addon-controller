// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of connmgr Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
)

var (
	namespace = "multicluster-endpoint"
)

func newTestDeployment(name string) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{appsv1.DeploymentCondition{
				Type:   "Available",
				Status: "True",
			}},
		},
	}
	return deployment
}

func newInstance() *multicloudv1beta1.Endpoint {
	instance := &multicloudv1beta1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint",
			Namespace: namespace,
		},
		Spec: multicloudv1beta1.EndpointSpec{
			Version: "3.2.1",
		},
	}
	return instance
}

func newInstanceInDeletion() *multicloudv1beta1.Endpoint {
	instance := &multicloudv1beta1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "endpoint",
			Namespace:         namespace,
			DeletionTimestamp: &metav1.Time{},
			Finalizers:        []string{"endpoint-connmgr"},
		},
		Spec: multicloudv1beta1.EndpointSpec{
			Version:         "3.2.1",
			ImagePullPolicy: "Always",
			ImagePullSecret: "image-pull",
		},
	}
	return instance
}

func newConnmgr() *multicloudv1beta1.ConnectionManager {
	connmgr := &multicloudv1beta1.ConnectionManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint-connmgr",
			Namespace: namespace,
		},
	}
	return connmgr
}

func TestCreateReconcile(t *testing.T) {
	instance := newInstance()
	connmgr := &multicloudv1beta1.ConnectionManager{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, connmgr)

	objs := []runtime.Object{instance}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "CREATE connmgr reconcile should success")
	assert.False(t, res, "CREATE connmgr reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-connmgr", Namespace: namespace}, connmgr)
	assert.NoError(t, err, "connmgr CR should be created")
}

func TestFinalizeReconcile(t *testing.T) {
	instance := newInstanceInDeletion()
	connmgr := &multicloudv1beta1.ConnectionManager{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, connmgr)

	objs := []runtime.Object{instance}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "FINALIZE connmgr reconcile should success")
	assert.False(t, res, "FINALIZE connmgr reconcile should return false")

	assert.Empty(t, instance.Finalizers, "Finalizer shoule be removed")
}

func TestUpdateReconcile(t *testing.T) {
	instance := newInstance()
	connmgr := newConnmgr()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, connmgr)

	objs := []runtime.Object{connmgr}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "UPDATE connmgr reconcile should success")
	assert.False(t, res, "UPDATE connmgr reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-connmgr", Namespace: namespace}, connmgr)
	assert.NoError(t, err, "connmgr CR should be created")
	assert.Equal(t, connmgr.Spec.ClusterName, instance.Spec.ClusterName, "connmgr CR ClusterName should be updated")
	assert.Equal(t, connmgr.Spec.ClusterNamespace, instance.Spec.ClusterNamespace, "connmgr CR ClusterNamespace should be updated")
	assert.Equal(t, connmgr.Spec.BootStrapConfig, instance.Spec.BootStrapConfig, "connmgr CR BootStrapConfig should be updated")
	assert.Equal(t, connmgr.Spec.ImagePullSecret, instance.Spec.ImagePullSecret, "connmgr CR ImagePullSecret should be updated")
}

func TestDeleteReconcile(t *testing.T) {
	instance := newInstanceInDeletion()
	connmgr := newConnmgr()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, connmgr)

	objs := []runtime.Object{connmgr}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "DELETE connmgr reconcile should success")
	assert.Equal(t, res, true, "DELETE connmgr reconcile should return true")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-connmgr", Namespace: namespace}, connmgr)
	assert.Error(t, err, "connmgr CR should be deleted")
}

func TestIsReady(t *testing.T) {
	instance := newInstanceInDeletion()
	deployment := newTestDeployment(instance.ObjectMeta.Name + "-connmgr")

	objs := []runtime.Object{deployment}
	cl := fake.NewFakeClient(objs...)

	res, err := IsReady(instance, cl)
	assert.NoError(t, err, "IsReady should success")
	assert.Equal(t, res, true, "IsReady should return true")
}

func TestIsNotReady(t *testing.T) {
	instance := newInstanceInDeletion()

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := IsReady(instance, cl)
	assert.NoError(t, err, "IsReady should success")
	assert.False(t, res, "IsReady should return false")
}
