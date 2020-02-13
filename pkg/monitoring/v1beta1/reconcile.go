// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of monitoring provides a reconciler for the monitoring component
package v1beta1

import (
	"context"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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

var log = logf.Log.WithName("monitoring")

// Reconcile reconciles the monitoring
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Monitoring")

	// Openshift Monitoring
	if inspect.Info.KubeVendor == inspect.KubeVendorOpenShift {
		log.Info("On Openshift, skip MonitoringCR Reconcile.")
		return false, nil
	}

	// 3.2.0 non-ICP klusterlet monitoring
	found := inspect.ICPPrometheusService(client)
	if found { //found ICP Prometheus
		log.Info("Found ICP prometheus service, skip MonitoringCR Reconcile.")
		return false, nil
	}

	//No ICP and Openshift Monitoring
	monitoringCR, err := newMonitoringCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired Monitoring CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, monitoringCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return false, err
	}

	foundMonitoringCR := &multicloudv1beta1.Monitoring{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: monitoringCR.Name, Namespace: monitoringCR.Namespace}, foundMonitoringCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Monitoring CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.PrometheusIntegration.Enabled {
					log.V(5).Info("Monitoring ENABLED")
					err = createClusteRolesForMonitoring(instance, client, scheme)
					if err != nil {
						log.Error(err, "fail to CREATE Monitoring ClusterRoles")
						return false, err
					}
					err = createRootCACert(instance, client, scheme)
					if err != nil {
						return false, err
					}
					err = createClusterIssuer(instance, client, scheme)
					if err != nil {
						return false, err
					}
					err = create(instance, monitoringCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE Monitoring CR")
						return false, err
					}
				} else {
					log.V(5).Info("Monitoring DISABLED")
					err = finalize(instance, monitoringCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE Monitoring CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err = finalize(instance, monitoringCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE Monitoring CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("Monitoring CR DOES exist")
		if foundMonitoringCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("Monitoring CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.PrometheusIntegration.Enabled {
				log.Info("instance IS NOT in deletion state and Monitoring ENABLED")
				err = update(instance, monitoringCR, foundMonitoringCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE Monitoring CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or Monitoring DISABLED")
				err = delete(foundMonitoringCR, client)
				if err != nil {
					log.Error(err, "Fail to DELETE Monitoring CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for Monitoring")
				return true, nil
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for Monitoring")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled Monitoring")
	return false, nil
}

func createClusteRolesForMonitoring(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	labels := map[string]string{
		"app":                                instance.Name,
		"kubernetes.io/bootstrapping":        "rbac-defaults",
		"rbac.icp.com/aggregate-to-icp-view": "true",
	}

	viewAggregateClusteRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "monitoring-view-aggregate",
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"alertrules"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"monitoringdashboards"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"alertmanagers", "prometheuses", "prometheusrules", "servicemonitors"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}

	if err := createClusteRoles(client, viewAggregateClusteRole); err != nil {
		log.Error(err, "Unable to create viewAggregateClusteRole")
		return err
	}

	adminAggregateClusteRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "monitoring-admin-aggregate",
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"alertrules"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"monitoringdashboards"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"alertmanagers", "prometheuses", "prometheusrules", "servicemonitors"},
				Verbs:     []string{"create", "delete", "deletecollection", "get", "list", "patch", "update", "watch"},
			},
		},
	}

	if err := createClusteRoles(client, adminAggregateClusteRole); err != nil {
		log.Error(err, "Unable to create adminAggregateClusteRole")
		return err
	}

	editAggregateClusteRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "monitoring-edit-aggregate",
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"alertrules"},
				Verbs:     []string{"get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"monitoringdashboards"},
				Verbs:     []string{"get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"alertmanagers", "prometheuses", "prometheusrules", "servicemonitors"},
				Verbs:     []string{"get", "list", "patch", "update", "watch"},
			},
		},
	}

	if err := createClusteRoles(client, editAggregateClusteRole); err != nil {
		log.Error(err, "Unable to create editAggregateClusteRole")
		return err
	}

	operateAggregateClusteRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "monitoring-operate-aggregate",
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"alertrules"},
				Verbs:     []string{"create", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoringcontroller.cloud.ibm.com"},
				Resources: []string{"monitoringdashboards"},
				Verbs:     []string{"create", "get", "list", "patch", "update", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"monitoring.coreos.com"},
				Resources: []string{"alertmanagers", "prometheuses", "prometheusrules", "servicemonitors"},
				Verbs:     []string{"create", "get", "list", "patch", "update", "watch"},
			},
		},
	}

	if err := createClusteRoles(client, operateAggregateClusteRole); err != nil {
		log.Error(err, "Unable to create operateAggregateClusteRole")
		return err
	}

	return nil
}

func createRootCACert(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	certificate := &certmanagerv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-monitoring-ca-cert",
			Namespace: instance.Namespace,
		},
		Spec: certmanagerv1alpha1.CertificateSpec{
			CommonName: instance.Name + "-monitonring",
			IssuerRef: certmanagerv1alpha1.ObjectReference{
				Name: instance.Name + "-self-signed",
				Kind: "ClusterIssuer",
			},
			SecretName:   instance.Name + "-monitoring-ca-cert",
			IsCA:         true,
			Organization: []string{"IBM"},
		},
	}
	err := controllerutil.SetControllerReference(instance, certificate, scheme)
	if err != nil {
		return err
	}

	foundCertificate := &certmanagerv1alpha1.Certificate{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: certificate.Name, Namespace: certificate.Namespace}, foundCertificate)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Monitoring CA Certificate")
		return client.Create(context.TODO(), certificate)
	}

	return err
}

func createClusterIssuer(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Name + "-monitoring",
		},
		Spec: certmanagerv1alpha1.IssuerSpec{
			IssuerConfig: certmanagerv1alpha1.IssuerConfig{
				CA: &certmanagerv1alpha1.CAIssuer{
					SecretName: instance.Name + "-monitoring-ca-cert",
				},
			},
		},
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Monitoring ClusterIssuer")
		return client.Create(context.TODO(), clusterIssuer)
	}

	return err
}

func createClusteRoles(client client.Client, clusterRole *rbacv1.ClusterRole) error {
	foundClusterRole := &rbacv1.ClusterRole{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRole.Name, Namespace: clusterRole.Namespace}, foundClusterRole)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ClusterRole", "Name", clusterRole.Name, "Namespace", clusterRole.Namespace)
			err = client.Create(context.TODO(), clusterRole)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return err
}

func newMonitoringCR(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.Monitoring, error) {
	labels := map[string]string{
		"app": instance.Name,
	}
	prometheusImage, err := instance.GetImage("prometheus")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "monitoring")
		return nil, err
	}
	configmapReloadImage, err := instance.GetImage("configmap-reload")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "configmapReload")
		return nil, err
	}

	routerImage, err := instance.GetImage("router")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "router")
		return nil, err
	}

	prometheusOperatorControllerImage, err := instance.GetImage("prometheus-operator-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "prometheusOperatorController")
		return nil, err
	}

	prometheusOperatorImage, err := instance.GetImage("prometheus-operator")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "prometheusOperator")
		return nil, err
	}

	prometheusConfigReloaderImage, err := instance.GetImage("prometheus-config-reloader")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "prometheusConfigReloader")
		return nil, err
	}

	curlImage, err := instance.GetImage("curl")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "curl")
		return nil, err
	}

	return &multicloudv1beta1.Monitoring{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-monitoring",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.MonitoringSpec{
			FullNameOverride:           instance.Name + "-monitoring",
			Enabled:                    instance.Spec.PrometheusIntegration.Enabled,
			Mode:                       "standard",
			ImagePullSecret:            instance.Spec.ImagePullSecret,
			MonitoringFullnameOverride: "monitoring-monitoring",
			PrometheusFullnameOverride: "monitoring-prometheus",
			GrafanaFullnameOverride:    "monitoring-grafana",
			TLS: multicloudv1beta1.MonitoringTLS{
				Enabled: true,
				Issuer:  instance.Name + "-monitoring",
				CA: multicloudv1beta1.MonitoringTLSCA{
					SecretName: instance.Name + "-monitoring-ca-cert",
				},
			},
			Prometheus: multicloudv1beta1.MonitoringPrometheus{
				Image: prometheusImage,
				EtcdTarget: multicloudv1beta1.MonitoringPrometheusEtcdTarget{
					Enabled: false,
				},
			},
			Alertmanager: multicloudv1beta1.MonitoringAlertmanager{
				Enabled: false,
			},
			Grafana: multicloudv1beta1.MonitoringGrafana{
				Enabled: false,
			},
			KubeStateMetrics: multicloudv1beta1.MonitoringKubeStateMetrics{
				Enabled: false,
			},
			NodeExporter: multicloudv1beta1.MonitoringNodeExporter{
				Enabled: false,
			},
			CollectdExporter: multicloudv1beta1.MonitoringCollectdExporter{
				Enabled: false,
			},
			ElasticsearchExporter: multicloudv1beta1.MonitoringElasticsearchExporter{
				Enabled: false,
			},
			ConfigmapReload: multicloudv1beta1.MonitoringConfigmapReload{
				Image: configmapReloadImage,
			},
			Router: multicloudv1beta1.MonitoringRouter{
				Image: routerImage,
			},
			PrometheusOperatorController: multicloudv1beta1.MonitoringPrometheusOperatorController{
				Image: prometheusOperatorControllerImage,
			},
			PrometheusOperator: multicloudv1beta1.MonitoringPrometheusOperator{
				Image: prometheusOperatorImage,
			},
			PrometheusConfigReloader: multicloudv1beta1.MonitoringPrometheusConfigReloader{
				Image: prometheusConfigReloaderImage,
			},
			Curl: multicloudv1beta1.MonitoringCurl{
				Image: curlImage,
			},
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Monitoring, client client.Client) error {
	log.Info("Creating a new Monitoring", "Monitoring.Namespace", cr.Namespace, "Monitoring.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE Monitoring CR")
		return err
	}

	// Adding Finalizer to Instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Monitoring, foundCR *multicloudv1beta1.Monitoring, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE Monitoring CR")
		return err
	}

	// Adding Finalizer to Instance if Finalizer does not exist
	// NOTE: This is to handle requeue due to failed instance update during creation
	for _, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			return nil
		}
	}
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func delete(foundCR *multicloudv1beta1.Monitoring, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.Monitoring, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Delete All Cluster Roles
			err := deleteClusterRole(instance, client)
			if err != nil {
				log.V(5).Info("Did not delete clusterrole for monitoring")
			}

			// Delete Root CA Certificate
			err = deleteRootCACert(instance, client)
			if err != nil {
				log.V(5).Info("Did not delete RootCACert for monitoring")
			}

			// Delete Cluster Issuer
			err = deleteClusterIssuer(instance, client)
			if err != nil {
				log.V(5).Info("Did not delete ClusterIssuer for monitoring")
			}

			//Delete Secret
			err = deleteSecrets(instance, client)
			if err != nil {
				log.V(5).Info("Did not delete Secret for monitoring")
			}

			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}

func deleteClusterRole(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	clusterRolesToDelete := []string{
		"monitoring-view-aggregate",
		"monitoring-admin-aggregate",
		"monitoring-edit-aggregate",
		"monitoring-operate-aggregate",
	}

	for _, clusterRoleToDelete := range clusterRolesToDelete {
		foundClusterRoleToDelete := &rbacv1.ClusterRole{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleToDelete, Namespace: ""}, foundClusterRoleToDelete)

		if err == nil {
			for _, ownerReference := range foundClusterRoleToDelete.OwnerReferences {
				if instance.APIVersion == ownerReference.APIVersion &&
					instance.Kind == ownerReference.Kind &&
					instance.Name == ownerReference.Name {
					err := client.Delete(context.TODO(), foundClusterRoleToDelete)
					if err != nil {
						log.Error(err, "Fail to DELETE Monitoring Cluster Role", "ClusterRole.Name", clusterRoleToDelete)
						return err
					}
					break
				}
			}
		}
	}
	return nil
}

func deleteRootCACert(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	foundCertificate := &certmanagerv1alpha1.Certificate{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-monitoring-ca-cert", Namespace: instance.Namespace}, foundCertificate)
	if err == nil {
		log.Info("Deleting Monitoring CA Certificate")
		return client.Delete(context.TODO(), foundCertificate)
	}

	if errors.IsNotFound(err) {
		log.Info("Cannot find Monitoring CA Certificate, igore the deletion")
		return nil
	}

	return err
}

func deleteClusterIssuer(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-monitoring"}, foundClusterIssuer)
	if err == nil {
		log.Info("Deleting Monitoring ClusterIssuer")
		return client.Delete(context.TODO(), foundClusterIssuer)
	}

	if errors.IsNotFound(err) {
		log.Info("Cannot find Monitoring ClusterIssuer, igore the deletion")
		return nil
	}

	return err
}

func deleteSecrets(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	secretsToDelete := []string{
		instance.Name + "-monitoring-ca-cert",
		"monitoring-monitoring-certs",
		"monitoring-monitoring-client-certs",
		"monitoring-monitoring-elasticsearch-exporter-client-certs",
		"monitoring-monitoring-exporter-certs",
	}

	for _, secretToDelete := range secretsToDelete {
		foundSecretToDelete := &corev1.Secret{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: instance.Namespace}, foundSecretToDelete)
		if err == nil {
			err := client.Delete(context.TODO(), foundSecretToDelete)
			if err != nil {
				log.Error(err, "Fail to DELETE Monitoring Secret", "Secret.Name", secretToDelete)
				return err
			}
		}
	}
	return nil
}
