// Package v1beta1 of tiller provides a reconciler for the Tiller
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

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

// TODO(liuhao): switch from klusterletv1alpha1 to multicloudv1beta1 for the component api

var log = logf.Log.WithName("tiller")

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Tiller")

	// ICP Tiller
	foundICPTillerService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, foundICPTillerService)
	if err == nil {
		log.Info("Found ICP Tiller, skip TillerCR Reconcile.")
		return false, nil
	}

	// No ICP Tiller
	tillerCR, err := newTillerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired Tiller CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, tillerCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundTillerCR := &multicloudv1beta1.Tiller{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: tillerCR.Name, Namespace: tillerCR.Namespace}, foundTillerCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Tiller CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.TillerIntegration.Enabled {
					log.V(5).Info("TillerIntegration ENABLED")
					err = create(instance, tillerCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE Tiller CR")
						return false, err
					}
				} else {
					log.V(5).Info("TillerIntegration DISABLED")
					err = finalize(instance, tillerCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE Tiller CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err = finalize(instance, tillerCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE Tiller CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("Tiller CR DOES exist")
		if foundTillerCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("Tiller CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.TillerIntegration.Enabled {
				log.Info("instance IS NOT in deletion state and TillerIntegration ENABLED")
				err = update(instance, tillerCR, foundTillerCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE Tiller CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or TillerIntegration DISABLED")
				err = delete(foundTillerCR, client)
				if err != nil {
					log.Error(err, "Fail to DELETE Tiller CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for Tiller")
				return true, nil
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for Tiller")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled Tiller")
	return false, nil
}

// TODO(liuhao): the following method need to be refactored as instance method of Tiller struct
func newTillerCR(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.Tiller, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	image, err := instance.GetImage("tiller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "tiller")
		return nil, err
	}

	return &multicloudv1beta1.Tiller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-tiller",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.TillerSpec{
			FullNameOverride: instance.Name + "-tiller",
			CACertIssuer:     instance.Name + "-self-signed",
			DefaultAdminUser: instance.Name + "-admin",
			Image:            image,
			ImagePullSecret:  instance.Spec.ImagePullSecret,
			KubeClusterType:  "noticp",
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Tiller, client client.Client) error {
	log.Info("Creating a new Tiller", "Tiller.Namespace", cr.Namespace, "Tiller.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE Tiller CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Tiller, foundCR *multicloudv1beta1.Tiller, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE Tiller CR")
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

func delete(foundCR *multicloudv1beta1.Tiller, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Tiller, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Delete Secrets
			secretsToDeletes := []string{
				instance.Name + "-tiller-ca-cert",
				instance.Name + "-tiller-server-cert",
				instance.Name + "-workmgr-tiller-client-certs",
				instance.Name + "-search-tiller-client-certs",
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
