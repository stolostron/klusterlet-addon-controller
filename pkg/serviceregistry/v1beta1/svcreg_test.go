// Package v1beta1 of serviceregistry Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
)

var (
	namespace = "multicluster-endpoint"
)

func newTestDeployment(name string) *extensionsv1beta1.Deployment {
	deployment := &extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: extensionsv1beta1.DeploymentStatus{
			Conditions: []extensionsv1beta1.DeploymentCondition{extensionsv1beta1.DeploymentCondition{
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
			ServiceRegistryConfig: multicloudv1beta1.EndpointServiceRegistrySpec{
				Enabled:                            true,
				DNSSuffix:                          "123",
				Plugins:                            "test",
				IstioIngressGateway:                "test",
				IstioserviceEntryRegistryNamespace: "test",
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
			Finalizers:        []string{"endpoint-svcreg"},
		},
		Spec: multicloudv1beta1.EndpointSpec{
			Version:         "3.2.1",
			ImagePullSecret: "image-pull",
			ServiceRegistryConfig: multicloudv1beta1.EndpointServiceRegistrySpec{
				Enabled: true,
			},
		},
	}
	return instance
}

func newSvcreg() *multicloudv1beta1.ServiceRegistry {
	svcreg := &multicloudv1beta1.ServiceRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint-svcreg",
			Namespace: namespace,
		},
	}
	return svcreg
}

func TestCreateReconcileWithEnable(t *testing.T) {
	instance := newInstanceWithEnable()
	svcreg := &multicloudv1beta1.ServiceRegistry{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "CREATE svcreg reconcile should success")
	assert.False(t, res, "CREATE svcreg reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-svcreg", Namespace: namespace}, svcreg)
	assert.NoError(t, err, "svcreg CR should be created")
}

func TestFinalizeReconcileWithDisable(t *testing.T) {
	instance := newInstanceWithDisable()
	svcreg := &multicloudv1beta1.ServiceRegistry{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "CREATE svcreg reconcile should success")
	assert.False(t, res, "CREATE svcreg reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-svcreg", Namespace: namespace}, svcreg)
	assert.Error(t, err, "svcreg CR should NOT be created")

	assert.Empty(t, instance.Finalizers, "Finalizer shoule be removed")
}

func TestFinalizeReconcileWithDeletionTimestamp(t *testing.T) {
	instance := newInstanceInDeletion()
	svcreg := &multicloudv1beta1.ServiceRegistry{}

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "FINALIZE svcreg reconcile should success")
	assert.False(t, res, "FINALIZE svcreg reconcile should return false")

	assert.Empty(t, instance.Finalizers, "Finalizer shoule be removed")
}

func TestUpdateReconcile(t *testing.T) {
	instance := newInstanceWithEnable()
	svcreg := newSvcreg()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{svcreg}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "UPDATE svcreg reconcile should success")
	assert.False(t, res, "UPDATE svcreg reconcile should return false")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-svcreg", Namespace: namespace}, svcreg)
	assert.NoError(t, err, "svcreg CR should be created")

	assert.Equal(t, svcreg.Spec.ClusterName, instance.Spec.ClusterName, "svcreg CR ClusterName should be updated")
	assert.Equal(t, svcreg.Spec.ClusterNamespace, instance.Spec.ClusterNamespace, "svcreg CR ClusterNamespace should be updated")
	assert.Equal(t, svcreg.Spec.FullNameOverride, instance.Name+"-svcreg", "svcreg CR FullNameOverride should be updated")
	assert.Equal(t, svcreg.Spec.ImagePullSecret, instance.Spec.ImagePullSecret, "svcreg CR ImagePullSecret should be updated")
	assert.Equal(t, svcreg.Spec.ConnectionManager, instance.Name+"-connmgr", "svcreg CR ConnectionManager should be updated")
	assert.Equal(t, svcreg.Spec.DNSSuffix, instance.Spec.ServiceRegistryConfig.DNSSuffix, "svcreg CR DNSSuffix should be updated")
	assert.Equal(t, svcreg.Spec.Plugins, instance.Spec.ServiceRegistryConfig.Plugins, "svcreg CR Plugins should be updated")
	assert.Equal(t, svcreg.Spec.IstioIngressGateway, instance.Spec.ServiceRegistryConfig.IstioIngressGateway, "svcreg CR IstioIngressGateway should be updated")
	assert.Equal(t, svcreg.Spec.IstioServiceEntryRegistryNamespace, instance.Spec.ServiceRegistryConfig.IstioserviceEntryRegistryNamespace, "svcreg CR IstioserviceEntryRegistryNamespace should be updated")
}

func TestDeleteReconcileWithDeletionTimestamp(t *testing.T) {
	instance := newInstanceInDeletion()
	svcreg := newSvcreg()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{svcreg}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "DELETE svcreg reconcile should success")
	assert.Equal(t, res, true, "DELETE svcreg reconcile should return true")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-svcreg", Namespace: namespace}, svcreg)
	assert.Error(t, err, "svcreg CR should be deleted")
}

func TestDeleteReconcileWithDisable(t *testing.T) {
	instance := newInstanceWithDisable()
	svcreg := newSvcreg()

	scheme := scheme.Scheme
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, instance)
	scheme.AddKnownTypes(multicloudv1beta1.SchemeGroupVersion, svcreg)

	objs := []runtime.Object{svcreg}
	cl := fake.NewFakeClient(objs...)

	res, err := Reconcile(instance, cl, scheme)
	assert.NoError(t, err, "DELETE svcreg reconcile should success")
	assert.Equal(t, res, true, "DELETE svcreg reconcile should return true")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.ObjectMeta.Name + "-svcreg", Namespace: namespace}, svcreg)
	assert.Error(t, err, "svcreg CR should be deleted")
}

func TestIsReady(t *testing.T) {
	instance := newInstanceInDeletion()
	deploymentSvcreg := newTestDeployment(instance.ObjectMeta.Name + "-svcreg")
	deploymentCoredns := newTestDeployment(instance.ObjectMeta.Name + "-svcreg-coredns")

	objs := []runtime.Object{deploymentSvcreg, deploymentCoredns}
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
