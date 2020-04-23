// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of component Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// fakecrdclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
	"github.com/open-cluster-management/endpoint-operator/version"
)

var (
	namespace    = "multicluster-endpoint"
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
	return multicloudv1beta1.LoadManifest(manifestPath)
}

func teardown() {
}

func newTestDeployment(name string) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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
			Version:         "3.2.1",
			ImagePullPolicy: "Always",
			ImagePullSecret: "image-pull",
		},
	}
	return instance
}

func TestCreateReconcile(t *testing.T) {
	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)
	instance := newInstance()
	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)

	err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "Create component reconcile should success")

	deploymentComponentOperator := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-component-operator", Namespace: namespace}, deploymentComponentOperator)
	assert.NoError(t, err, "component deployment should be created")

	clusterRole := &rbacv1.ClusterRole{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-component-operator"}, clusterRole)
	assert.NoError(t, err, "clusterRole should be created")

	serviceAccount := &corev1.ServiceAccount{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-component-operator", Namespace: namespace}, serviceAccount)
	assert.NoError(t, err, "serviceAccount should be created")

	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-component-operator"}, clusterRoleBinding)
	assert.NoError(t, err, "clusterRoleBinding should be created")
}

func TestUpdateReconcile(t *testing.T) {
	instance := newInstance()

	deploymentComponentOperator := newTestDeployment(instance.Name + "-component-operator")

	objs := []runtime.Object{deploymentComponentOperator}
	cl := fake.NewFakeClient(objs...)

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)

	err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "Update component reconcile should success")

	foundDeployment := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: deploymentComponentOperator.Name, Namespace: deploymentComponentOperator.Namespace}, foundDeployment)
	assert.NoError(t, err, "GET component deployment should success")

	assert.Equal(t, foundDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy, instance.Spec.ImagePullPolicy, "update component reconcile should update imagepullpolicy")
	assert.Equal(t, foundDeployment.Spec.Template.Spec.ImagePullSecrets[0].Name, instance.Spec.ImagePullSecret, "update component reconcile should update imagepullsecret")
}

// The fakeclientset from apiextensions-apiserver pkg has a problem and it cannot match real clientset
// func TestInstallComponentCRDs(t *testing.T) {
// 	objs := []runtime.Object{}
// 	clientset := fakecrdclientset.NewSimpleClientset(objs...)
// 	err := InstallComponentCRDs(clientset)
// 	if err != nil {
// 		t.Fatalf("InstallComponentCRDs error: (%v)", err)
// 	}
// }
