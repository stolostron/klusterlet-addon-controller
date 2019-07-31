// Package searchcollector provides a reconciler for the search collector
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package searchcollector

import (
	"context"

	mcmv1alpha1 "github.ibm.com/IBMPrivateCloud/hcm-api/pkg/apis/mcm/v1alpha1"
	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("searchcollector")

// Reconcile reconciles the search collector
func Reconcile(instance *klusterletv1alpha1.KlusterletService, c client.Client, scheme *runtime.Scheme) error {
	// Deployed on hub
	clusteStatusList := &mcmv1alpha1.ClusterStatusList{}
	err := c.List(context.TODO(), &client.ListOptions{}, clusteStatusList)
	if err == nil {
		log.Info("Found clusterstatus.mcm.ibm.com, this is a hub cluster, skip SearchCollector Reconcile.")
		return nil
	}

	// Not deployed on hub
	searchCollectorCR, err := newSearchCollectorCR(instance, c)
	if err != nil {
		log.Error(err, "Fail to generate desired SearchCollector CR")
		return err
	}
	err = controllerutil.SetControllerReference(instance, searchCollectorCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundSearchCollectorCR := &klusterletv1alpha1.SearchCollector{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: searchCollectorCR.Name, Namespace: searchCollectorCR.Namespace}, foundSearchCollectorCR)
	if err != nil {
		if errors.IsNotFound(err) {
			// Search Collector does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// Klusterlet Service is NOT being deleted
				if instance.Spec.SearchCollectorConfig.Enabled {
					// Search Collector is ENABLED
					// Create the CR and add the Finalizer to the instance
					log.Info("Creating a new SearchCollector", "SearchCollector.Namespace", searchCollectorCR.Namespace, "SearchCollector.Name", searchCollectorCR.Name)
					err = c.Create(context.TODO(), searchCollectorCR)
					if err != nil {
						log.Error(err, "Fail to CREATE SearchCollector CR")
						return err
					}

					// Adding Finalizer to KlusterletService instance
					instance.Finalizers = append(instance.Finalizers, searchCollectorCR.Name)
				}
			} else {
				// Klusterlet Service is being deleted
				// Cleanup Secrets
				secretsToDeletes := []string{
					searchCollectorCR.Name + "-tiller-client-certs",
				}

				for _, secretToDelete := range secretsToDeletes {
					foundSecretToDelete := &corev1.Secret{}
					err = c.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: searchCollectorCR.Namespace}, foundSecretToDelete)
					if err == nil {
						err = c.Delete(context.TODO(), foundSecretToDelete)
						if err != nil {
							log.Error(err, "Fail to DELETE ConnectionManager Secret", "Secret.Name", secretToDelete)
							return err
						}
					}
				}

				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == searchCollectorCR.Name {
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
		if foundSearchCollectorCR.GetDeletionTimestamp() == nil {
			//Search Collector DOES exist
			if instance.GetDeletionTimestamp() == nil && instance.Spec.SearchCollectorConfig.Enabled {
				// KlusterletService NOT in deletion state and Search Collector is ENABLED
				foundSearchCollectorCR.Spec = searchCollectorCR.Spec
				err = c.Update(context.TODO(), foundSearchCollectorCR)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE SearchCollector CR")
					return err
				}
			} else {
				// KlusterletService in deletion state or Search Collector is DISABLED
				err = c.Delete(context.TODO(), foundSearchCollectorCR)
				if err != nil {
					log.Error(err, "Fail to DELETE SearchCollector CR")
					return err
				}
			}
		}
	}

	return nil
}

func newSearchCollectorCR(cr *klusterletv1alpha1.KlusterletService, client client.Client) (*klusterletv1alpha1.SearchCollector, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	image, err := cr.GetImage("search-collector")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "search-collector")
		return nil, err
	}

	return &klusterletv1alpha1.SearchCollector{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-search",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.SearchCollectorSpec{
			FullNameOverride:  cr.Name + "-search",
			ClusterName:       cr.Spec.ClusterName,
			ClusterNamespace:  cr.Spec.ClusterNamespace,
			ConnectionManager: cr.Name + "-connmgr",
			TillerIntegration: newSearchCollectorTillerIntegration(cr, client),
			Image:             image,
			ImagePullSecret:   cr.Spec.ImagePullSecret,
		},
	}, err
}

func newSearchCollectorTillerIntegration(cr *klusterletv1alpha1.KlusterletService, client client.Client) klusterletv1alpha1.SearchCollectorTillerIntegration {
	if cr.Spec.TillerIntegration.Enabled {
		// ICP Tiller
		icpTillerServiceEndpoint := tiller.GetICPTillerServiceEndpoint(client)
		if icpTillerServiceEndpoint != "" {
			return klusterletv1alpha1.SearchCollectorTillerIntegration{
				Enabled:       true,
				Endpoint:      icpTillerServiceEndpoint,
				CertIssuer:    "icp-ca-issuer",
				AutoGenSecret: true,
				User:          tiller.GetICPTillerDefaultAdminUser(client),
			}
		}

		// KlusterletOperator deployed Tiller
		return klusterletv1alpha1.SearchCollectorTillerIntegration{
			Enabled:       true,
			Endpoint:      cr.Name + "-tiller" + ":44134",
			CertIssuer:    cr.Name + "-tiller",
			AutoGenSecret: true,
			User:          cr.Name + "-admin",
		}
	}

	return klusterletv1alpha1.SearchCollectorTillerIntegration{
		Enabled: false,
	}
}
