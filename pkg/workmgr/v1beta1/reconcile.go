// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of workmgr Defines the Reconciliation logic and required setup for WorkManager CR.
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

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
	"github.com/open-cluster-management/endpoint-operator/pkg/inspect"
)

var log = logf.Log.WithName("workmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling WorkManager")

	workMgrCR, err := newWorkManagerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired WorkManager CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, workMgrCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return false, err
	}

	foundWorkMgrCR := &multicloudv1beta1.WorkManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: workMgrCR.Name, Namespace: workMgrCR.Namespace}, foundWorkMgrCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("WorkManager CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				err = create(instance, workMgrCR, client)
				if err != nil {
					log.Error(err, "fail to CREATE WorkManager CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err = finalize(instance, workMgrCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE WorkManager CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("WorkManager CR DOES exist")
		if foundWorkMgrCR.GetDeletionTimestamp() == nil {
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("WorkManager CR IS NOT in deletion state")
				if err = update(instance, workMgrCR, foundWorkMgrCR, client); err != nil {
					log.Error(err, "fail to UPDATE WorkManager CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err = delete(foundWorkMgrCR, client)
				if err != nil {
					log.Error(err, "Fail to DELETE WorkManager CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for WorkManager")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for WorkManager")
			return true, err
		}
	}
	reqLogger.Info("Successfully Reconciled WorkManager")
	return false, nil
}

func newWorkManagerCR(cr *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.WorkManager, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	workMgrImage, err := cr.GetImage("work-manager")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "work-manager")
		return nil, err
	}

	clusterLabels := cr.Spec.ClusterLabels
	if clusterLabels != nil {
		if clusterLabels["cloud"] == "auto-detect" {
			clusterLabels["cloud"] = string(inspect.Info.CloudVendor)
		}

		if clusterLabels["vendor"] == "auto-detect" {
			clusterLabels["vendor"] = string(inspect.Info.KubeVendor)
		}
	}

	return &multicloudv1beta1.WorkManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-workmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.WorkManagerSpec{
			FullNameOverride: cr.Name + "-workmgr",

			ClusterName:      cr.Spec.ClusterName,
			ClusterNamespace: cr.Spec.ClusterNamespace,
			ClusterLabels:    clusterLabels,

			ConnectionManager: cr.Name + "-connmgr",

			Service: newWorkManagerServiceConfig(),
			Ingress: newWorkManagerIngressConfig(client),

			WorkManagerConfig: multicloudv1beta1.WorkManagerConfig{
				Enabled: true,
				Image:   workMgrImage,
			},

			ImagePullSecret: cr.Spec.ImagePullSecret,
		},
	}, nil
}

func newWorkManagerServiceConfig() multicloudv1beta1.WorkManagerService {
	workManagerService := multicloudv1beta1.WorkManagerService{}

	switch kubeVendor := inspect.Info.KubeVendor; kubeVendor {
	case inspect.KubeVendorAKS:
		fallthrough
	case inspect.KubeVendorEKS:
		fallthrough
	case inspect.KubeVendorGKE:
		fallthrough
	case inspect.KubeVendorIKS:
		workManagerService.ServiceType = "LoadBalancer"
	default:
		workManagerService.ServiceType = "ClusterIP"
	}

	return workManagerService
}

func newWorkManagerIngressConfig(c client.Client) multicloudv1beta1.WorkManagerIngress {
	workManagerIngress := multicloudv1beta1.WorkManagerIngress{}

	switch kubeVendor := inspect.Info.KubeVendor; kubeVendor {
	case inspect.KubeVendorOpenShift:
		workManagerIngress.IngressType = "Route"
	case inspect.KubeVendorICP:
		workManagerIngress.IngressType = "Ingress"
	default:
		workManagerIngress.IngressType = "None"
	}

	//TODO(liuhao): user specified hostname and port override
	return workManagerIngress
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.WorkManager, client client.Client) error {
	log.Info("Creating a new WorkManager", "WorkManager.Namespace", cr.Namespace, "WorkManager.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE WorkManager CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.WorkManager, foundCR *multicloudv1beta1.WorkManager, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE WorkManager CR")
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

func delete(foundCR *multicloudv1beta1.WorkManager, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.WorkManager, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
