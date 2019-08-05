//Package v1alpha1 of policyctrl Defines the Reconciliation logic and required setup for PolicyController.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1alpha1

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
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

var log = logf.Log.WithName("policyctrl")

// Reconcile the PolicyController Component
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	policyCtrlCR := newPolicyControllerCR(instance)
	err := controllerutil.SetControllerReference(instance, policyCtrlCR, scheme)
	if err != nil {
		log.Info("Failed to set the Reference for the PolicyController: ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
		return err
	}

	policyCtrlCR.Spec.DeployedOnHub = inspect.DeployedOnHub(client)

	foundPolicyCtrlCR := &klusterletv1alpha1.PolicyController{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: policyCtrlCR.Name, Namespace: policyCtrlCR.Namespace}, foundPolicyCtrlCR)

	if err != nil && !errors.IsNotFound(err) {
		log.Info("Unexpected error while GET PolicyController CR ", "PolicyController.Name:", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
		return err
	}
	isCRFound := err == nil

	if instance.DeletionTimestamp != nil {
		// Klusterlet: deleting
		if isCRFound {
			// Policy CR: Found
			err = deletePolicyController(instance, client, policyCtrlCR, foundPolicyCtrlCR)
			if err != nil {
				log.Info("Failed to delete the PolicyController ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
				return err
			}
			return nil
		}

		// Policy CR: Not Found
		err = cleanUpAfterPolicyCRDeletion(instance, client, policyCtrlCR)
		if err != nil {
			log.Info("Failed to clean up after  the PolicyController deleting ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
			return err
		}

		return nil
	}

	// Klusterlet: Not deleting
	if isCRFound {
		// Policy CR: Found
		if instance.Spec.PolicyController.Enabled {
			//TODO(diane): handle Update
			return nil
		}

		// Policy: Disabled
		err = deletePolicyController(instance, client, policyCtrlCR, foundPolicyCtrlCR)
		if err != nil {
			log.Info("Failed to delete the PolicyController: ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
			return err
		}

		return nil
	}

	// Klusterlet: Not deleting
	// Policy CR: Not found
	if instance.Spec.PolicyController.Enabled {
		// Policy Component: enabled
		// create the policy CR
		err = createPolicyController(instance, client, policyCtrlCR)
		if err != nil {
			log.Info("Failed to create the PolicyController: ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
			return err
		}
		instance.Finalizers = append(instance.Finalizers, policyCtrlCR.Name)
		return nil
	}

	// Klusterlet: Not deleting
	// Policy CR: Not found
	// Policy Component: Disable
	if policyFinalizerExists(instance, policyCtrlCR) {
		// Finalizer: Exists
		// Clean up and remove finalizer
		err = cleanUpAfterPolicyCRDeletion(instance, client, policyCtrlCR)
		if err != nil {
			log.Info("Failed to clean up for PolicyController: ", "PolicyController.Name: ", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
			return err
		}
		removePolicyFinalizer(instance, policyCtrlCR)
	}
	return nil
}

func createPolicyController(instance *klusterletv1alpha1.KlusterletService, c client.Client, policyCtrlCR *klusterletv1alpha1.PolicyController) error {
	finalexist := policyFinalizerExists(instance, policyCtrlCR)
	if finalexist {
		log.Info("Finalizer " + policyCtrlCR.Name + " is already existed. Exit the Creating Policy Controller CR Process")
		return nil
	}

	log.Info("Creating a new PolicyController ", "PolicyController.Name:", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
	return c.Create(context.TODO(), policyCtrlCR)
}

func deletePolicyController(instance *klusterletv1alpha1.KlusterletService, c client.Client, policyCtrlCR *klusterletv1alpha1.PolicyController, foundPolicyCtrlCR *klusterletv1alpha1.PolicyController) error {
	// Check if the finalizer is existed or not. If not, do nothing.
	if !policyFinalizerExists(instance, policyCtrlCR) {
		log.Info("Finalizer " + policyCtrlCR.Name + " is not existed. Exit the Deleting Policy Controller Process")
		return nil
	}

	// Check if there is a PolicyControllerCR in the cluster or not.
	var err error

	// Make sure the found Policy CR is NOT in the state of deleting.
	if foundPolicyCtrlCR.DeletionTimestamp == nil {
		log.Info("Deleting the PolicyController ", "PolicyController.Name:", policyCtrlCR.Name, ", PolicyController.Namespace:", policyCtrlCR.Namespace)
		err = c.Delete(context.TODO(), policyCtrlCR)
		if err != nil && errors.IsNotFound(err) {
			log.Info("No existing policy controller found to delete.")
			return nil
		}
	}

	return err
}

func newPolicyControllerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.PolicyController {
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
			//TODO(diane): add Image
		},
	}
}

func removePolicyFinalizer(instance *klusterletv1alpha1.KlusterletService, policyCtrlCR *klusterletv1alpha1.PolicyController) {
	for i, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			break
		}
	}
}

func cleanUpAfterPolicyCRDeletion(instance *klusterletv1alpha1.KlusterletService, c client.Client, policyCtrlCR *klusterletv1alpha1.PolicyController) error {
	if !policyFinalizerExists(instance, policyCtrlCR) {
		// Finalizer: Not exists
		log.Info("Finalizer " + policyCtrlCR.Name + " is not existed. No need to CleanUp and exit the CleanUp after Deletion Policy Controller Process")
		return nil
	}

	// Finalizer: Not exists
	jobname := policyCtrlCR.Name + "-post-delete"
	existedJob := &batchv1.Job{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: jobname, Namespace: policyCtrlCR.Namespace}, existedJob)

	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	// Job: Existed
	if err == nil {
		log.Info("The job for Job is existed. Deleting the existed Job")
		// Status: Not Successed - Not Finished
		if existedJob.Status.Succeeded == 0 {
			return nil
		}

		// Status: Successed
		err = c.Delete(context.TODO(), existedJob)
		if err != nil {
			log.Info("Failed to delete the existing job " + jobname)
		}
	}

	// Job: Not Existed
	// var ttl int32 = 1
	var activesecond int64 = 60
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyCtrlCR.Name + "-post-delete",
			Namespace: policyCtrlCR.Namespace,
			Labels: map[string]string{
				"app": instance.Name,
			},
		},

		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: &ttl,
			ActiveDeadlineSeconds: &activesecond,
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
	err = c.Create(context.TODO(), job)
	if err == nil {
		removePolicyFinalizer(instance, policyCtrlCR)
	}
	return err
}

func policyFinalizerExists(instance *klusterletv1alpha1.KlusterletService, policyCtrlCR *klusterletv1alpha1.PolicyController) bool {
	finalexist := false
	for _, finalizer := range instance.Finalizers {
		if finalizer == policyCtrlCR.Name {
			log.Info("Find the finalizer: " + finalizer)
			finalexist = true
			break
		}
	}
	return finalexist
}
