package workmgr

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("workmgr")

func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	workmgrCR := newWorkManagerCR(instance, client)
	err := controllerutil.SetControllerReference(instance, workmgrCR, scheme)
	if err != nil {
		return err
	}

	foundWorkManager := &klusterletv1alpha1.WorkManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: workmgrCR.Name, Namespace: workmgrCR.Namespace}, foundWorkManager)
	if err != nil && errors.IsNotFound(err) {
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
		foundICPTillerService := &corev1.Service{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, foundICPTillerService)
		if err == nil {
			tillerServiceHostname := foundICPTillerService.Name + "." + foundICPTillerService.Namespace
			var tillerServicePort int32

			for _, port := range foundICPTillerService.Spec.Ports {
				if port.Name == "grpc" && port.Protocol == "TCP" {
					tillerServicePort = port.Port
				}
			}

			return klusterletv1alpha1.WorkManagerTillerIntegration{
				Enabled:       true,
				Endpoint:      tillerServiceHostname + ":" + strconv.FormatInt(int64(tillerServicePort), 10),
				CertIssuer:    "icp-ca-issuer",
				AutoGenSecret: true,
				User:          getICPTillerDefaultAdminUser(client),
			}
		}

		// KlusterletOperator deployed Tiller
		return klusterletv1alpha1.WorkManagerTillerIntegration{
			Enabled:       true,
			Endpoint:      cr.Name + "-tiller" + ":44134",
			CertIssuer:    cr.Name + "-tiller",
			AutoGenSecret: true,
			User:          cr.Name + "admin",
		}
	}

	return klusterletv1alpha1.WorkManagerTillerIntegration{
		Enabled: false,
	}
}

func getICPTillerDefaultAdminUser(client client.Client) string {
	findICPTillerDeployment := &extensionsv1beta1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, findICPTillerDeployment)
	if err != nil {
		return "admin"
	}
	for _, container := range findICPTillerDeployment.Spec.Template.Spec.Containers {
		if container.Name == "tiller" {
			for _, env := range container.Env {
				if env.Name == "default_admin_user" {
					return env.Value
				}
			}
		}
	}
	return "admin"
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

func newWorkManagerCR(cr *klusterletv1alpha1.KlusterletService, client client.Client) *klusterletv1alpha1.WorkManager {
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
			Tiller:            cr.Name + "-tiller",

			TillerIntegration:     newWorkManagerTillerIntegration(cr, client),
			PrometheusIntegration: newWorkManagerPrometheusIntegration(cr, client),

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

			//TODO: non OpenShift
			Service: klusterletv1alpha1.WorkManagerService{
				ServiceType: "ClusterIP",
			},
			//TODO: non OpenShift
			Ingress: klusterletv1alpha1.WorkManagerIngress{
				IngressType: "Route",
			},
		},
	}
}
