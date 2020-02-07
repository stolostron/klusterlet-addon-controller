// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

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

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
)

var log = logf.Log.WithName("policyctrl")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
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

	foundPolicyCtrlCR := &multicloudv1beta1.PolicyController{}
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

func newPolicyControllerCR(cr *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.PolicyController, error) {
	image, err := cr.GetImage("policy-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "policy-controller")
		return nil, err
	}

	labels := map[string]string{
		"app": cr.Name,
	}
	return &multicloudv1beta1.PolicyController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-policyctrl",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.PolicyControllerSpec{
			FullNameOverride:            cr.Name + "-policyctrl",
			ClusterName:                 cr.Spec.ClusterName,
			ClusterNamespace:            cr.Spec.ClusterNamespace,
			ConnectionManager:           cr.Name + "-connmgr",
			Image:                       image,
			ImagePullSecret:             cr.Spec.ImagePullSecret,
			DeployedOnHub:               inspect.DeployedOnHub(client),
			PostDeleteJobServiceAccount: cr.Name + "-component-operator",
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.PolicyController, c client.Client) error {
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

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.PolicyController, foundCR *multicloudv1beta1.PolicyController, client client.Client) error {
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

func delete(policyCR *multicloudv1beta1.PolicyController, c client.Client) error {
	return c.Delete(context.TODO(), policyCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, policyCtrlCR *multicloudv1beta1.PolicyController) {
	for i, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return
		}
	}
	log.V(5).Info("No Policy Controller Finalizer in the Endpoint.")
}
