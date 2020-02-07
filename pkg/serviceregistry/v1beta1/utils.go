// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of serviceregistry Defines the Reconciliation logic and required setup for serviceregistry.
package v1beta1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
)

// IsReady helps the other components to check whether the connmgr pod is ready
func IsReady(instance *multicloudv1beta1.Endpoint, c client.Client) (bool, error) {
	var svcregIsReady, coreDNSIsReady bool

	foundSvcRegDeployment := &appsv1.Deployment{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-svcreg", Namespace: instance.Namespace}, foundSvcRegDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the svcreg deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	log.V(5).Info("Svcreg deployment exist")
	for _, condition := range foundSvcRegDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			svcregIsReady = true
		}
	}

	foundSvcRegCoreDNSDeployment := &appsv1.Deployment{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-svcreg-coredns", Namespace: instance.Namespace}, foundSvcRegCoreDNSDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the svcreg coreDNS deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	log.V(5).Info("Svcreg coreDNS deployment exist")
	for _, condition := range foundSvcRegCoreDNSDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			coreDNSIsReady = true
		}
	}

	return svcregIsReady && coreDNSIsReady, nil
}
