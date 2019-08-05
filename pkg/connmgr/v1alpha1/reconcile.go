//Package v1alpha1 of connmgrs Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1alpha1

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("connmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
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
			// ConnectionManager CR does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				log.Info("Creating a new ConnectionManager", "ConnectionManager.Namespace", connMgrCR.Namespace, "ConnectionManager.Name", connMgrCR.Name)
				err := client.Create(context.TODO(), connMgrCR)
				if err != nil {
					log.Error(err, "Fail to CREATE ConnectionManager CR")
					return err
				}

				// Adding Finalizer to KlusterletService instance
				instance.Finalizers = append(instance.Finalizers, connMgrCR.Name)
			} else {
				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == connMgrCR.Name {
						// Cleanup Secrets
						secretsToDeletes := []string{
							connMgrCR.Name + "-cert-store",
							connMgrCR.Name + "-hub-kubeconfig",
						}

						for _, secretToDelete := range secretsToDeletes {
							foundSecretToDelete := &corev1.Secret{}
							err = client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: connMgrCR.Namespace}, foundSecretToDelete)
							if err == nil {
								err = client.Delete(context.TODO(), foundSecretToDelete)
								if err != nil {
									log.Error(err, "Fail to DELETE ConnectionManager Secret", "Secret.Name", secretToDelete)
									return err
								}
							}
						}
						instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
						break
					}
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		if foundConnMgrCR.GetDeletionTimestamp() == nil {
			// ConnectionManager CR does exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				foundConnMgrCR.Spec = connMgrCR.Spec
				err = client.Update(context.TODO(), foundConnMgrCR)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE ConnectionManager CR")
					return err
				}
			} else {
				// KlusterletService in deletion state
				err = client.Delete(context.TODO(), foundConnMgrCR)
				if err != nil {
					log.Error(err, "Fail to DELETE ConnectionManager CR")
					return err
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled ConnectionManager")
	return nil
}

func newConnectionManagerCR(cr *klusterletv1alpha1.KlusterletService) (*klusterletv1alpha1.ConnectionManager, error) {
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
