// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package v1beta1 of component Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	multicloudv1beta1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/multicloud/v1beta1"
)

var log = logf.Log.WithName("component")

// Reconcile Resolves differences in the running state of the endpoint-component-operator deployment
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling ComponentOperator")

	var err error

	// Create Component Operator ClusteRole
	clusterRole := newClusterRole(instance)
	foundClusterRole := &rbacv1.ClusterRole{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRole.Name, Namespace: clusterRole.Namespace}, foundClusterRole); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ClusterRole", "Name", clusterRole.Name, "Namespace", clusterRole.Namespace)
			err = client.Create(context.TODO(), clusterRole)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create Component Operator ServiceAccount
	serviceAccount := newServiceAccount(instance)
	if err := controllerutil.SetControllerReference(instance, serviceAccount, scheme); err != nil {
		return err
	}

	foundServiceAccount := &corev1.ServiceAccount{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ServiceAccount", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
			err = client.Create(context.TODO(), serviceAccount)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create Component Operator ClusterRoleBinding
	clusterRoleBinding := newClusterRoleBinding(instance)
	foundClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name, Namespace: clusterRoleBinding.Namespace}, foundClusterRoleBinding); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating ClusterRoleBinding", "Name", clusterRoleBinding.Name, "Namespace", clusterRoleBinding.Namespace)
			err = client.Create(context.TODO(), clusterRoleBinding)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Create or Update Component Operator Deployment
	deployment, err := newDeployment(instance)
	if err != nil {
		log.Error(err, "Fail to desired component operator deployment")
		return err
	}
	if err := controllerutil.SetControllerReference(instance, deployment, scheme); err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundDeployment := &appsv1.Deployment{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Component Operator Deployment")
			if err := client.Create(context.TODO(), deployment); err != nil {
				log.Error(err, "Fail to CREATE Component Operator Deployment")
				return err
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		log.Info("Updating Component Operator Deployment")
		foundDeployment.Spec = deployment.Spec
		if err := client.Update(context.TODO(), foundDeployment); err != nil {
			log.Error(err, "Fail to UPDATE Component Operator Deployment")
			return err
		}
	}

	reqLogger.Info("Successfully Reconciled ComponentOperator")
	return nil
}

func newClusterRoleBinding(instance *multicloudv1beta1.Endpoint) *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app": instance.Name,
	}

	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   instance.Name + "-component-operator",
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      instance.Name + "-component-operator",
				Namespace: instance.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: instance.Name + "-component-operator",
		},
	}
}

func newClusterRole(instance *multicloudv1beta1.Endpoint) *rbacv1.ClusterRole {
	labels := map[string]string{
		"app": instance.Name,
	}

	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   instance.Name + "-component-operator",
			Labels: labels,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				APIGroups:       nil,
				NonResourceURLs: []string{"*"},
				Resources:       []string{},
				Verbs:           []string{"*"},
			},
		},
	}
}

func newServiceAccount(instance *multicloudv1beta1.Endpoint) *corev1.ServiceAccount {
	labels := map[string]string{
		"app": instance.Name,
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-component-operator",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
	}

	return serviceAccount
}

func newDeployment(instance *multicloudv1beta1.Endpoint) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	var deploymentImage string
	if instance.Spec.ComponentOperatorImage != "" {
		deploymentImage = instance.Spec.ComponentOperatorImage
	} else {
		_, imageRepository, err := instance.GetImage("component-operator")
		if err != nil {
			return nil, err
		}
		deploymentImage = imageRepository
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-component-operator",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": instance.Name + "-component-operator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": instance.Name + "-component-operator",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: instance.Name + "-component-operator",
					Containers: []corev1.Container{
						{
							Name:            "endpoint-component-operator",
							Image:           deploymentImage,
							ImagePullPolicy: instance.Spec.ImagePullPolicy,
							Env: []corev1.EnvVar{
								{
									Name:  "WATCH_NAMESPACE",
									Value: os.Getenv("WATCH_NAMESPACE"),
								},
								{
									Name:  "OPERATOR_NAME",
									Value: "endpoint-component-operator",
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
							},
							Args: []string{
								"--watches-file=" + watchesFile(instance),
							},
						},
					},
				},
			},
		},
	}

	if instance.Spec.ImagePullSecret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: instance.Spec.ImagePullSecret,
			},
		}
	}

	return deployment, nil
}

// TODO need to change this mechanism, version of operator shouldn't have
// to align with that of the component operator
func watchesFile(instance *multicloudv1beta1.Endpoint) string {
	versionSplit := strings.Split(instance.Spec.Version, "-")
	version := versionSplit[0]
	if version == "3.2.1.1910" {
		version = "3.2.1"
	}
	return "/opt/helm/versions/" + version + "/watches.yaml"
}
