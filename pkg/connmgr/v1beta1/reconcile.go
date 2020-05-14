// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of connmgr Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	klusterletv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1beta1"
)

var log = logf.Log.WithName("connmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1beta1.Klusterlet, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Klusterlet.Namespace", instance.Namespace, "Klusterlet.Name", instance.Name)
	reqLogger.Info("Reconciling ConnectionManager")

	connMgrCR, err := newConnectionManagerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired ConnectionManager CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, connMgrCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundConnMgrCR := &klusterletv1beta1.ConnectionManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: connMgrCR.Name, Namespace: connMgrCR.Namespace}, foundConnMgrCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("ConnectionManager CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				err := create(instance, connMgrCR, client)
				if err != nil {
					log.Error(err, "fail to CREATE ConnectionManager CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err := finalize(instance, connMgrCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE ConnectionManager CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("ConnectionManager CR DOES exist")
		if foundConnMgrCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("ConnectionManager CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				err := update(instance, connMgrCR, foundConnMgrCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE ConnectionManager CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				if foundConnMgrCR.GetDeletionTimestamp() == nil {
					err = delete(foundConnMgrCR, client)
					if err != nil {
						log.Error(err, "Fail to DELETE ConnectionManager CR")
						return false, err
					}
					reqLogger.Info("Requeueing Reconcile for ConnectionManager")
					return true, err
				}
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for ConnectionManager")
			return true, err
		}
	}

	reqLogger.Info("Successfully Reconciled ConnectionManager")
	return false, nil
}

func newConnectionManagerCR(instance *klusterletv1beta1.Klusterlet) (*klusterletv1beta1.ConnectionManager, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := klusterletv1beta1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageKey, imageRepository, err := instance.GetImage("connection-manager")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "connection-manager")
		return nil, err
	}
	gv.ImageOverrides[imageKey] = imageRepository

	// if BootStrapConfig is empty, adds default value

	_, ok := instance.Spec.BootStrapConfig["hubSecret"]
	if !ok {
		if instance.Spec.BootStrapConfig == nil {
			instance.Spec.BootStrapConfig = make(map[string]string)
		}
		instance.Spec.BootStrapConfig["hubSecret"] = instance.Namespace + "/klusterlet-bootstrap"
	}

	return &klusterletv1beta1.ConnectionManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-connmgr",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1beta1.ConnectionManagerSpec{
			ClusterName:      instance.Spec.ClusterName,
			ClusterNamespace: instance.Spec.ClusterNamespace,
			BootStrapConfig:  instance.Spec.BootStrapConfig,
			FullNameOverride: instance.Name + "-connmgr",
			GlobalValues:     gv,
		},
	}, nil
}

func create(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.ConnectionManager, client client.Client) error {
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

func update(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.ConnectionManager, foundCR *klusterletv1beta1.ConnectionManager, client client.Client) error {
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

func delete(foundCR *klusterletv1beta1.ConnectionManager, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.ConnectionManager, client client.Client) error {
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
