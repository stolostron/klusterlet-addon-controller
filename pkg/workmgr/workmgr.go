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
		// KlusterletOperator deployed Tiller
		foundTillerService := &corev1.Service{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: cr.Name + "-tiller", Namespace: cr.Namespace}, foundTillerService)
		if err == nil {
			tillerServiceHostname := foundTillerService.Name + "." + foundTillerService.Namespace
			var tillerServicePort int32

			for _, port := range foundTillerService.Spec.Ports {
				if port.Name == "grpc" && port.Protocol == "TCP" {
					tillerServicePort = port.Port
				}
			}

			return klusterletv1alpha1.WorkManagerTillerIntegration{
				Enabled:       true,
				Endpoint:      tillerServiceHostname + ":" + strconv.FormatInt(int64(tillerServicePort), 10),
				CertIssuer:    cr.Name + "-tiller",
				AutoGenSecret: true,
				User:          "admin",
			}
		}
		//TODO: ICP Tiller
		//NOTE: we can actually detect the default admin user by decoding the tiller server cert...

		log.Info("Unable to locate TillerService")
	}

	return klusterletv1alpha1.WorkManagerTillerIntegration{
		Enabled: false,
	}
}

func newWorkManagerPrometheusIntegration(cr *klusterletv1alpha1.KlusterletService, client client.Client) klusterletv1alpha1.WorkManagerPrometheusIntegration {
	if cr.Spec.PrometheusIntegration.Enabled {
		//OpenShift Prometheus Service
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
		//TODO: ICP Prometheus Service
		//TODO: KlusterletOperator deployed Prometheus
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
