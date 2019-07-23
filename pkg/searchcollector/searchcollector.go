// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package searchcollector

import (
	"context"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mcmv1alpha1 "github.ibm.com/IBMPrivateCloud/hcm-api/pkg/apis/mcm/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("searchcollector")

func Reconcile(instance *klusterletv1alpha1.KlusterletService, c client.Client, scheme *runtime.Scheme) error {
	// Deployed on hub
	clusteStatusList := &mcmv1alpha1.ClusterStatusList{}
	err := c.List(context.TODO(), &client.ListOptions{}, clusteStatusList)
	if err == nil {
		log.Info("Found clusterstatus.mcm.ibm.com, this is a hub cluster, skip SearchCollector Reconcile.")
		return nil
	}

	// Not deployed on hub
	searchCollectorCR := newSearchCollectorCR(instance, c)
	err = controllerutil.SetControllerReference(instance, searchCollectorCR, scheme)
	if err != nil {
		return err
	}

	foundSearchCollectorCR := &klusterletv1alpha1.SearchCollector{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: searchCollectorCR.Name, Namespace: searchCollectorCR.Namespace}, foundSearchCollectorCR)
	if err != nil && errors.IsNotFound(err) && instance.Spec.SearchCollectorConfig.Enabled {
		searchCollectorCR.Spec.TillerIntegration = newSearchCollectorTillerIntegration(instance, c)
		log.Info("Creating a new SearchCollector", "SearchCollector.Namespace", searchCollectorCR.Namespace, "SearchCollector.Name", searchCollectorCR.Name)
		err = c.Create(context.TODO(), searchCollectorCR)
		if err != nil {
			return err
		}
	} else if err == nil && instance.Spec.SearchCollectorConfig.Enabled == false {
		log.Info("Deleting SearchCollector", "SearchCollector.Namespace", foundSearchCollectorCR.Namespace, "SearchCollector.Name", foundSearchCollectorCR.Name)
		err = c.Delete(context.TODO(), foundSearchCollectorCR)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func newSearchCollectorCR(cr *klusterletv1alpha1.KlusterletService, client client.Client) *klusterletv1alpha1.SearchCollector {
	labels := map[string]string{
		"app": cr.Name,
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
			Image: image.Image{
				Repository: "ibmcom/search-collector",
				Tag:        "3.2.0",
				PullPolicy: "IfNotPresent",
			},
		},
	}
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
