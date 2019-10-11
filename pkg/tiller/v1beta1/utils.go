// Package v1beta1 of tiller provides a reconciler for the Tiller
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
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

// CheckDependency helps the other components to check whether they depends on tiller, and helps them to restart the pod if tiller is created after them
func CheckDependency(instance *multicloudv1beta1.Endpoint, c client.Client, componentName string) error {
	foundPodList := &corev1.PodList{}

	err := c.List(context.TODO(), &client.ListOptions{Namespace: instance.Namespace}, foundPodList)
	if err != nil {
		log.Error(err, "Fail to list all the pods")
		return err
	}

	var (
		componentPodName, tillerPodName     string
		componentPod                        corev1.Pod
		componentTimestamp, tillerTimestamp metav1.Time
	)

	//go through the pod list to find the tiller pod and component pod
	for _, pod := range foundPodList.Items {
		if strings.Contains(pod.ObjectMeta.Name, "tiller") {
			log.V(5).Info("Found tiller pod")
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					tillerTimestamp = condition.LastTransitionTime
					tillerPodName = pod.ObjectMeta.Name
					break
				}
			}
		} else if strings.Contains(pod.ObjectMeta.Name, componentName) {
			log.V(5).Info("Found component pod", "Component.Name", componentName)
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					componentTimestamp = condition.LastTransitionTime
					componentPodName = pod.ObjectMeta.Name
					componentPod = pod
					break
				}
			}
		}
	}

	//cannot find tiller pod and component pod means do not need to make any action
	if tillerPodName == "" || componentPodName == "" {
		log.V(5).Info("Tiller or component pods are not found or not ready, will ignore the check", "Component.Name", componentName)
		return nil
	}

	log.V(5).Info("Both tiller and component pods are found and ready", "Component.Name", componentName)
	//find tiller and tiller is created after the component pod requirs to restart the component pod
	if tillerTimestamp.Time.After(componentTimestamp.Time) {
		log.V(5).Info("tiller pod is created after component pod, need to restart component pod", "Component.Name", componentName)
		err := c.Delete(context.TODO(), &componentPod)
		if err != nil {
			log.Error(err, "Fail to delete the component pod", "Component.Name", componentName)
			return err
		}
	}

	return nil
}
