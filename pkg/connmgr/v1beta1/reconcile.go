//Package v1beta1 of connmgrs Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO(liuhao): switch from klusterletv1alpha1.ConnectionManager to multicloudv1beta1.ConnectionManager

var log = logf.Log.WithName("connmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("KlusterletService.Namespace", instance.Namespace, "KlusterletService.Name", instance.Name)
	reqLogger.Info("Reconciling ConnectionManager")

	connMgrCR, err := newConnectionManagerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired ConnectionManager CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, connMgrCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundConnMgrCR := &klusterletv1alpha1.ConnectionManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: connMgrCR.Name, Namespace: connMgrCR.Namespace}, foundConnMgrCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("ConnectionManager CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("Instance IS NOT in deletion state")
				err := create(instance, connMgrCR, client)
				if err != nil {
					log.Error(err, "fail to CREATE ConnectionManager CR")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				err := finalize(instance, connMgrCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE ConnectionManager CR")
					return err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		log.V(5).Info("ConnectionManager CR DOES exist")
		if foundConnMgrCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("ConnectionManager CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil {
				log.Info("Instance IS NOT in deletion state")
				err := update(instance, connMgrCR, foundConnMgrCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE ConnectionManager CR")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				if foundConnMgrCR.GetDeletionTimestamp() == nil {
					err := delete(foundConnMgrCR, client)
					if err != nil {
						log.Error(err, "Fail to DELETE ConnectionManager CR")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled ConnectionManager")
	return nil
}

func newConnectionManagerCR(cr *multicloudv1beta1.Endpoint) (*klusterletv1alpha1.ConnectionManager, error) {
	image, err := cr.GetImage("connection-manager")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "connection-manager")
		return nil, err
	}

	labels := map[string]string{
		"app": cr.Name,
	}

	return &klusterletv1alpha1.ConnectionManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-connmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.ConnectionManagerSpec{
			ClusterName:      cr.Spec.ClusterName,
			ClusterNamespace: cr.Spec.ClusterNamespace,
			BootStrapConfig:  cr.Spec.BootStrapConfig,
			FullNameOverride: cr.Name + "-connmgr",
			Image:            image,
			ImagePullSecret:  cr.Spec.ImagePullSecret,
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.ConnectionManager, client client.Client) error {
	log.Info("Creating a new ConnectionManager", "ConnectionManager.Namespace", cr.Namespace, "ConnectionManager.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE ConnectionManager CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.ConnectionManager, foundCR *klusterletv1alpha1.ConnectionManager, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE ConnectionManager CR")
		return err
	}

	// Adding Finalizer to instance if Finalizer does not exist
	// NOTE: This is to handle requeue due to failed instance update during creation
	for _, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			return nil
		}
	}
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func delete(foundCR *klusterletv1alpha1.ConnectionManager, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.ConnectionManager, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Delete Secrets
			secretsToDeletes := []string{
				cr.Name + "-cert-store",
				cr.Name + "-hub-kubeconfig",
			}

			for _, secretToDelete := range secretsToDeletes {
				foundSecretToDelete := &corev1.Secret{}
				err := client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: cr.Namespace}, foundSecretToDelete)
				if err == nil {
					err := client.Delete(context.TODO(), foundSecretToDelete)
					if err != nil {
						log.Error(err, "Fail to DELETE ConnectionManager Secret", "Secret.Name", secretToDelete)
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
