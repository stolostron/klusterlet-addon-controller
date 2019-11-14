// Package v1beta1 of migration Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
)

var (
	namespace = "multicluster-endpoint"
	replica   = int32(1)
)

func newDeployment(name string) *extensionsv1beta1.Deployment {
	deployment := &extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: extensionsv1beta1.DeploymentSpec{
			Replicas: &replica,
		},
	}
	return deployment
}

func newNewDeployment(name string) *extensionsv1beta1.Deployment {
	newDeployment := &extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: extensionsv1beta1.DeploymentSpec{
			Replicas: &replica,
		},
		Status: extensionsv1beta1.DeploymentStatus{
			Conditions: []extensionsv1beta1.DeploymentCondition{extensionsv1beta1.DeploymentCondition{
				Type:   "Available",
				Status: "True",
			}},
		},
	}
	return newDeployment
}

func newSecret(name string) *corev1.Secret {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return secret
}

func newDaemonset(name string) *extensionsv1beta1.DaemonSet {
	daemonset := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return daemonset
}

func newNewDaemonset(name string) *extensionsv1beta1.DaemonSet {
	daemonset := &extensionsv1beta1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: extensionsv1beta1.DaemonSetStatus{
			CurrentNumberScheduled: int32(1),
			NumberReady:            int32(1),
		},
	}
	return daemonset
}

func newTestClientMigrationClient() client.Client {
	deploymentWorkmgr := newDeployment("multicluster-endpoint-ibm-klusterlet-klusterlet")
	deploymentConnmgr := newDeployment("multicluster-endpoint-ibm-klusterlet-operator")
	deploymentSvcreg := newDeployment("multicluster-endpoint-ibm-klusterlet-service-registry")
	deploymentCoredns := newDeployment("multicluster-endpoint-ibm-klusterlet-coredns")
	deploymentPolicy := newDeployment("multicluster-endpoint-policy-compliance")
	deploymentSearch := newDeployment("multicluster-endpoint-search-search-collector")
	deploymentWeaveCollector := newDeployment("multicluster-endpoint-topology-weave-collector")
	deploymentWeaveApp := newDeployment("multicluster-endpoint-topology-weave-scope-app")

	deploymentNewWorkmgr := newNewDeployment("endpoint-workmgr")
	deploymentNewConnmgr := newNewDeployment("endpoint-connmgr")
	deploymentNewAppmgr := newNewDeployment("endpoint-appmgr")
	deploymentNewSvcreg := newNewDeployment("endpoint-svcreg")
	deploymentNewSvcregCoredns := newNewDeployment("endpoint-svcreg-coredns")
	deploymentNewPolicy := newNewDeployment("endpoint-policyctrl")
	deploymentNewSearch := newNewDeployment("endpoint-search")
	deploymentNewWeaveCollector := newNewDeployment("endpoint-topology-weave-collector")
	deploymentNewWeaveApp := newNewDeployment("endpoint-topology-weave-scope-app")

	daemonNewWeavescope := newNewDaemonset("endpoint-topology-weave-scope")

	secretHubKubeconfig := newSecret("multicluster-endpoint-hub-kubeconfig")
	secretCertStore := newSecret("multicluster-endpoint-ibm-klusterlet-cert-store")

	daemonsetWeavescope := newDaemonset("multicluster-endpoint-topology-weave-scope")

	objs := []runtime.Object{deploymentWorkmgr,
		deploymentConnmgr, deploymentSvcreg, deploymentCoredns, deploymentPolicy, deploymentSearch, deploymentWeaveCollector,
		deploymentWeaveApp, deploymentNewWorkmgr, deploymentNewConnmgr, deploymentNewAppmgr, deploymentNewSvcreg,
		deploymentNewSvcregCoredns, deploymentNewPolicy, deploymentNewSearch, deploymentNewWeaveCollector, deploymentNewWeaveApp,
		daemonNewWeavescope, secretHubKubeconfig, secretCertStore, daemonsetWeavescope,
	}
	cl := fake.NewFakeClient(objs...)
	return cl
}

func newInstance() *multicloudv1beta1.Endpoint {
	instance := &multicloudv1beta1.Endpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint",
			Namespace: namespace,
		},
	}
	return instance
}

func TestReconcile(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := Reconcile(instance, cl)
	if err != nil {
		t.Fatalf("Reconcile error: (%v)", err)
	}

	assert.False(t, res, "migrate reconcile res should be false")

	deploymentWorkmgr := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-klusterlet", Namespace: namespace}, deploymentWorkmgr)
	assert.NoError(t, err, "old workmgr deployment should exist")
	assert.Equal(t, *deploymentWorkmgr.Spec.Replicas, int32(0), "old workmgr deployment should be scaled down")

	deploymentConnmgr := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-operator", Namespace: namespace}, deploymentConnmgr)
	assert.NoError(t, err, "old connmgr deployment should exist")
	assert.Equal(t, *deploymentConnmgr.Spec.Replicas, int32(0), "old connmgr deployment should be scaled down")

	deploymentSvcreg := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-service-registry", Namespace: namespace}, deploymentSvcreg)
	assert.NoError(t, err, "old svcreg deployment should exist")
	assert.Equal(t, *deploymentSvcreg.Spec.Replicas, int32(0), "old svcreg deployment should be scaled down")

	deploymentCoredns := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-coredns", Namespace: namespace}, deploymentCoredns)
	assert.NoError(t, err, "old Coredns deployment should exist")
	assert.Equal(t, *deploymentWorkmgr.Spec.Replicas, int32(0), "old Coredns deployment should be scaled down")

	deploymentPolicy := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-policy-compliance", Namespace: namespace}, deploymentPolicy)
	assert.NoError(t, err, "old Policy deployment should exist")
	assert.Equal(t, *deploymentPolicy.Spec.Replicas, int32(0), "old Policy deployment should be scaled down")

	deploymentSearch := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-search-search-collector", Namespace: namespace}, deploymentSearch)
	assert.NoError(t, err, "old Search deployment should exist")
	assert.Equal(t, *deploymentSearch.Spec.Replicas, int32(0), "old Search deployment should be scaled down")

	deploymentWeaveCollector := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-collector", Namespace: namespace}, deploymentWeaveCollector)
	assert.NoError(t, err, "old WeaveCollector deployment should exist")
	assert.Equal(t, *deploymentWeaveCollector.Spec.Replicas, int32(0), "old WeaveCollector deployment should be scaled down")

	deploymentWeaveApp := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-scope-app", Namespace: namespace}, deploymentWeaveApp)
	assert.NoError(t, err, "old WeaveApp deployment should exist")
	assert.Equal(t, *deploymentWeaveApp.Spec.Replicas, int32(0), "old WeaveApp deployment should be scaled down")

	daemonsetWeavescope := &extensionsv1beta1.DaemonSet{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-scope", Namespace: namespace}, daemonsetWeavescope)
	assert.NoError(t, err, "old Weavescope daemonset should exist")
	assert.NotEmpty(t, daemonsetWeavescope.Spec.Template.Spec.NodeSelector, "old Weavescope daemonset should be scaled down")

	assert.False(t, instance.Spec.Migration, "the migration parameter in the instance should be false")
}

func TestMigrateTopologyCollector(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migrateTopologyCollector(instance, cl)

	assert.NoError(t, err, "migrateTopologyCollector should success")
	assert.Equal(t, res, true, "migrateTopologyCollector res should be true")

	deploymentWeaveCollector := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-collector", Namespace: namespace}, deploymentWeaveCollector)
	assert.NoError(t, err, "old WeaveCollector deployment should exist")
	assert.Equal(t, *deploymentWeaveCollector.Spec.Replicas, int32(0), "old WeaveCollector deployment should be scaled down")

	deploymentWeaveApp := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-scope-app", Namespace: namespace}, deploymentWeaveApp)
	assert.NoError(t, err, "old WeaveApp deployment should exist")
	assert.Equal(t, *deploymentWeaveApp.Spec.Replicas, int32(0), "old WeaveApp deployment should be scaled down")

	daemonsetWeavescope := &extensionsv1beta1.DaemonSet{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-scope", Namespace: namespace}, daemonsetWeavescope)
	assert.NoError(t, err, "old Weavescope daemonset should exist")
	assert.NotEmpty(t, daemonsetWeavescope.Spec.Template.Spec.NodeSelector, "old Weavescope daemonset should be scaled down")

	assert.False(t, instance.Spec.Migration, "the migration parameter in the instance should be false")
}

func TestMigrateSearchCollector(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migrateSearchCollector(instance, cl)

	assert.NoError(t, err, "migrateSearchCollector should success")
	assert.Equal(t, res, true, "migrateSearchCollector res should be true")

	deploymentSearch := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-search-search-collector", Namespace: namespace}, deploymentSearch)
	assert.NoError(t, err, "old Search deployment should exist")
	assert.Equal(t, *deploymentSearch.Spec.Replicas, int32(0), "old Search deployment should be scaled down")
}

func TestMigratePolicyController(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migratePolicyController(instance, cl)

	assert.NoError(t, err, "migratePolicyController should success")
	assert.Equal(t, res, true, "migratePolicyController res should be true")

	deploymentPolicy := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-policy-compliance", Namespace: namespace}, deploymentPolicy)
	assert.NoError(t, err, "old Policy deployment should exist")
	assert.Equal(t, *deploymentPolicy.Spec.Replicas, int32(0), "old Policy deployment should be scaled down")
}

func TestMigrateServiceRegistry(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migrateServiceRegistry(instance, cl)

	assert.NoError(t, err, "migrateServiceRegistry should success")
	assert.Equal(t, res, true, "migrateServiceRegistry res should be true")

	deploymentSvcreg := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-service-registry", Namespace: namespace}, deploymentSvcreg)
	assert.NoError(t, err, "old svcreg deployment should exist")
	assert.Equal(t, *deploymentSvcreg.Spec.Replicas, int32(0), "old svcreg deployment should be scaled down")

	deploymentCoredns := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-coredns", Namespace: namespace}, deploymentCoredns)
	assert.NoError(t, err, "old Coredns deployment should exist")
	assert.Equal(t, *deploymentCoredns.Spec.Replicas, int32(0), "old Coredns deployment should be scaled down")
}

func TestMigrateConnectionManager(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migrateConnectionManager(instance, cl)

	assert.NoError(t, err, "migrateConnectionManager should success")
	assert.Equal(t, res, true, "migrateConnectionManager res should be true")

	deploymentConnmgr := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-operator", Namespace: namespace}, deploymentConnmgr)
	assert.NoError(t, err, "old connmgr deployment should exist")
	assert.Equal(t, *deploymentConnmgr.Spec.Replicas, int32(0), "old connmgr deployment should be scaled down")
}

func TestMigrateWorkManager(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	res, err := migrateWorkManager(instance, cl)

	assert.NoError(t, err, "migrateWorkManager should success")
	assert.Equal(t, res, true, "migrateWorkManager res should be true")

	deploymentWorkmgr := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-klusterlet", Namespace: namespace}, deploymentWorkmgr)
	assert.NoError(t, err, "old workmgr deployment should exist")
	assert.Equal(t, *deploymentWorkmgr.Spec.Replicas, int32(0), "old workmgr deployment should be scaled down")
}

func TestScaleDownDeployment(t *testing.T) {
	cl := newTestClientMigrationClient()

	err := scaleDownDeployment(cl, "multicluster-endpoint-ibm-klusterlet-klusterlet", namespace)
	assert.NoError(t, err, "scale down deployment should success")

	foundDeployment := &extensionsv1beta1.Deployment{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-ibm-klusterlet-klusterlet", Namespace: namespace}, foundDeployment)
	assert.NoError(t, err, "find deployment should success")
	assert.Equal(t, *foundDeployment.Spec.Replicas, int32(0), "should scale down the deployment, raplica should be 0")
}

func TestDeploymentExist(t *testing.T) {
	cl := newTestClientMigrationClient()

	res, err := deploymentExist(cl, "multicluster-endpoint-ibm-klusterlet-klusterlet", namespace)

	assert.NoError(t, err, "check deployment exist should success")
	assert.Equal(t, res, true, "'multicluster-endpoint-ibm-klusterlet-klusterlet' should not exist")

	res, err = deploymentExist(cl, "mcm", namespace)

	assert.NoError(t, err, "check deployment exist should success")
	assert.False(t, res, "'multicluster-endpoint-ibm-klusterlet-klusterlet' should exist")
}

func TestScaleDownDaemonset(t *testing.T) {
	cl := newTestClientMigrationClient()

	err := scaleDownDaemonset(cl, "multicluster-endpoint-topology-weave-scope", "multicluster-endpoint")
	assert.NoError(t, err, "scale down daemonset should success")

	daemonsetWeavescope := &extensionsv1beta1.DaemonSet{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "multicluster-endpoint-topology-weave-scope", Namespace: namespace}, daemonsetWeavescope)
	assert.NoError(t, err, "old Weavescope daemonset should exist")
	assert.NotEmpty(t, daemonsetWeavescope.Spec.Template.Spec.NodeSelector, "old Weavescope daemonset should be scaled down")
}

func TestMigrateSecrets(t *testing.T) {
	cl := newTestClientMigrationClient()
	instance := newInstance()

	err := migrateSecrets(instance, cl)
	assert.NoError(t, err, "migrateSecrets should success")

	foundSecret := &corev1.Secret{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "klusterlet-bootstrap", Namespace: namespace}, foundSecret)
	assert.NoError(t, err, "migrate secret KlusterletBootstrap should success")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-connmgr-hub-kubeconfig", Namespace: namespace}, foundSecret)
	assert.NoError(t, err, "migrate secret HubKubeconfig should success")

	err = cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-connmgr-cert-store", Namespace: namespace}, foundSecret)
	assert.NoError(t, err, "migrate secret CertStore should success")
}

func TestMigrateSecret(t *testing.T) {
	cl := newTestClientMigrationClient()

	err := migrateSecret(cl, "multicluster-endpoint-hub-kubeconfig", "klusterlet-bootstrap", "multicluster-endpoint")
	assert.NoError(t, err, "migrateSecret should success")

	foundSecret := &corev1.Secret{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "klusterlet-bootstrap", Namespace: namespace}, foundSecret)
	assert.NoError(t, err, "migrate secret KlusterletBootstrap should success")
}
