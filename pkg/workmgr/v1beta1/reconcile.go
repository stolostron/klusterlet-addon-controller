//Package v1beta1 of workmgr Defines the Reconciliation logic and required setup for WorkManager CR.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	tiller "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("workmgr")

// TODO(liuhao): switch from klusterletv1alpha1 to multicloudv1beta1 for the WorkManager related structs

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Tiller")

	workMgrCR, err := newWorkManagerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired WorkManager CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, workMgrCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return err
	}

	foundWorkMgrCR := &klusterletv1alpha1.WorkManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: workMgrCR.Name, Namespace: workMgrCR.Namespace}, foundWorkMgrCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("WorkManager CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("Instance IS NOT in deletion state")
				err := create(instance, workMgrCR, client)
				if err != nil {
					log.Error(err, "fail to CREATE WorkManager CR")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				err := finalize(instance, workMgrCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE WorkManager CR")
					return err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		log.V(5).Info("WorkManager CR DOES exist")
		if foundWorkMgrCR.GetDeletionTimestamp() == nil {
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("WorkManager CR IS NOT in deletion state")
				err := update(instance, workMgrCR, foundWorkMgrCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE WorkManager CR")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				if foundWorkMgrCR.GetDeletionTimestamp() == nil {
					err := delete(foundWorkMgrCR, client)
					if err != nil {
						log.Error(err, "Fail to DELETE WorkManager CR")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled WorkManager")
	return nil
}

func newWorkManagerTillerIntegration(cr *multicloudv1beta1.Endpoint, client client.Client) klusterletv1alpha1.WorkManagerTillerIntegration {
	if cr.Spec.TillerIntegration.Enabled {
		// ICP Tiller
		icpTillerServiceEndpoint := tiller.GetICPTillerServiceEndpoint(client)
		if icpTillerServiceEndpoint != "" {
			return klusterletv1alpha1.WorkManagerTillerIntegration{
				Enabled:           true,
				HelmReleasePrefix: "md",
				Endpoint:          icpTillerServiceEndpoint,
				CertIssuer:        "icp-ca-issuer",
				AutoGenSecret:     true,
				User:              tiller.GetICPTillerDefaultAdminUser(client),
			}
		}

		// KlusterletOperator deployed Tiller
		return klusterletv1alpha1.WorkManagerTillerIntegration{
			Enabled:           true,
			HelmReleasePrefix: "md",
			Endpoint:          cr.Name + "-tiller" + ":44134",
			CertIssuer:        cr.Name + "-tiller",
			AutoGenSecret:     true,
			User:              cr.Name + "-admin",
		}
	}

	return klusterletv1alpha1.WorkManagerTillerIntegration{
		Enabled: false,
	}
}

func newWorkManagerPrometheusIntegration(cr *multicloudv1beta1.Endpoint, client client.Client) klusterletv1alpha1.WorkManagerPrometheusIntegration {
	if cr.Spec.PrometheusIntegration.Enabled {
		// OpenShift Prometheus Service
		foundOpenshiftPrometheusService := &corev1.Service{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: "prometheus-k8s", Namespace: "openshift-monitoring"}, foundOpenshiftPrometheusService)
		if err == nil { //found OpenShift Prometheus
			return klusterletv1alpha1.WorkManagerPrometheusIntegration{
				Enabled:        true,
				Service:        "openshift-monitoring/prometheus-k8s",
				Secret:         "",
				UseBearerToken: true,
			}
		}

		// ICP Prometheus Service
		foundICPPrometheusService := &corev1.Service{}
		err = client.Get(context.TODO(), types.NamespacedName{Name: "prometheus-k8s", Namespace: "openshift-monitoring"}, foundICPPrometheusService)
		if err == nil { //found ICP Prometheus
			return klusterletv1alpha1.WorkManagerPrometheusIntegration{
				Enabled:        true,
				Service:        "kube-system/monitoring-prometheus",
				Secret:         "kube-system/monitoring-monitoring-client-certs",
				UseBearerToken: false,
			}
		}

		//TODO(liuhao): KlusterletOperator deployed Prometheus
		return klusterletv1alpha1.WorkManagerPrometheusIntegration{
			Enabled: false,
		}
	}

	return klusterletv1alpha1.WorkManagerPrometheusIntegration{
		Enabled: false,
	}
}

func newWorkManagerCR(cr *multicloudv1beta1.Endpoint, client client.Client) (*klusterletv1alpha1.WorkManager, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	workMgrImage, err := cr.GetImage("work-manager")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "work-manager")
		return nil, err
	}

	deployableImage, err := cr.GetImage("deployable")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "deployable")
		return nil, err
	}

	return &klusterletv1alpha1.WorkManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-workmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.WorkManagerSpec{
			FullNameOverride: cr.Name + "-workmgr",

			ClusterName:      cr.Spec.ClusterName,
			ClusterNamespace: cr.Spec.ClusterNamespace,
			ClusterLabels:    cr.Spec.ClusterLabels,

			ConnectionManager: cr.Name + "-connmgr",

			TillerIntegration:     newWorkManagerTillerIntegration(cr, client),
			PrometheusIntegration: newWorkManagerPrometheusIntegration(cr, client),
			Service:               newWorkManagerServiceConfig(),
			Ingress:               newWorkManagerIngressConfig(client),

			WorkManagerConfig: klusterletv1alpha1.WorkManagerConfig{
				Enabled: true,
				Image:   workMgrImage,
			},

			DeployableConfig: klusterletv1alpha1.DeployableConfig{
				Enabled: true,
				Image:   deployableImage,
			},

			ImagePullSecret: cr.Spec.ImagePullSecret,
		},
	}, nil
}

func newWorkManagerServiceConfig() klusterletv1alpha1.WorkManagerService {
	workManagerService := klusterletv1alpha1.WorkManagerService{}

	switch kubeVendor := inspect.Info.KubeVendor; kubeVendor {
	case inspect.KubeVendorAKS:
		fallthrough
	case inspect.KubeVendorEKS:
		fallthrough
	case inspect.KubeVendorGKE:
		fallthrough
	case inspect.KubeVendorIKS:
		workManagerService.ServiceType = "NodePort"
	default:
		workManagerService.ServiceType = "ClusterIP"
	}

	return workManagerService
}

func newWorkManagerIngressConfig(c client.Client) klusterletv1alpha1.WorkManagerIngress {
	workManagerIngress := klusterletv1alpha1.WorkManagerIngress{}

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

func create(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.WorkManager, client client.Client) error {
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

func update(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.WorkManager, foundCR *klusterletv1alpha1.WorkManager, client client.Client) error {
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

func delete(foundCR *klusterletv1alpha1.WorkManager, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *klusterletv1alpha1.WorkManager, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Deletes Secrets
			secretsToDeletes := []string{
				cr.Name + "-tiller-client-certs",
			}

			for _, secretToDelete := range secretsToDeletes {
				foundSecretToDelete := &corev1.Secret{}
				err := client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: cr.Namespace}, foundSecretToDelete)
				if err == nil {
					err := client.Delete(context.TODO(), foundSecretToDelete)
					if err != nil {
						log.Error(err, "Fail to DELETE WorkManager Secret", "Secret.Name", secretToDelete)
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
