// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of searchcollector provides a reconciler for the search collector
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

var log = logf.Log.WithName("searchcollector")

// Reconcile reconciles the search collector
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling SearchCollector")

	// Deployed on hub
	if inspect.DeployedOnHub(client) {
		log.Info("Found clusterstatus.mcm.ibm.com, this is a hub cluster, skip SearchCollector Reconcile.")
		return false, nil
	}

	// Not deployed on hub
	searchCollectorCR, err := newSearchCollectorCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired SearchCollector CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, searchCollectorCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundSearchCollectorCR := &multicloudv1beta1.SearchCollector{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: searchCollectorCR.Name, Namespace: searchCollectorCR.Namespace}, foundSearchCollectorCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("SearchCollector DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.SearchCollectorConfig.Enabled {
					log.V(5).Info("SearchCollector ENABLED")
					err := create(instance, searchCollectorCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE SearchCollector CR")
						return false, err
					}
				} else {
					log.V(5).Info("SearchCollector DISABLED")
					err := finalize(instance, searchCollectorCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE SearchCollector CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				err := finalize(instance, searchCollectorCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE SearchCollector CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("SearchCollector CR DOES exist")
		if foundSearchCollectorCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("SearchCollector IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.SearchCollectorConfig.Enabled {
				log.V(5).Info("instance IS NOT in deletion state and Search Collector is ENABLED")
				err = update(instance, searchCollectorCR, foundSearchCollectorCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE SearchCollector CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or Search Collector is DISABLED")
				err := delete(foundSearchCollectorCR, client)
				if err != nil {
					log.Error(err, "fail to DELETE SearchCollector CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for SearchCollector")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for SearchCollector")
			return true, err
		}
	}

	reqLogger.Info("Successfully Reconciled SearchCollector")
	return false, nil
}

// TODO(liuhao): the following method need to be refactored as instance method of SearchCollector struct
func newSearchCollectorCR(instance *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.SearchCollector, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	image, err := instance.GetImage("search-collector")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "search-collector")
		return nil, err
	}

	return &multicloudv1beta1.SearchCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-search",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.SearchCollectorSpec{
			FullNameOverride:  instance.Name + "-search",
			ClusterName:       instance.Spec.ClusterName,
			ClusterNamespace:  instance.Spec.ClusterNamespace,
			ConnectionManager: instance.Name + "-connmgr",
			Image:             image,
			ImagePullSecret:   instance.Spec.ImagePullSecret,
		},
	}, err
}

func create(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.SearchCollector, client client.Client) error {
	// Create the CR and add the Finalizer to the instance
	log.Info("Creating a new SearchCollector", "SearchCollector.Namespace", cr.Namespace, "SearchCollector.Name", cr.Name)
	err := client.Create(context.TODO(), cr)
	if err != nil {
		log.Error(err, "Fail to CREATE SearchCollector CR")
		return err
	}

	// Adding Finalizer to instance
	instance.Finalizers = append(instance.Finalizers, cr.Name)
	return nil
}

func update(instance *multicloudv1beta1.Endpoint, cr *multicloudv1beta1.SearchCollector, foundCR *multicloudv1beta1.SearchCollector, client client.Client) error {
	foundCR.Spec = cr.Spec
	err := client.Update(context.TODO(), foundCR)
	if err != nil && !errors.IsConflict(err) {
		log.Error(err, "fail to UPDATE SearchCollector CR")
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

func delete(foundCR *multicloudv1beta1.SearchCollector, client client.Client) error {
	return client.Delete(context.TODO(), foundCR)
}

func finalize(instance *multicloudv1beta1.Endpoint, searchCollectorCR *multicloudv1beta1.SearchCollector, client client.Client) error {
	for i, finalizer := range instance.Finalizers {
		if finalizer == searchCollectorCR.Name {
			instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
			return nil
		}
	}
	return nil
}
