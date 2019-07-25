//Package topology Defines the Reconciliation logic and required setup for topology collector CR.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package topology

import (
	"context"

	openshiftsecurityv1 "github.com/openshift/api/security/v1"
	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("topology")

//Reconcile business logic for topology
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	topologyCollector := newTopologyCollectorCR(instance)
	if !instance.Spec.TopologyIntegration.Enabled {
		err := client.Delete(context.TODO(), topologyCollector)
		if err != nil && errors.IsNotFound(err) {
			log.Info("No existing topology collector found to delete.")
			return nil
		}
		if err != nil {
			log.Error(err, "Ran into error trying to delete existing topology collector")
			return err
		}
		log.Info("Topology disabled, skip topology reconciling")
		return nil
	}

	err := controllerutil.SetControllerReference(instance, topologyCollector, scheme)
	if err != nil {
		log.Error(err, "Ran into error trying to set topology controller referenc")
		return err
	}

	foundTopologyCollector := &klusterletv1alpha1.TopologyCollector{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: topologyCollector.Name, Namespace: topologyCollector.Namespace}, foundTopologyCollector)
	if err != nil && errors.IsNotFound(err) {
		runtime, err := determineRuntime(client)
		if err != nil {
			log.Error(err, "Error determining cluster node runtime")
			return err
		}
		topologyCollector.Spec.ContainerRuntime = runtime
		err = createServiceAccount(client, scheme, instance, topologyCollector)
		if err != nil {
			log.Error(err, "Error creating Service Account for topology collector")
			return err
		}

		err = client.Create(context.TODO(), topologyCollector)
		if err != nil {
			log.Error(err, "Error creating topology collector instance")
			return err
		}
		log.Info("Created topology collector instance successfully")
		return nil
	}

	if err != nil {
		log.Error(err, "Error retrieving existing topology collector instance")
		return err
	}

	foundTopologyCollector.Spec = topologyCollector.Spec        //for now we will update using new spec
	err = client.Update(context.TODO(), foundTopologyCollector) //update existing
	if err != nil {
		log.Error(err, "Error updating existing topology collector")
		return err
	}
	log.Info("Successfully updated topology collector instance")
	return nil
}

func determineRuntime(kubeclient client.Client) (string, error) {
	nodelist := &corev1.NodeList{}
	//ops.Namespace = ""
	err := kubeclient.List(context.TODO(), &client.ListOptions{}, nodelist)
	if err != nil {
		log.Error(err, "Error listing nodes in cluster")
		return "", err
	}
	runtime := nodelist.Items[0].Status.NodeInfo.ContainerRuntimeVersion
	return strings.Split(runtime, ":")[0], nil //format of container runtime in node info is runtime://version

}

func newTopologyCollectorCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.TopologyCollector {
	labels := map[string]string{
		"app": cr.Name,
	}

	return &klusterletv1alpha1.TopologyCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-topology",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.TopologyCollectorSpec{

			FullNameOverride:  cr.Name + "-topology",
			ClusterName:       cr.Spec.ClusterName,
			ClusterNamespace:  cr.Spec.ClusterNamespace,
			ConnectionManager: cr.Name + "-connmgr",
			ContainerRuntime:  "", //"docker",
			Enabled:           true,
			UpdateInterval:    cr.Spec.TopologyIntegration.CollectorUpdateInterval,
			CACertIssuer:      cr.Name + "-self-signed",
			ServiceAccount: klusterletv1alpha1.TopologyCollectorServiceAccount{
				Name: cr.Name + "-topology-collector",
			},
		},
	}
}

func createServiceAccount(client client.Client, scheme *runtime.Scheme, instance *klusterletv1alpha1.KlusterletService, topology *klusterletv1alpha1.TopologyCollector) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      topology.Spec.ServiceAccount.Name,
			Namespace: topology.Namespace,
		},
	}
	err := controllerutil.SetControllerReference(instance, serviceAccount, scheme)
	if err != nil {
		return err
	}

	foundServiceAccount := &corev1.ServiceAccount{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ServiceAccount", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
		err = client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	foundPrivilegedSCC := &openshiftsecurityv1.SecurityContextConstraints{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: "privileged", Namespace: ""}, foundPrivilegedSCC)
	// if client.Get return error that means no privileged SCC in that case skip adding user to scc and ignore error
	if err != nil {
		user := "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name
		log.Info("Adding User to SCC", "User", user, "SCC", foundPrivilegedSCC.Name)
		foundPrivilegedSCC.Users = append(foundPrivilegedSCC.Users, user)
		err = client.Update(context.TODO(), foundPrivilegedSCC)
		if err != nil {
			return err
		}
	}
	return nil
}
