/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package tiller

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

var log = logf.Log.WithName("tiller")

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("KlusterletService.Namespace", instance.Namespace, "KlusterletService.Name", instance.Name)
	reqLogger.Info("Reconciling Tiller")

	// No Tiller Integration
	if !instance.Spec.TillerIntegration.Enabled {
		log.Info("Tiller Integration disabled, skip Tiller Reconcile.")
		return nil
	}

	// ICP Tiller
	foundICPTillerService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, foundICPTillerService)
	if err == nil {
		log.Info("Found ICP Tiller, skip TillerCR Reconcile.")
		return nil
	}

	// No ICP Tiller
	tillerCR, err := newTillerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired Tiller CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, tillerCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundTillerCR := &klusterletv1alpha1.Tiller{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: tillerCR.Name, Namespace: tillerCR.Namespace}, foundTillerCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Tiller CR does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				if instance.Spec.TillerIntegration.Enabled {
					// Tiller Integration Enabled
					log.Info("Creating a new Tiller", "Tiller.Namespace", tillerCR.Namespace, "Tiller.Name", tillerCR.Name)
					err = client.Create(context.TODO(), tillerCR)
					if err != nil {
						log.Error(err, "Fail to CREATE Tiller CR")
						return err
					}

					// Adding Finalizer to KlusterletService instance
					instance.Finalizers = append(instance.Finalizers, tillerCR.Name)
				}
			} else {
				// Delete Secrets
				secretsToDeletes := []string{
					tillerCR.Name + "-ca-cert",
					tillerCR.Name + "-server-cert",
				}

				for _, secretToDelete := range secretsToDeletes {
					foundSecretToDelete := &corev1.Secret{}
					err = client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: tillerCR.Namespace}, foundSecretToDelete)
					if err == nil {
						err = client.Delete(context.TODO(), foundSecretToDelete)
						if err != nil {
							log.Error(err, "Fail to DELETE ConnectionManager Secret", "Secret.Name", secretToDelete)
							return err
						}
					}
				}

				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == tillerCR.Name {
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
		if foundTillerCR.GetDeletionTimestamp() == nil {
			// Tiller CR does exist
			if instance.GetDeletionTimestamp() == nil && instance.Spec.TillerIntegration.Enabled {
				// KlusterletService NOT in deletion state
				foundTillerCR.Spec = tillerCR.Spec
				err = client.Update(context.TODO(), foundTillerCR)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE ConnectionManager CR")
					return err
				}
			} else {
				// KlusterletService in deletion state or tillerIntegration disabled
				if foundTillerCR.GetDeletionTimestamp() == nil {
					// Delete Tiller CR
					err = client.Delete(context.TODO(), foundTillerCR)
					if err != nil {
						log.Error(err, "Fail to DELETE Tiller CR")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled Tiller")
	return nil
}

func newTillerCR(cr *klusterletv1alpha1.KlusterletService) (*klusterletv1alpha1.Tiller, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	image, err := cr.GetImage("tiller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "tiller")
		return nil, err
	}

	return &klusterletv1alpha1.Tiller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-tiller",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.TillerSpec{
			FullNameOverride: cr.Name + "-tiller",
			CACertIssuer:     cr.Name + "-self-signed",
			DefaultAdminUser: cr.Name + "-admin",
			Image:            image,
			ImagePullSecret:  cr.Spec.ImagePullSecret,
		},
	}, nil
}
