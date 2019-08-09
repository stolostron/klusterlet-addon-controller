// Package v1beta1 of serviceregistry provides a reconciler for the search collector
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

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

// TODO(liuhao): the following method need to be refactored as instance method of ServiceRegistry struct
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
			Name:      instance.Name + "-service-registry",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.ServiceRegistrySpec{
			FullNameOverride: instance.Name + "-service-registry",
			Enabled:          instance.Spec.ServiceRegistryConfig.Enabled,
			ServiceRegistry: multicloudv1beta1.ServiceRegistryImage{
				Image: serviceRegistryImage,
			},
			CoreDNS: multicloudv1beta1.CoreDNS{
				Image:          coreDNSImage,
				DNSSuffix:      instance.Spec.ServiceRegistryConfig.CoreDNS.DNSSuffix,
				Plugins:        instance.Spec.ServiceRegistryConfig.CoreDNS.Plugins,
				ClusterProxyIP: instance.Spec.ServiceRegistryConfig.CoreDNS.ClusterProxyIP,
			},
			ImagePullSecret: instance.Spec.ImagePullSecret,
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
			// Remove finalizer
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
