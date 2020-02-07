// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of topology Defines the Reconciliation logic and required setup for topology collector CR.
package v1beta1

import (
	"context"
	"strings"

	openshiftsecurityv1 "github.com/openshift/api/security/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/utils"
)

var log = logf.Log.WithName("topology")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling TopologyCollector")

	topologyCollectorCR, err := newTopologyCollectorCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired TopologyCollector CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, topologyCollectorCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return false, err
	}

	// TODO(tonytran): split up weavescope and TopologyCollector
	foundTopologyCollector := &multicloudv1beta1.TopologyCollector{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: topologyCollectorCR.Name, Namespace: topologyCollectorCR.Namespace}, foundTopologyCollector)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("TopologyCollector CR DOES NOT exist")
			if instance.GetDeletionTimestamp() != nil {
				log.V(5).Info("instance IS in deletion state")
				err = finalize(instance, topologyCollectorCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE TopologyCollector CR")
					return false, err
				}

				reqLogger.Info("Successfully Reconciled TopologyCollector")
				return false, nil
			}

			if instance.Spec.TopologyCollectorConfig.Enabled {
				log.V(5).Info("TopologyCollector ENABLED")
				err = createServiceAccount(client, scheme, instance, topologyCollectorCR)
				if err != nil {
					log.Error(err, "Fail to CREATE ServiceAccount for TopologyCollector", "TopologyCollector.Name", topologyCollectorCR.Name)
					return false, err
				}

				log.Info("Creating a new TopologyCollector", "TopologyCollector.Namespace", topologyCollectorCR.Namespace, "ConnectionManager.Name", topologyCollectorCR.Name)
				err = client.Create(context.TODO(), topologyCollectorCR)
				if err != nil {
					log.Error(err, "Fail to CREATE TopologyCollector CR")
					return false, err
				}

				instance.Finalizers = append(instance.Finalizers, topologyCollectorCR.Name)
				reqLogger.Info("Successfully Reconciled TopologyCollector")
				return false, nil
			}

			log.V(5).Info("TopologyCollector DISABLED")
			err = finalize(instance, topologyCollectorCR, client)
			if err != nil {
				log.Error(err, "fail to FINALIZE TopologyCollector CR")
				return false, err
			}

			reqLogger.Info("Successfully Reconciled TopologyCollector")
			return false, nil
		}

		log.Error(err, "Unexpected ERROR")
		return false, err
	}

	if foundTopologyCollector.GetDeletionTimestamp() == nil {
		if instance.GetDeletionTimestamp() != nil || !instance.Spec.TopologyCollectorConfig.Enabled {
			err = client.Delete(context.TODO(), topologyCollectorCR)
			if err != nil {
				log.Error(err, "Fail to DELETE TopologyCollector CR")
				return false, err
			}

			reqLogger.Info("Requeueing Reconcile for TopologyCollector")
			return true, nil
		}

		// Endpoint NOT in deletion state AND found, update
		foundTopologyCollector.Spec = topologyCollectorCR.Spec
		err = client.Update(context.TODO(), foundTopologyCollector)
		if err != nil {
			log.Error(err, "Fail to UPDATE TopologyCollector CR")
			return false, nil
		}

		// Adding Finalizer to instance if Finalizer does not exist
		// NOTE: This is to handle requeue due to failed instance update during creation
		for _, finalizer := range instance.Finalizers {
			if finalizer == topologyCollectorCR.Name {
				return false, nil
			}
		}
		instance.Finalizers = append(instance.Finalizers, topologyCollectorCR.Name)
	} else {
		reqLogger.Info("Requeueing Reconcile for TopologyCollector")
		return true, nil
	}

	reqLogger.Info("Successfully Reconciled TopologyCollector")
	return false, nil
}

func newTopologyCollectorCR(cr *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.TopologyCollector, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	weaveImage, err := cr.GetImage("weave")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "weave")
		return nil, err
	}

	collectorImage, err := cr.GetImage("topology-collector")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "topology-collector")
		return nil, err
	}

	routerImage, err := cr.GetImage("router")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "routers")
		return nil, err
	}

	return &multicloudv1beta1.TopologyCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-topology",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.TopologyCollectorSpec{
			FullNameOverride:  cr.Name + "-topology",
			ClusterName:       cr.Spec.ClusterName,
			ClusterNamespace:  cr.Spec.ClusterNamespace,
			ConnectionManager: cr.Name + "-connmgr",
			ContainerRuntime:  determineRuntime(client),
			Enabled:           true,
			UpdateInterval:    cr.Spec.TopologyCollectorConfig.CollectorUpdateInterval,
			CACertIssuer:      cr.Name + "-self-signed",
			ServiceAccount: multicloudv1beta1.TopologyCollectorServiceAccount{
				Name: cr.Name + "-topology-collector",
			},
			WeaveImage:      weaveImage,
			CollectorImage:  collectorImage,
			RouterImage:     routerImage,
			ImagePullSecret: cr.Spec.ImagePullSecret,
		},
	}, nil
}

func determineRuntime(kubeclient client.Client) string {
	nodelist := &corev1.NodeList{}
	if err := kubeclient.List(context.TODO(), nodelist, &client.ListOptions{}); err != nil {
		log.Error(err, "Error listing nodes in cluster, assuming ContainerRuntime is docker")
		return "docker"
	}
	runtime := nodelist.Items[0].Status.NodeInfo.ContainerRuntimeVersion
	return strings.Split(runtime, ":")[0] //format of container runtime in node info is runtime://version
}

func createServiceAccount(client client.Client, scheme *runtime.Scheme, instance *multicloudv1beta1.Endpoint, topology *multicloudv1beta1.TopologyCollector) error {
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
	if err == nil {
		user := "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name
		log.Info("Adding User to SCC", "User", user, "SCC", foundPrivilegedSCC.Name)
		foundPrivilegedSCC.Users = append(foundPrivilegedSCC.Users, user)
		foundPrivilegedSCC.Users = utils.UniqueStringSlice(foundPrivilegedSCC.Users)
		err = client.Update(context.TODO(), foundPrivilegedSCC)
		if err != nil {
			return err
		}
	}
	return nil
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.TopologyCollector, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Delete Secrets
			secretsToDeletes := []string{
				cr.Name + "-ca-cert",
				cr.Name + "-server-secret",
				cr.Name + "-client-secret",
			}

			for _, secretToDelete := range secretsToDeletes {
				foundSecretToDelete := &corev1.Secret{}
				err := client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: cr.Namespace}, foundSecretToDelete)
				if err == nil {
					err = client.Delete(context.TODO(), foundSecretToDelete)
					if err != nil {
						log.Error(err, "Fail to DELETE TopologyCollector Secret", "Secret.Name", secretToDelete)
						return err
					}
				}
			}
			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
