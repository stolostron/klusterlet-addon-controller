// Package v1beta1 of appmgr provides a reconciler for the ApplicationManager
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	helmcrd "github.ibm.com/IBMMulticloudPlatform/helm-crd/pkg/apis/helm.bitnami.com/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
)

var log = logf.Log.WithName("appmgr")

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ApplicationManager, client client.Client) error {
	log.Info("Creating a new ApplicationManager", "ApplicationManager.Namespace", cr.Namespace, "ApplicationManager.Name", cr.Name)

	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE ApplicationManager CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ApplicationManager, foundCR *multicloudv1beta1.ApplicationManager, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE ApplicationManager CR")
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

func delete(foundCR *multicloudv1beta1.ApplicationManager, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ApplicationManager, client client.Client) error {
	foundHelmCRD := &crdv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "helmreleases.helm.bitnami.com", Namespace: ""}, foundHelmCRD)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("HelmCRD not found, skipping delete")
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		err = cleanUpHelmCRs(client)
		if err != nil {
			log.Error(err, "Failed to clean up Helm CRs")
			return err
		}
	}

	err = cleanUpSecret(instance, client, cr)
	if err != nil {
		return err
	}
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}

func cleanUpHelmCRs(c client.Client) error {
	log.Info("Cleaning up Helm CRs")
	helmCRList := &helmcrd.HelmReleaseList{}
	err := c.List(context.TODO(), &client.ListOptions{}, helmCRList)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Could not find HelmCR list")
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		for _, helmCR := range helmCRList.Items {
			log.Info("Removing finalizer for CR " + helmCR.ObjectMeta.Name + " in Namespace " + helmCR.ObjectMeta.Namespace)
			helmCR.ObjectMeta.Finalizers = []string{}
			err = c.Update(context.TODO(), &helmCR)
			if err != nil {
				log.Error(err, "Failed to remove finalizer for CR "+helmCR.ObjectMeta.Name)
				return err
			}
			log.Info("Deleting CR " + helmCR.ObjectMeta.Name + " in Namespace " + helmCR.ObjectMeta.Namespace)
			err = c.Delete(context.TODO(), &helmCR)
			if err != nil {
				log.Error(err, "Failed to DELETE CR "+helmCR.ObjectMeta.Name)
				return err
			}
		}
	}

	return nil
}
