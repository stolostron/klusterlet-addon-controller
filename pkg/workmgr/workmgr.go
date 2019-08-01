// Package workmgr ....
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package workmgr

import (
	"context"

	openshiftroutev1 "github.com/openshift/api/route/v1"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("workmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
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
			if instance.GetDeletionTimestamp() == nil {
				log.Info("Creating a new WorkManager", "WorkManager.Namespace", workMgrCR.Namespace, "WorkManager.Name", workMgrCR.Name)

				err := client.Create(context.TODO(), workMgrCR)
				if err != nil {
					log.Error(err, "Fail to CREATE WorkManager CR")
					return err
				}

				// Adding Finalizer to KlusterletService instance
				instance.Finalizers = append(instance.Finalizers, workMgrCR.Name)
			} else {
				// Cleanup Secrets
				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == workMgrCR.Name {
						secretsToDeletes := []string{
							workMgrCR.Name + "-tiller-client-certs",
						}

						for _, secretToDelete := range secretsToDeletes {
							foundSecretToDelete := &corev1.Secret{}
							err = client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: workMgrCR.Namespace}, foundSecretToDelete)
							if err == nil {
								err = client.Delete(context.TODO(), foundSecretToDelete)
								if err != nil {
									log.Error(err, "Fail to DELETE WorkManager Secret", "Secret.Name", secretToDelete)
									return err
								}
							}
						}
						instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
						break
					}
				}

			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		if foundWorkMgrCR.GetDeletionTimestamp() == nil {
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				foundWorkMgrCR.Spec = workMgrCR.Spec
				err = client.Update(context.TODO(), foundWorkMgrCR)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE WorkManager CR")
					return err
				}
			} else {
				// KlusterletService in deletion state
				err = client.Delete(context.TODO(), foundWorkMgrCR)
				if err != nil {
					log.Error(err, "Fail to DELETE WorkManager CR")
					return err
				}
			}
		}
	}

	return nil
}

func newWorkManagerTillerIntegration(cr *klusterletv1alpha1.KlusterletService, client client.Client) klusterletv1alpha1.WorkManagerTillerIntegration {
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

func newWorkManagerPrometheusIntegration(cr *klusterletv1alpha1.KlusterletService, client client.Client) klusterletv1alpha1.WorkManagerPrometheusIntegration {
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

		//TODO: KlusterletOperator deployed Prometheus
		return klusterletv1alpha1.WorkManagerPrometheusIntegration{
			Enabled: false,
		}
	}

	return klusterletv1alpha1.WorkManagerPrometheusIntegration{
		Enabled: false,
	}
}

func newWorkManagerCR(cr *klusterletv1alpha1.KlusterletService, client client.Client) (*klusterletv1alpha1.WorkManager, error) {
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
	//TODO: IKS EKS GKE AKS

	// Other
	return klusterletv1alpha1.WorkManagerService{
		ServiceType: "ClusterIP",
	}
}

func newWorkManagerIngressConfig(c client.Client) klusterletv1alpha1.WorkManagerIngress {
	// OpenShift
	routeList := &openshiftroutev1.RouteList{}
	err := c.List(context.TODO(), &client.ListOptions{}, routeList)
	if err == nil {
		return klusterletv1alpha1.WorkManagerIngress{
			IngressType: "Route",
			//TODO: user specified hostname and port override
		}
	}

	// ICP Nginx Ingress
	foundICPIngressDaemonSet := &extensionsv1beta1.DaemonSet{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: "nginx-ingress-controller", Namespace: "kube-system"}, foundICPIngressDaemonSet)
	if err == nil {
		return klusterletv1alpha1.WorkManagerIngress{
			IngressType: "Ingress",
			//TODO: user specified hostname and port override
		}
	}

	// Other
	return klusterletv1alpha1.WorkManagerIngress{
		IngressType: "None",
	}
}
