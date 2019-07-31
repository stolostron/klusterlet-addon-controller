/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package tiller

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetICPTillerDefaultAdminUser gets the ICP tiller default admin user
func GetICPTillerDefaultAdminUser(client client.Client) string {
	findICPTillerDeployment := &extensionsv1beta1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, findICPTillerDeployment)
	if err != nil {
		return "admin"
	}

	for _, container := range findICPTillerDeployment.Spec.Template.Spec.Containers {
		if container.Name == "tiller" {
			for _, env := range container.Env {
				if env.Name == "default_admin_user" {
					return env.Value
				}
			}
		}
	}

	return "admin"
}

// GetICPTillerServiceEndpoint gets the ICP tiller endpoint
func GetICPTillerServiceEndpoint(client client.Client) string {
	foundICPTillerService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, foundICPTillerService)
	if err != nil {
		return ""
	}
	if err == nil {
		tillerServiceHostname := foundICPTillerService.Name + "." + foundICPTillerService.Namespace
		var tillerServicePort int32

		for _, port := range foundICPTillerService.Spec.Ports {
			if port.Name == "grpc" && port.Protocol == "TCP" {
				tillerServicePort = port.Port
			}
		}

		return tillerServiceHostname + ":" + strconv.FormatInt(int64(tillerServicePort), 10)
	}
	return ""
}
