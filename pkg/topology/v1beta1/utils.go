// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of topology Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
)

// IsReady helps the other components to check whether the topology pod is ready
func IsReady(instance *multicloudv1beta1.Endpoint, c client.Client) (bool, error) {
	var collectorIsReady, appIsReady, daemonsetIsReady bool
	foundCollectorDeployment := &appsv1.Deployment{}
	err := c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-topology-weave-collector", Namespace: instance.Namespace}, foundCollectorDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the weave collector deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	log.V(5).Info("Found weave collector deployment")
	for _, condition := range foundCollectorDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			collectorIsReady = true
		}
	}

	foundAppDeployment := &appsv1.Deployment{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-topology-weave-scope-app", Namespace: instance.Namespace}, foundAppDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the weave app deployment")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	log.V(5).Info("Found weave app deployment")
	for _, condition := range foundAppDeployment.Status.Conditions {
		if condition.Type == "Available" && condition.Status == "True" {
			appIsReady = true
		}
	}

	foundDaemonset := &appsv1.DaemonSet{}
	err = c.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-topology-weave-scope", Namespace: instance.Namespace}, foundDaemonset)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Cannot find the weave daemonset")
			return false, nil
		}
		log.Error(err, "Unexpected ERROR")
		return false, err
	}
	log.V(5).Info("Found weave daemonset")
	if foundDaemonset.Status.CurrentNumberScheduled == foundDaemonset.Status.NumberReady {
		daemonsetIsReady = true
	}

	return collectorIsReady && appIsReady && daemonsetIsReady, nil
}
