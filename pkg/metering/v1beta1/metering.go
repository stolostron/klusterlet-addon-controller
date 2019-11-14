// Package v1beta1 of metering provides a reconciler for the Metering
// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"bytes"
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
)

var log = logf.Log.WithName("metering")

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, client client.Client) error {
	log.Info("Creating a new Metering", "Metering.Namespace", cr.Namespace, "Metering.Name", cr.Name)

	//Create image pull secret for metering if it is in ICP
	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		err := checkAndCreateSecretForMetering(instance, cr, client)
		if err != nil {
			return err
		}
	}

	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE Metering CR")
		return err
	}

	// Adding Finalizer to Instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, foundCR *multicloudv1beta1.Metering, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE Tiller CR")
		return err
	}

	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		err = checkAndUpdateSecretForMetering(instance, cr, client)
		if err != nil {
			log.Error(err, "Fail to UPDATE image pull SECRET for metering")
			return err
		}
	}

	// Adding Finalizer to Instance if Finalizer does not exist
	// NOTE: This is to handle requeue due to failed instance update during creation
	for _, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			return nil
		}
	}
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func delete(instance *multicloudv1beta1.Endpoint, foundCR *multicloudv1beta1.Metering, client client.Client) error {
	err := client.Delete(context.TODO(), foundCR)
	if err != nil {
		return err
	}

	// delete the secret for metering
	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		err = checkAndDeleteSecretForMetering(instance, foundCR, client)
		if err != nil {
			log.Error(err, "Fail to DELETE image pull SECRET for metering")
			return err
		}
	}
	return err
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}

func checkAndCreateSecretForMetering(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, client client.Client) error {
	if len(cr.Spec.ImagePullSecrets) == 0 || cr.Spec.ImagePullSecrets[0] == "" {
		log.Info("There is no secret name for metering, cannot operate secret creation.")
		return nil
	}

	secretNameForICP := cr.Spec.ImagePullSecrets[0]
	foundSecretForICP := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: secretNameForICP, Namespace: cr.Namespace}, foundSecretForICP)
	if err != nil {
		if errors.IsNotFound(err) {
			// copy the image pull secret from instance's namespace and create a new one in namespace Kube-system.
			foundSecretInInstanceNameSpace := &corev1.Secret{}
			err := client.Get(context.TODO(), types.NamespacedName{Name: secretNameForICP, Namespace: instance.Namespace}, foundSecretInInstanceNameSpace)
			if err != nil {
				if errors.IsNotFound(err) {
					log.Info("Cannot find the secret", "Secret.Namespace", instance.Namespace, "Secret.Name", secretNameForICP)
					return nil
				}
				log.Info("Unexpect Err.")
				return err
			}

			secretToCreate := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretNameForICP,
					Namespace: cr.Namespace,
				},
				Data: foundSecretInInstanceNameSpace.Data,
				Type: corev1.SecretTypeDockerConfigJson,
			}

			log.Info("Create Secret: ", "Secret.Namespace", cr.Namespace, "Secret.Name", secretNameForICP)
			err = client.Create(context.TODO(), secretToCreate)
			if err != nil {
				log.Info("Failed to create Secret", "Secret.Namespace", cr.Namespace, "Secret.Name", secretNameForICP)
			}
		} else {
			log.Info("Unexpect Err.")
		}
		return err
	}
	return nil
}

func checkAndDeleteSecretForMetering(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, client client.Client) error {
	if len(cr.Spec.ImagePullSecrets) < 1 || cr.Spec.ImagePullSecrets[0] == "" {
		log.Info("There is no secret name for metering, cannot operate secret deletion.")
		return nil
	}
	secretNameForICP := cr.Spec.ImagePullSecrets[0]
	foundSecretForICP := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: secretNameForICP, Namespace: cr.Namespace}, foundSecretForICP)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("There is no secret for metering, no need to operate secret deletion.")
			return nil
		}
		log.Info("Unexpect ERROR")
		return err
	}

	err = client.Delete(context.TODO(), foundSecretForICP)
	if err != nil {
		log.Info("Failed to DELETE Secret", "Secret.Namespace", cr.Namespace, "Secret.Name", secretNameForICP)
		return err
	}
	return nil
}

func checkAndUpdateSecretForMetering(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Metering, client client.Client) error {
	if len(cr.Spec.ImagePullSecrets) < 1 || cr.Spec.ImagePullSecrets[0] == "" {
		log.Info("No Image Pull Secret For Metering")
		return nil
	}
	pullSecretNameForMetering := cr.Spec.ImagePullSecrets[0]

	if instance.Spec.ImagePullSecret == "" {
		log.Info("No Image Pull Secret For Endpoint")
		return nil
	}
	pullSecretNameForEndpoint := instance.Spec.ImagePullSecret

	pullSecretForEndpoint := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: pullSecretNameForEndpoint, Namespace: instance.Namespace}, pullSecretForEndpoint)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("There is no image pull secret for Endpoint")
			return nil
		}
		log.Info("Unexpect ERROR")
		return err
	}

	pullSecretForMetering := &corev1.Secret{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: pullSecretNameForMetering, Namespace: cr.Namespace}, pullSecretForMetering)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("There is no image pull secret specific for metering.")
			//create the secret
			secretToCreate := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pullSecretForEndpoint.Name,
					Namespace: cr.Namespace,
				},
				Data: pullSecretForEndpoint.Data,
				Type: corev1.SecretTypeDockerConfigJson,
			}

			log.Info("Create Secret: ", "Secret.Namespace", cr.Namespace, "Secret.Name", pullSecretForEndpoint.Name)
			err = client.Create(context.TODO(), secretToCreate)
			if err != nil {
				log.Info("Failed to create Secret", "Secret.Namespace", cr.Namespace, "Secret.Name", pullSecretForEndpoint.Name)
				return nil
			}
		}
		log.Info("Unexpect ERROR")
		return err
	}

	doUpdate := false
	if !bytes.Equal(pullSecretForMetering.Data[".dockerconfigjson"], pullSecretForEndpoint.Data[".dockerconfigjson"]) {
		pullSecretForMetering.Data[".dockerconfigjson"] = pullSecretForEndpoint.Data[".dockerconfigjson"]
		doUpdate = true
	}

	if pullSecretForMetering.Type != pullSecretForEndpoint.Type {
		pullSecretForMetering.Type = pullSecretForEndpoint.Type
	}

	if doUpdate {
		err = client.Update(context.TODO(), pullSecretForMetering)
		if err != nil {
			log.Info("Failed to update Metering Secret.")
			return err
		}
	}

	return nil
}
