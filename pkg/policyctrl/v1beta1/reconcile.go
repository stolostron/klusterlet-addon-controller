// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of policyctrl Defines the Reconciliation logic and required setup for PolicyController.
package v1beta1

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	klusterletv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1beta1"
	"github.com/open-cluster-management/endpoint-operator/pkg/inspect"
)

var log = logf.Log.WithName("policyctrl")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1beta1.Klusterlet, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Klusterlet.Namespace", instance.Namespace, "Klusterlet.Name", instance.Name)
	reqLogger.Info("Reconciling Policy Controller")

	policyCtrlCR, err := newPolicyControllerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired PolicyController CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, policyCtrlCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundPolicyCtrlCR := &klusterletv1beta1.PolicyController{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: policyCtrlCR.Name, Namespace: policyCtrlCR.Namespace}, foundPolicyCtrlCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("PolicyController CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.PolicyController.Enabled {
					log.V(5).Info("PolicyController ENABLED")
					err := create(instance, policyCtrlCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE PolicyController CR")
						return false, err
					}
				} else {
					log.V(5).Info("PolicyController DISABLED")
					finalize(instance, policyCtrlCR)
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				finalize(instance, policyCtrlCR)
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("PolicyController CR DOES exist")
		if foundPolicyCtrlCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("PolicyController CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.PolicyController.Enabled {
				log.V(5).Info("instance IS NOT in deletion state and PolicyController is ENABLED")
				err := update(instance, policyCtrlCR, foundPolicyCtrlCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE PolicyController CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or PolicyController is DISABLED")
				err := delete(foundPolicyCtrlCR, client)
				if err != nil {
					log.Error(err, "Fail to DELETE PolicyController CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for PolicyController")
				return true, nil
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for PolicyController")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled Policy Controller")
	return false, nil
}

func newPolicyControllerCR(instance *klusterletv1beta1.Klusterlet,
	client client.Client,
) (*klusterletv1beta1.PolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := klusterletv1beta1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageKey, imageRepository, err := instance.GetImage("policy-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "policy-controller")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	return &klusterletv1beta1.PolicyController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-policyctrl",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1beta1.PolicyControllerSpec{
			FullNameOverride:            instance.Name + "-policyctrl",
			ClusterName:                 instance.Spec.ClusterName,
			ClusterNamespace:            instance.Spec.ClusterNamespace,
			ConnectionManager:           instance.Name + "-connmgr",
			GlobalValues:                gv,
			DeployedOnHub:               inspect.DeployedOnHub(client),
			PostDeleteJobServiceAccount: instance.Name + "-component-operator",
		},
	}, nil
}

func create(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.PolicyController, c client.Client) error {
	log.Info("Creating a new PolicyController", "PolicyController.Namespace", cr.Namespace, "PolicyController.Name", cr.Name)
	err := c.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE PolicyController CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *klusterletv1beta1.Klusterlet, cr *klusterletv1beta1.PolicyController, foundCR *klusterletv1beta1.PolicyController, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "fail to UPDATE SearchCollector CR")
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

func delete(policyCR *klusterletv1beta1.PolicyController, c client.Client) error {
	return c.Delete(context.TODO(), policyCR)
}

func finalize(instance *klusterletv1beta1.Klusterlet, policyCtrlCR *klusterletv1beta1.PolicyController) {
	for i, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return
		}
	}
	log.V(5).Info("No Policy Controller Finalizer in the Klusterlet.")
}
