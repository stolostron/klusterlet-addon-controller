// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of certpolicycontroller provides a reconciler for the search collector
package v1beta1

import (
	"context"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
	"github.com/open-cluster-management/endpoint-operator/pkg/inspect"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("certpolicycontroller")

// Reconcile reconciles the search collector
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling CertPolicyController")

	// Deployed on hub
	if inspect.DeployedOnHub(client) {
		log.Info("Found clusterstatus.mcm.ibm.com, this is a hub cluster, skip CertPolicyController Reconcile.")
		return false, nil
	}

	// Not deployed on hub
	certPolicyControllerCR, err := newCertPolicyControllerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired CertPolicyController CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, certPolicyControllerCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundCertPolicyControllerCR := &multicloudv1beta1.CertPolicyController{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: certPolicyControllerCR.Name, Namespace: certPolicyControllerCR.Namespace}, foundCertPolicyControllerCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("CertPolicyController DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.CertPolicyControllerConfig.Enabled {
					log.V(5).Info("CertPolicyController ENABLED")
					err := create(instance, certPolicyControllerCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE CertPolicyController CR")
						return false, err
					}
				} else {
					log.V(5).Info("CertPolicyController DISABLED")
					err := finalize(instance, certPolicyControllerCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE CertPolicyController CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err := finalize(instance, certPolicyControllerCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE CertPolicyController CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("CertPolicyController CR DOES exist")
		if foundCertPolicyControllerCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("CertPolicyController IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.CertPolicyControllerConfig.Enabled {
				log.V(5).Info("instance IS NOT in deletion state and Search Collector is ENABLED")
				err = update(instance, certPolicyControllerCR, foundCertPolicyControllerCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE CertPolicyController CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or Search Collector is DISABLED")
				err := delete(foundCertPolicyControllerCR, client)
				if err != nil {
					log.Error(err, "fail to DELETE CertPolicyController CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for CertPolicyController")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for CertPolicyController")
			return true, err
		}
	}

	reqLogger.Info("Successfully Reconciled CertPolicyController")
	return false, nil
}

// TODO(liuhao): the following method need to be refactored as instance method of CertPolicyController struct
func newCertPolicyControllerCR(instance *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.CertPolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	var imageShaDigests = make(map[string]string, 1)
	image, imageShaDigests, err := instance.GetImage("cert-policy-controller", imageShaDigests)
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cert-policy")
		return nil, err
	}

	return &multicloudv1beta1.CertPolicyController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-certpolicyctrl",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.CertPolicyControllerSpec{
			FullNameOverride:  instance.Name + "-certpolicyctrl",
			ClusterName:       instance.Spec.ClusterName,
			ClusterNamespace:  instance.Spec.ClusterNamespace,
			ConnectionManager: instance.Name + "-connmgr",
			Image:             image,
			ImageShaDigests:   imageShaDigests,
			ImagePullSecret:   instance.Spec.ImagePullSecret,
		},
	}, err
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.CertPolicyController, client client.Client) error {
	// Create the CR and add the Finalizer to the instance
	log.Info("Creating a new CertPolicyController", "CertPolicyController.Namespace", cr.Namespace, "CertPolicyController.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE CertPolicyController CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.CertPolicyController, foundCR *multicloudv1beta1.CertPolicyController, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "fail to UPDATE CertPolicyController CR")
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

func delete(foundCR *multicloudv1beta1.CertPolicyController, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, certPolicyControllerCR *multicloudv1beta1.CertPolicyController, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == certPolicyControllerCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
