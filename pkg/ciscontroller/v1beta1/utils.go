// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of ciscontroller provides a reconciler for the CISPolicy
package v1beta1

import (
	"context"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IsReady helps the other components to check whether the ciscontroller pod is ready
func IsReady(instance *multicloudv1beta1.Endpoint, c client.Client) (bool, error) {
	foundDeployment := &appsv1.Deployment{}

	err := c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-search", Namespace: instance.Namespace}, foundDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the search deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}

	log.V(5).Info("Found search deployment")
	for _, condition := range foundDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			return true, nil
		}
	}
	return false, nil
}
