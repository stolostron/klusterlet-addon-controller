// Package v1beta1 of migration Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"math/rand"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	appmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/appmgr/v1beta1"
	connmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/connmgr/v1beta1"
	policyctrl "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/policyctrl/v1beta1"
	searchcollector "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/searchcollector/v1beta1"
	serviceregistry "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/serviceregistry/v1beta1"
	topology "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/topology/v1beta1"
	workmgr "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/workmgr/v1beta1"
)

var log = logf.Log.WithName("migration")

const (
	// Connmgr ConnectionManager
	Connmgr string = "connmgr"
	// Workmgr WorkManager
	Workmgr string = "workmgr"
	// SvcReg ServiceRegistry
	SvcReg string = "svcreg"
	// CoreDNS CoreDNS
	CoreDNS string = "coreDNS"
	// PolicyCtrl PolicyController
	PolicyCtrl string = "policyctrl"
	// SearchCollector SearchCollector
	SearchCollector string = "searchcollector"
)

// Reconcile Resolves the migration from 3.2.0 to 3.2.1.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Migration")

	err := migrateSecrets(instance, client)
	if err != nil {
		return false, err
	}

	completed := true

	// ConnectionManager
	connMgrMigrated, err := migrateConnectionManager(instance, client)
	if err != nil {
		return false, err
	}
	if connMgrMigrated {
		log.Info("Connmgr Migrated")
	}
	completed = completed && connMgrMigrated

	// WorkManager
	workMgrMigrated, err := migrateWorkManager(instance, client)
	if err != nil {
		return false, err
	}
	if workMgrMigrated {
		log.Info("Workmgr Migrated")
	}
	completed = completed && workMgrMigrated

	// ServiceRegistry
	svcRegMigrated, err := migrateServiceRegistry(instance, client)
	if err != nil {
		return false, err
	}
	if svcRegMigrated {
		log.Info("Svcreg Migrated")
	}
	completed = completed && svcRegMigrated

	// PolicyController
	policyCtrlMigrated, err := migratePolicyController(instance, client)
	if err != nil {
		return false, err
	}
	if policyCtrlMigrated {
		log.Info("Policy Migrated")
	}
	completed = completed && policyCtrlMigrated

	// SearchCollector
	searchMigrated, err := migrateSearchCollector(instance, client)
	if err != nil {
		return false, err
	}
	if searchMigrated {
		log.Info("Search Migrated")
	}
	completed = completed && searchMigrated

	// TopologyCollector
	topologyMigrated, err := migrateTopologyCollector(instance, client)
	if err != nil {
		return false, err
	}
	if topologyMigrated {
		log.Info("Topology Migrated")
	}
	completed = completed && topologyMigrated

	if completed {
		instance.Spec.Migration = false
		return false, nil
	}

	reqLogger.Info("Successfully Reconciled Migration")
	return true, nil
}

func migrateTopologyCollector(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	const weaveCollector string = "multicluster-endpoint-topology-weave-collector"
	const weaveApp string = "multicluster-endpoint-topology-weave-scope-app"
	const weaveScope string = "multicluster-endpoint-topology-weave-scope"

	oldDeploymentExist, err := deploymentExist(client, weaveCollector, instance.Namespace)
	if err != nil {
		return false, err
	}
	if !oldDeploymentExist {
		log.V(5).Info("TopologyCollector is disable in 3.2.0, no need for migration.")
		return true, nil
	}

	log.V(5).Info("Enable TopologyCollector")
	instance.Spec.TopologyCollectorConfig.Enabled = true

	topologyReady, err := topology.IsReady(instance, client)
	if err != nil {
		return false, err
	}
	if topologyReady {
		err = scaleDownDeployment(client, weaveCollector, instance.Namespace)
		if err != nil {
			return false, err
		}

		err = scaleDownDeployment(client, weaveApp, instance.Namespace)
		if err != nil {
			return false, err
		}

		err = scaleDownDaemonset(client, weaveScope, instance.Namespace)
		if err != nil {
			return false, err
		}
	}

	return topologyReady, nil
}

func migrateSearchCollector(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	const searchDeployment string = "multicluster-endpoint-search-search-collector"

	oldDeploymentExist, err := deploymentExist(client, searchDeployment, instance.Namespace)
	if err != nil {
		return false, err
	}
	if !oldDeploymentExist {
		log.V(5).Info("SearchCollector is disable in 3.2.0, no need for migration.")
		return true, nil
	}

	log.V(5).Info("Enable SearchCollector")
	instance.Spec.SearchCollectorConfig.Enabled = true

	searchCollectorIsReady, err := searchcollector.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	if searchCollectorIsReady {
		err = scaleDownDeployment(client, searchDeployment, instance.Namespace)
		if err != nil {
			return false, err
		}
	}

	return searchCollectorIsReady, nil
}

func migratePolicyController(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	const policyDeployment string = "multicluster-endpoint-policy-compliance"

	oldDeploymentExist, err := deploymentExist(client, policyDeployment, instance.Namespace)
	if err != nil {
		return false, err
	}
	if !oldDeploymentExist {
		log.V(5).Info("PolicyController is disable in 3.2.0, no need for migration.")
		return true, nil
	}

	log.V(5).Info("Enabling PolicyController")
	instance.Spec.PolicyController.Enabled = true

	policyCtrlIsReady, err := policyctrl.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	if policyCtrlIsReady {
		err = scaleDownDeployment(client, policyDeployment, instance.Namespace)
		if err != nil {
			return false, err
		}
	}

	return policyCtrlIsReady, nil
}

func migrateServiceRegistry(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	const svcreg string = "multicluster-endpoint-ibm-klusterlet-service-registry"
	const coredns string = "multicluster-endpoint-ibm-klusterlet-coredns"

	oldDeploymentExist, err := deploymentExist(client, svcreg, instance.Namespace)
	if err != nil {
		return false, err
	}
	if !oldDeploymentExist {
		log.V(5).Info("ServiceRegistry is disable in 3.2.0, no need for migration.")
		return true, nil
	}

	log.V(5).Info("Enabling service registry")
	instance.Spec.ServiceRegistryConfig.Enabled = true

	svcregIsReady, err := serviceregistry.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	if svcregIsReady {
		err = scaleDownDeployment(client, svcreg, instance.Namespace)
		if err != nil {
			return false, err
		}

		err = scaleDownDeployment(client, coredns, instance.Namespace)
		if err != nil {
			return false, err
		}
	}

	return svcregIsReady, nil
}

func migrateConnectionManager(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	connmgrIsReady, err := connmgr.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	if connmgrIsReady {
		err = scaleDownDeployment(client, "multicluster-endpoint-ibm-klusterlet-operator", instance.Namespace)
		if err != nil {
			return false, err
		}
	}

	return connmgrIsReady, nil
}

func migrateWorkManager(instance *multicloudv1beta1.Endpoint, client client.Client) (bool, error) {
	log.V(5).Info("Enabling ApplicationManager")
	instance.Spec.ApplicationManagerConfig.Enabled = true

	workmgrIsReady, err := workmgr.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	appmgrIsReady, err := appmgr.IsReady(instance, client)
	if err != nil {
		return false, err
	}

	if workmgrIsReady && appmgrIsReady {
		err = scaleDownDeployment(client, "multicluster-endpoint-ibm-klusterlet-klusterlet", instance.Namespace)
		if err != nil {
			return false, err
		}
	}
	return (workmgrIsReady && appmgrIsReady), err
}

func scaleDownDeployment(client client.Client, name string, namespace string) error {
	foundDeployment := &extensionsv1beta1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, foundDeployment)
	if err != nil {
		return err
	}

	if foundDeployment.Spec.Replicas == nil || int(*foundDeployment.Spec.Replicas) != 0 {
		foundDeployment.Spec.Replicas = new(int32)
		err = client.Update(context.TODO(), foundDeployment)
		if err != nil {
			log.Error(err, "Fail to scale down deployment")
			return err
		}
	}

	return nil
}

func deploymentExist(client client.Client, deploymentName string, namespace string) (bool, error) {
	oldDeployment := &extensionsv1beta1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: namespace}, oldDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Component is disable in 3.2.0, no need for migration.")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	return true, nil
}

func scaleDownDaemonset(client client.Client, name string, namespace string) error {
	foundDaemonset := &extensionsv1beta1.DaemonSet{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, foundDaemonset)
	if err != nil {
		return err
	}

	if foundDaemonset.Spec.Template.Spec.NodeSelector == nil || len(foundDaemonset.Spec.Template.Spec.NodeSelector) == 0 {
		//generate random stream of 12 character key and value for node selector
		rand.Seed(time.Now().UnixNano())
		chars := []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		length := 12
		var keyBuilder strings.Builder
		var valueBuilder strings.Builder
		for i := 0; i < length; i++ {
			keyBuilder.WriteRune(chars[rand.Intn(len(chars))])
			valueBuilder.WriteRune(chars[rand.Intn(len(chars))])
		}

		foundDaemonset.Spec.Template.Spec.NodeSelector = map[string]string{
			keyBuilder.String(): valueBuilder.String(),
		}

		err = client.Update(context.TODO(), foundDaemonset)
		if err != nil {
			log.Error(err, "Fail to scale down Daemonset")
			return err
		}
	}

	return nil
}

func migrateSecrets(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	// HAO: this is probably not needed since the import process will create a new one anyway
	err := migrateSecret(client, "multicluster-endpoint-hub-kubeconfig", "klusterlet-bootstrap", instance.Namespace)
	if err != nil {
		return err
	}

	err = migrateSecret(client, "multicluster-endpoint-hub-kubeconfig", instance.Name+"-connmgr-hub-kubeconfig", instance.Namespace)
	if err != nil {
		return err
	}

	err = migrateSecret(client, "multicluster-endpoint-ibm-klusterlet-cert-store", instance.Name+"-connmgr-cert-store", instance.Namespace)
	if err != nil {
		return err
	}

	return nil
}

func migrateSecret(client client.Client, oldName string, newName string, namespace string) error {
	newSecret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: newName, Namespace: namespace}, newSecret)
	if err == nil {
		log.V(5).Info("secret already exist", "Secret.Name", newName)
		return nil
	}

	if !errors.IsNotFound(err) {
		log.Error(err, "Unexpected ERROR")
		return err
	}

	log.V(5).Info("migrating secret", "oldName", oldName, "newName", newName)
	oldSecret := &corev1.Secret{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: oldName, Namespace: namespace}, oldSecret)
	if err != nil {
		return err
	}

	newSecret.Name = newName
	newSecret.Namespace = oldSecret.Namespace
	newSecret.Data = oldSecret.Data
	newSecret.Type = oldSecret.Type
	err = client.Create(context.TODO(), newSecret)

	if err != nil {
		log.Error(err, "Fail to migrate secret", "oldSecret", oldSecret, "newSecret", newSecret)
		return err
	}

	log.V(5).Info("Successfully migrated secret", "oldName", oldName, "newName", newName)
	return nil
}
