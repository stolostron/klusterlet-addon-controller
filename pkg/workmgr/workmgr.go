//Package v1alpha1 Defines the API to support Multicluster Endpoints (klusterlets).
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
//

package workmgr

import (
	"context"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller"

	corev1 "k8s.io/api/core/v1"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"

	openshiftroutev1 "github.com/openshift/api/route/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("workmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	workmgrCR := newWorkManagerCR(instance)
	err := controllerutil.SetControllerReference(instance, workmgrCR, scheme)
	if err != nil {
		return err
	}

	foundWorkManager := &klusterletv1alpha1.WorkManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: workmgrCR.Name, Namespace: workmgrCR.Namespace}, foundWorkManager)
	if err != nil && errors.IsNotFound(err) {
		workmgrCR.Spec.TillerIntegration = newWorkManagerTillerIntegration(instance, client)
		workmgrCR.Spec.PrometheusIntegration = newWorkManagerPrometheusIntegration(instance, client)
		workmgrCR.Spec.Service = newWorkManagerServiceConfig()
		workmgrCR.Spec.Ingress = newWorkManagerIngressConfig(client)

		log.Info("Creating a new WorkManager", "WorkManager.Namespace", workmgrCR.Namespace, "WorkManager.Name", workmgrCR.Name)
		err = client.Create(context.TODO(), workmgrCR)
		if err != nil {
			return err
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
				Service:        "prometheus-k8s/openshift-monitoring",
				Secret:         "kube-system/monitoring-monitoring-client-certs",
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

func newWorkManagerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.WorkManager {
	labels := map[string]string{
		"app": cr.Name,
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

			WorkManagerConfig: klusterletv1alpha1.WorkManagerConfig{
				Enabled: true,
				Image: image.Image{
					Repository: "ibmcom/mcm-klusterlet",
					Tag:        "3.2.0",
					PullPolicy: "IfNotPresent",
				},
			},

			DeployableConfig: klusterletv1alpha1.DeployableConfig{
				Enabled: true,
				Image: image.Image{
					Repository: "ibmcom/deployable",
					Tag:        "3.2.0",
					PullPolicy: "IfNotPresent",
				},
			},
		},
	}
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
