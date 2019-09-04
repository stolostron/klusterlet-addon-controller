// Package v1beta1 of serviceregistry provides a reconciler for the search collector
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"strconv"
	"time"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("serviceregistry")

// Reconcile reconciles the service registry
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling ServiceRegistry")

	serviceRegisryCR, err := newServiceRegistryCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired ServiceRegisry CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, serviceRegisryCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return false, err
	}

	foundServiceRegisryCR := &multicloudv1beta1.ServiceRegistry{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceRegisryCR.Name, Namespace: serviceRegisryCR.Namespace}, foundServiceRegisryCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("ServiceRegistry CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.ServiceRegistryConfig.Enabled {
					log.V(5).Info("ServiceRegistry ENABLED")
					err = createCoreDNSConfigmap(instance, client, serviceRegisryCR, scheme)
					if err != nil {
						log.Error(err, "fail to CREATE ServiceRegistry CoreDNS Configmap")
						return false, err
					}
					err = create(instance, serviceRegisryCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE ServiceRegistry CR")
						return false, err
					}
				} else {
					log.V(5).Info("ServiceRegistry DISABLED")
					err = finalize(instance, serviceRegisryCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE ServiceRegistry CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err = finalize(instance, serviceRegisryCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE ServiceRegistry CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("ServiceRegisry CR DOES exist")
		if foundServiceRegisryCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("ServiceRegisry CR is not in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.ServiceRegistryConfig.Enabled {
				log.Info("Instance IS NOT in deletion state and ServiceRegistry ENABLED")
				err := update(instance, serviceRegisryCR, foundServiceRegisryCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE ServiceRegisry CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or ServiceRegistry DISABLED")
				err = delete(foundServiceRegisryCR, client)
				if err != nil {
					log.Error(err, "Fail to DELETE ServiceRegistry CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for ConnectionManager")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for ConnectionManager")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled ServiceRegisry")
	return false, nil
}

func createCoreDNSConfigmap(instance *multicloudv1beta1.Endpoint, client client.Client, cr *multicloudv1beta1.ServiceRegistry, scheme *runtime.Scheme) error {
	labels := map[string]string{
		"app": instance.Name,
	}

	timeSecond := strconv.FormatInt(time.Now().Unix(), 10)

	dnsSuffix := cr.Spec.DNSSuffix
	if dnsSuffix == "" {
		dnsSuffix = "mcm.svc"
	}

	coreDNSConfigmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-svcreg-coredns",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{
			"Corefile":       ".:53 {\n    errors\n    log\n    health\n    reload\n    cache 30\n    file /etc/coredns/svcregistry.db mcm.svc {\n      reload 5s\n    }\n}",
			"svcregistry.db": dnsSuffix + ".    IN    SOA    mcm-svc-registry-dns." + instance.Namespace + ".svc.cluster.local. hostmaster.cluster.local. " + timeSecond + " 7200 1800 86400 30",
		},
	}

	err := controllerutil.SetControllerReference(instance, coreDNSConfigmap, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return err
	}

	err = client.Create(context.TODO(), coreDNSConfigmap)
	if err != nil {
		log.Error(err, "Fail to CREATE CoreDNS Configmap")
		return err
	}

	return nil
}

func newServiceRegistryCR(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.ServiceRegistry, error) {
	labels := map[string]string{
		"app": instance.Name,
	}
	serviceRegistryImage, err := instance.GetImage("service-registry")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "service registry")
		return nil, err
	}
	coreDNSImage, err := instance.GetImage("coredns")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "coredns")
		return nil, err
	}

	return &multicloudv1beta1.ServiceRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-svcreg",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.ServiceRegistrySpec{
			FullNameOverride: instance.Name + "-svcreg",
			ServiceRegistry: multicloudv1beta1.ServiceRegistryImage{
				Image: serviceRegistryImage,
			},
			CoreDNS: multicloudv1beta1.CoreDNSImage{
				Image: coreDNSImage,
			},
			ConnectionManager:                  instance.Name + "-connmgr",
			DNSSuffix:                          instance.Spec.ServiceRegistryConfig.DNSSuffix,
			Plugins:                            instance.Spec.ServiceRegistryConfig.Plugins,
			IstioIngressGateway:                instance.Spec.ServiceRegistryConfig.IstioIngressGateway,
			IstioServiceEntryRegistryNamespace: instance.Spec.ServiceRegistryConfig.IstioserviceEntryRegistryNamespace,
			ImagePullSecret:                    instance.Spec.ImagePullSecret,
		},
	}, nil
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ServiceRegistry, client client.Client) error {
	log.Info("Creating a new ServiceRegistry", "ServiceRegistry.Namespace", cr.Namespace, "ServiceRegistry.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE ServiceRegistry CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ServiceRegistry, foundCR *multicloudv1beta1.ServiceRegistry, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "Fail to UPDATE ServiceRegistry CR")
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

func delete(foundCR *multicloudv1beta1.ServiceRegistry, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.ServiceRegistry, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == cr.Name {
			// Delete CoreDNSConfigmap
			err := deleteCoreDNSConfigmap(instance, client)
			if err != nil {
				log.Error(err, "Fail to delete CoreDNSConfigmap for serviceReistry")
				return err
			}

			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}

func deleteCoreDNSConfigmap(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	foundCoreDNSConfigmap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-svcreg-coredns", Namespace: instance.Namespace}, foundCoreDNSConfigmap)
	if err == nil {
		log.Info("Deleting serviceReistry CoreDNSConfigmap")
		return client.Delete(context.TODO(), foundCoreDNSConfigmap)
	}

	if errors.IsNotFound(err) {
		log.Info("Cannot find serviceReistry CoreDNSConfigmap, igore the deletion")
		return nil
	}

	return err
}
