//Package v1beta1 of policyctrl Defines the Reconciliation logic and required setup for PolicyController.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

	batchv1 "k8s.io/api/batch/v1"
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

var log = logf.Log.WithName("policyctrl")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	requeue := false
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Policy Controller")

	policyCtrlCR, err := newPolicyControllerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired PolicyController CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, policyCtrlCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundPolicyCtrlCR := &klusterletv1alpha1.PolicyController{}
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
					requeue, err = finalize(instance, policyCtrlCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE PolicyController CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				requeue, err = finalize(instance, policyCtrlCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE PolicyController CR")
					return false, err
				}
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
					log.Error(err, "fail to DELETE PolicyController CR")
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
	return requeue, nil
}

func newPolicyControllerCR(cr *multicloudv1beta1.Endpoint) (*klusterletv1alpha1.PolicyController, error) {
	image, err := cr.GetImage("policy-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "policy-controller")
		return nil, err
	}

	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.PolicyController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-policyctrl",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.PolicyControllerSpec{
			FullNameOverride:  cr.Name + "-policyctrl",
			ClusterName:       cr.Spec.ClusterName,
			ClusterNamespace:  cr.Spec.ClusterNamespace,
			ConnectionManager: cr.Name + "-connmgr",
			Image:             image,
			ImagePullSecret:   cr.Spec.ImagePullSecret,
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.PolicyController, c client.Client) error {
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

func update(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.PolicyController, foundCR *klusterletv1alpha1.PolicyController, client client.Client) error {
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

func delete(policyCR *klusterletv1alpha1.PolicyController, c client.Client) error {
	return c.Delete(context.TODO(), policyCR)
}

func removePolicyFinalizer(instance *multicloudv1beta1.Endpoint, policyCtrlCR *klusterletv1alpha1.PolicyController) {
	for i, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			break
		}
	}
}

func finalize(instance *multicloudv1beta1.Endpoint, policyCtrlCR *klusterletv1alpha1.PolicyController, c client.Client) (bool, error) {
	if !policyFinalizerExists(instance, policyCtrlCR) {
		// Finalizer: Not exists
		log.V(5).Info("Finalizer:  " + policyCtrlCR.Name + " is not existed. No need to CleanUp and exit the CleanUp after Deletion Policy Controller Process")
		return false, nil
	}

	// Finalizer: Exists
	jobname := policyCtrlCR.Name + "-post-delete"
	existedJob := &batchv1.Job{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: jobname, Namespace: policyCtrlCR.Namespace}, existedJob)
	if err == nil {
		log.Info("The job for Job is existed. Deleting the existed Job")
		if existedJob.Status.Succeeded == 1 {
			// Status:  Successed
			err = c.Delete(context.TODO(), existedJob)
			if err != nil {
				log.Info("Failed to delete the existing job " + jobname)
			}
			removePolicyFinalizer(instance, policyCtrlCR)
			return false, nil
		}
		// Status:  Deleting, need to requeue and try to delete it in next loop.
		return true, nil
	}

	if errors.IsNotFound(err) {
		//image, err := instance.GetImage("policy-controller-post-deletion")
		// if err != nil {
		// 	log.Info("Failed to get image for Policy post deletion " + jobname)
		// }
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      policyCtrlCR.Name + "-post-delete",
				Namespace: policyCtrlCR.Namespace,
				Labels: map[string]string{
					"app": instance.Name,
				},
			},

			Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyCtrlCR.Name + "-post-delete",
						Labels: map[string]string{
							"name": policyCtrlCR.Name + "-post-delete",
						},
					},
					Spec: corev1.PodSpec{
						RestartPolicy: "Never",
						Containers: []corev1.Container{
							{
								//TODO(diane): use image_util to generate image information
								Name:            "policy-post-delete-job",
								Image:           "ibmcom/mcm-compliance:3.2.0",
								ImagePullPolicy: "IfNotPresent",
								Command:         []string{"uninstall-crd"},
								Args:            []string{"--removelist=compliances,policies,alerttargets"},
							},
						},
					},
				},
			},
		}
		return false, c.Create(context.TODO(), job)
	}

	return false, err
}

func policyFinalizerExists(instance *multicloudv1beta1.Endpoint, policyCtrlCR *klusterletv1alpha1.PolicyController) bool {
	finalexist := false
	for _, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			log.V(5).Info("Find the finalizer: " + finalizer)
			finalexist = true
			break
		}
	}
	return finalexist
}
