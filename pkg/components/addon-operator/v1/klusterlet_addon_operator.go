// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1

import (
	"context"
	"fmt"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
)

// constant for klusterlet addon operator
const (
	KlusterletAddonOperator  = "klusterlet-addon-operator"
	KlusterletAddonNamespace = "open-cluster-management-agent-addon"
	ClusterRolePrefix        = "open-cluster-management:"
)

// NewClusterRoleBinding - template for cluster role bindiing
func NewClusterRoleBinding(instance *agentv1.KlusterletAddonConfig) *rbacv1.ClusterRoleBinding {
	labels := map[string]string{
		"app": instance.Name,
	}

	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ClusterRolePrefix + KlusterletAddonOperator,
			Labels: labels,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      KlusterletAddonOperator,
				Namespace: KlusterletAddonNamespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: ClusterRolePrefix + KlusterletAddonOperator,
		},
	}
}

// NewClusterRole - template for cluster role
func NewClusterRole(instance *agentv1.KlusterletAddonConfig) *rbacv1.ClusterRole {
	labels := map[string]string{
		"app": instance.Name,
	}

	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   ClusterRolePrefix + KlusterletAddonOperator,
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

// NewServiceAccount - template for service account
func NewServiceAccount(instance *agentv1.KlusterletAddonConfig, namespace string) *corev1.ServiceAccount {
	labels := map[string]string{
		"app": instance.Name,
	}

	serviceAccount := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KlusterletAddonOperator,
			Namespace: namespace,
			Labels:    labels,
		},
	}

	return serviceAccount
}

// NewNamespace - template for namespace
func NewNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: KlusterletAddonNamespace,
		},
	}
}

// NewDeployment -  template for klusterlet addon operator
func NewDeployment(instance *agentv1.KlusterletAddonConfig, namespace string) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	var deploymentImage string
	if instance.Spec.ComponentOperatorImage != "" {
		deploymentImage = instance.Spec.ComponentOperatorImage
	} else {
		_, imageRepository, err := instance.GetImage("addon-operator")
		if err != nil {
			return nil, err
		}
		deploymentImage = imageRepository
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      KlusterletAddonOperator,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": KlusterletAddonOperator,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": KlusterletAddonOperator,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: KlusterletAddonOperator,
					Containers: []corev1.Container{
						{
							Name:            KlusterletAddonOperator,
							Image:           deploymentImage,
							ImagePullPolicy: instance.Spec.ImagePullPolicy,
							Env: []corev1.EnvVar{
								{
									Name:  "WATCH_NAMESPACE",
									Value: os.Getenv("WATCH_NAMESPACE"),
								},
								{
									Name:  "OPERATOR_NAME",
									Value: KlusterletAddonOperator,
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

// NewImagePullSecret returns a secret for dockerconfig
// data is copied from instance.Namespace/instance.Spec.ImagePullSecret or POD_NAMESPACE/DEFAULT_IMAGE_PULL_SECRET
func NewImagePullSecret(instance *agentv1.KlusterletAddonConfig, client client.Client) (*corev1.Secret, error) {
	if instance.Spec.ImagePullSecret == "" {
		return nil, nil
	}

	secret := &corev1.Secret{}
	secretNsN := types.NamespacedName{
		Name:      instance.Spec.ImagePullSecret,
		Namespace: instance.Namespace,
	}
	defaultSecretNsN := types.NamespacedName{
		Name:      os.Getenv("DEFAULT_IMAGE_PULL_SECRET"),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
	//fetch secret from cluster namespace
	if err := client.Get(context.TODO(), secretNsN, secret); err != nil {
		if !errors.IsNotFound(err) && secretNsN.Name != defaultSecretNsN.Name {
			//fail to fetch cluster namespace secret and secret name is explicitly set to a value different from default
			return nil, err
		}

		//if not found fetch default secret from pod namespace
		if err := client.Get(context.TODO(), defaultSecretNsN, secret); err != nil {
			//fail to fetch default secret
			return nil, err
		}
	}

	//invalid secret type check
	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("secret is not of type corev1.SecretTypeDockerConfigJson")
	}

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: KlusterletAddonNamespace,
		},
		Data: secret.Data,
		Type: secret.Type,
	}, nil
}
