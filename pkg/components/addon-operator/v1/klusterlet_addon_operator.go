// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

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

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
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
			Annotations: map[string]string{
				"workload.openshift.io/allowed": "management",
			},
		},
	}
}

// NewDeployment -  template for klusterlet addon operator
func NewDeployment(instance *agentv1.KlusterletAddonConfig, namespace string) (*appsv1.Deployment, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	deploymentImage, err := instance.GetImage("klusterlet_addon_operator")
	if err != nil {
		return nil, err
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
					Annotations: map[string]string{
						"target.workload.openshift.io/management": `{"effect": "PreferredDuringScheduling"}`,
					},
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
					NodeSelector: instance.Spec.NodeSelector,
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/infra",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
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
func NewImagePullSecret(pullSecretNamespace, pullSecret string, client client.Client) (*corev1.Secret, error) {
	if pullSecret == "" {
		return nil, nil
	}

	// pullSecret and pullSecretNamespace are from imageRegistry
	// if failed get from default env
	secret := &corev1.Secret{}
	secretNsN := types.NamespacedName{
		Name:      pullSecret,
		Namespace: pullSecretNamespace,
	}
	defaultSecretNsN := types.NamespacedName{
		Name:      os.Getenv("DEFAULT_IMAGE_PULL_SECRET"),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
	//fetch secret from customized namespace
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
