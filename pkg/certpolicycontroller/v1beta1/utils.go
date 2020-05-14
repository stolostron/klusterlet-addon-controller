// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of certpolicy Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	klusterletv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1beta1"
)

// IsReady helps the other components to check whether the certpolicy pod is ready
func IsReady(instance *klusterletv1beta1.Klusterlet, c client.Client) (bool, error) {
	foundDeployment := &appsv1.Deployment{}

	err := c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-certpolicy", Namespace: instance.Namespace}, foundDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the certpolicy deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}

	log.V(5).Info("Found certpolicy deployment")
	for _, condition := range foundDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			return true, nil
		}
	}
	return false, nil
}
