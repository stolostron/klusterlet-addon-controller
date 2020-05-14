// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of ciscontroller provides a reconciler for the search collector
package v1beta1

import (
	"context"

	"github.com/open-cluster-management/endpoint-operator/pkg/inspect"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	klusterletv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1beta1"
)

var log = logf.Log.WithName("ciscontroller")

// Reconcile reconciles the search collector
func Reconcile(instance *klusterletv1beta1.Klusterlet, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Klusterlet.Namespace", instance.Namespace, "EndpoKlusterletint.Name", instance.Name)
	reqLogger.Info("Reconciling CISController")

	// Deployed on hub
	if inspect.DeployedOnHub(client) {
		log.Info("Found clusterstatus.mcm.ibm.com, this is a hub cluster, skip CISController Reconcile.")
		return false, nil
	}

	// Not deployed on hub
	cisControllerCR, err := newCISControllerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired CISController CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, cisControllerCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundCISControllerCR := &klusterletv1beta1.CISController{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: cisControllerCR.Name, Namespace: cisControllerCR.Namespace}, foundCISControllerCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("CISController DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.CISControllerConfig.Enabled {
					log.V(5).Info("CISController ENABLED")
					err := create(instance, cisControllerCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE CISController CR")
						return false, err
					}
				} else {
					log.V(5).Info("CISController DISABLED")
					err := finalize(instance, cisControllerCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE CISController CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err := finalize(instance, cisControllerCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE CISController CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("CISController CR DOES exist")
		if foundCISControllerCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("CISController IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.CISControllerConfig.Enabled {
				log.V(5).Info("instance IS NOT in deletion state and Search Collector is ENABLED")
				err = update(instance, cisControllerCR, foundCISControllerCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE CISController CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or Search Collector is DISABLED")
				err := delete(foundCISControllerCR, client)
				if err != nil {
					log.Error(err, "fail to DELETE CISController CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for CISController")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for CISController")
			return true, err
		}
	}

	reqLogger.Info("Successfully Reconciled CISController")
	return false, nil
}

// TODO(liuhao): the following method need to be refactored as instance method of CISController struct
func newCISControllerCR(instance *klusterletv1beta1.Klusterlet, client client.Client) (*klusterletv1beta1.CISController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := klusterletv1beta1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 5),
	}

	imageKey, imageRepository, err := instance.GetImage("cis-controller-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-controller")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-crawler")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-crawler")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-drishti")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-drishti")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-minio")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-minio")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-minio-cleaner")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-minio-cleaner")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	return &klusterletv1beta1.CISController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-cisctrl",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1beta1.CISControllerSpec{
			FullNameOverride:  instance.Name + "-cisctrl",
			ClusterName:       instance.Spec.ClusterName,
			ClusterNamespace:  instance.Spec.ClusterNamespace,
			ConnectionManager: instance.Name + "-connmgr",
			GlobalValues:      gv,
			IsOpenShift:       inspect.Info.KubeVendor == inspect.KubeVendorOpenShift,
		},
	}, err
}

func create(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.CISController, client client.Client) error {
	// Create the CR and add the Finalizer to the instance
	log.Info("Creating a new CISController", "CISController.Namespace", cr.Namespace, "CISController.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE CISController CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.CISController, foundCR *klusterletv1beta1.CISController, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "fail to UPDATE CISController CR")
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

func delete(foundCR *klusterletv1beta1.CISController, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *klusterletv1beta1.Klusterlet, cisControllerCR *klusterletv1beta1.CISController, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cisControllerCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
