// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of appmgr Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

func newInstanceWithEnable() *multicloudv1beta1.Endpoint {
	instance := &multicloudv1beta1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint",
			Namespace: namespace,
		},
		Spec: multicloudv1beta1.EndpointSpec{
			ClusterName:      "test",
			ClusterNamespace: "test",
			Version:          "3.2.1",
			ApplicationManagerConfig: multicloudv1beta1.EndpointApplicationManagerSpec{
				Enabled: true,
			},
		},
	}
	return instance
}

func newInstanceWithDisable() *multicloudv1beta1.Endpoint {
	instance := &multicloudv1beta1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint",
			Namespace: namespace,
		},
		Spec: multicloudv1beta1.EndpointSpec{
			ClusterName:      "test",
			ClusterNamespace: "test",
			Version:          "3.2.1",
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
			Finalizers:        []string{"endpoint-appmgr"},
		},
		Spec: multicloudv1beta1.EndpointSpec{
			Version:         "3.2.1",
			ImagePullSecret: "image-pull",
			ApplicationManagerConfig: multicloudv1beta1.EndpointApplicationManagerSpec{
				Enabled: true,
			},
		},
	}
	return instance
}

func newAppmgr() *multicloudv1beta1.ApplicationManager {
	appmgr := &multicloudv1beta1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint-appmgr",
			Namespace: namespace,
		},
	}
	return appmgr
}

func TestCreateReconcileWithEnable(t *testing.T) {
	instance := newInstanceWithEnable()
	appmgr := &multicloudv1beta1.ApplicationManager{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "CREATE appmgr reconcile should success")
	assert.False(t, res, "CREATE appmgr reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-appmgr", Namespace: namespace}, appmgr)
	assert.NoError(t, err, "appmgr CR should be created")
}

func TestFinalizeReconcileWithDisable(t *testing.T) {
	instance := newInstanceWithDisable()
	appmgr := &multicloudv1beta1.ApplicationManager{}
	crd := &crdv1beta1.CustomResourceDefinition{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, crd)

	objs := []runtime.Object{instance}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "FINALIZE appmgr reconcile should success")
	assert.False(t, res, "FINALIZE appmgr reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-appmgr", Namespace: namespace}, appmgr)
	assert.Error(t, err, "appmgr CR should NOT be created")

	assert.Empty(t, instance.Finalizers, "Finalizer shoule be removed")
}

func TestFinalizeReconcileWithDeletionTimestamp(t *testing.T) {
	instance := newInstanceInDeletion()
	appmgr := &multicloudv1beta1.ApplicationManager{}
	crd := &crdv1beta1.CustomResourceDefinition{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, crd)

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "FINALIZE appmgr reconcile should success")
	assert.False(t, res, "FINALIZE appmgr reconcile should return false")

	assert.Empty(t, instance.Finalizers, "Finalizer shoule be removed")
}

func TestUpdateReconcile(t *testing.T) {
	instance := newInstanceWithEnable()
	appmgr := newAppmgr()
	crd := &crdv1beta1.CustomResourceDefinition{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, crd)

	objs := []runtime.Object{appmgr}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "UPDATE appmgr reconcile should success")
	assert.False(t, res, "UPDATE appmgr reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-appmgr", Namespace: namespace}, appmgr)
	assert.NoError(t, err, "appmgr CR should be created")

	assert.Equal(t, appmgr.Spec.ClusterName, instance.Spec.ClusterName, "appmgr CR ClusterName should be updated")
	assert.Equal(t, appmgr.Spec.ClusterNamespace, instance.Spec.ClusterNamespace, "appmgr CR ClusterNamespace should be updated")
	assert.Equal(t, appmgr.Spec.FullNameOverride, instance.Name+"-appmgr", "appmgr CR FullNameOverride should be updated")
	assert.Equal(t, appmgr.Spec.ImagePullSecret, instance.Spec.ImagePullSecret, "appmgr CR ImagePullSecret should be updated")
}

func TestDeleteReconcileWithDeletionTimestamp(t *testing.T) {
	instance := newInstanceInDeletion()
	appmgr := newAppmgr()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)

	objs := []runtime.Object{appmgr}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "DELETE appmgr reconcile should success")
	assert.Equal(t, res, true, "DELETE appmgr reconcile should return true")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-appmgr", Namespace: namespace}, appmgr)
	assert.Error(t, err, "appmgr CR should be deleted")
}

func TestDeleteReconcileWithDisable(t *testing.T) {
	instance := newInstanceWithDisable()
	appmgr := newAppmgr()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, appmgr)

	objs := []runtime.Object{appmgr}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "DELETE appmgr reconcile should success")
	assert.Equal(t, res, true, "DELETE appmgr reconcile should return true")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-appmgr", Namespace: namespace}, appmgr)
	assert.Error(t, err, "appmgr CR should be deleted")
}

func TestIsReady(t *testing.T) {
	instance := newInstanceInDeletion()
	deploymentAppmgr := newTestDeployment(instance.ObjectMeta.Name + "-appmgr")

	objs := []runtime.Object{deploymentAppmgr}
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
