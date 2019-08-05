// Package v1alpha1 of serviceregistry provides a reconciler for the search collector
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1alpha1

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("serviceRegistry")

// Reconcile reconciles the service registry
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("KlusterletService.Namespace", instance.Namespace, "KlusterletService.Name", instance.Name)
	reqLogger.Info("Reconciling ServiceRegistry")

	serviceRegisryCR, err := newServiceRegistryCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired ServiceRegisry CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, serviceRegisryCR, scheme)
	if err != nil {
		log.Error(err, "Error setting controller reference")
		return err
	}

	foundServiceRegisryCR := &klusterletv1alpha1.ServiceRegistry{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceRegisryCR.Name, Namespace: serviceRegisryCR.Namespace}, foundServiceRegisryCR)
	if err != nil {
		if errors.IsNotFound(err) {
			//Service Registry does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// Klusterlet Service is NOT being deleted
				if instance.Spec.ServiceRegistryConfig.Enabled {
					// Service Registry is ENABLED
					// Create the CR and add the Finalizer to the instance
					log.Info("Creating a new ServiceRegisry", "ServiceRegisry.Namespace", serviceRegisryCR.Namespace, "ServiceRegisry.Name", serviceRegisryCR.Name)
					err = client.Create(context.TODO(), serviceRegisryCR)
					if err != nil {
						log.Error(err, "Fail to CREATE ServiceRegisry CR")
						return err
					}

					// Adding Finalizer to KlusterletService instance
					instance.Finalizers = append(instance.Finalizers, serviceRegisryCR.Name)
				}
			} else {
				// Klusterlet Service is being deleted
				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == serviceRegisryCR.Name {
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
		//ServiceRegisry DOES exist
		if foundServiceRegisryCR.GetDeletionTimestamp() == nil {
			if instance.GetDeletionTimestamp() == nil && instance.Spec.ServiceRegistryConfig.Enabled {
				// KlusterletService NOT in deletion state and Service Registry is ENABLED
				// Update the ServiceRegisryCR
				foundServiceRegisryCR.Spec = serviceRegisryCR.Spec
				err = client.Update(context.TODO(), foundServiceRegisryCR)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE ServiceRegisry CR")
					return err
				}
			} else {
				// KlusterletService in deletion state or Service Registry is DISABLED
				err = client.Delete(context.TODO(), foundServiceRegisryCR)
				if err != nil {
					log.Error(err, "Fail to DELETE ServiceRegisry CR")
					return err
				}
			}
		}
	}

	return nil
}

func newServiceRegistryCR(cr *klusterletv1alpha1.KlusterletService) (*klusterletv1alpha1.ServiceRegistry, error) {
	labels := map[string]string{
		"app": cr.Name,
	}
	serviceRegistryImage, err := cr.GetImage("service-registry")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "service registry")
		return nil, err
	}
	coreDNSImage, err := cr.GetImage("coredns")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "coredns")
		return nil, err
	}

	return &klusterletv1alpha1.ServiceRegistry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-service-registry",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.ServiceRegistrySpec{
			FullNameOverride: cr.Name + "-service-registry",
			Enabled:          cr.Spec.ServiceRegistryConfig.Enabled,
			ServiceRegistry: klusterletv1alpha1.ServiceRegistryImage{
				Image: serviceRegistryImage,
			},
			CoreDNS: klusterletv1alpha1.CoreDNS{
				Image:          coreDNSImage,
				DNSSuffix:      cr.Spec.ServiceRegistryConfig.CoreDNS.DNSSuffix,
				Plugins:        cr.Spec.ServiceRegistryConfig.CoreDNS.Plugins,
				ClusterProxyIP: cr.Spec.ServiceRegistryConfig.CoreDNS.ClusterProxyIP,
			},
		},
	}, nil
}
