//Package v1beta1 of component Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"
	"os"
	"strings"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconcile Resolves differences in the running state of the klusterlet-component-operator deployment
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling ComponentOperator")

	var err error

	// Create Component Operator ClusteRole
	clusterRole := newClusterRole(instance)
	err = controllerutil.SetControllerReference(instance, clusterRole, scheme)
	if err != nil {
		return err
	}

	foundClusterRole := &rbacv1.ClusterRole{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterRole.Name, Namespace: clusterRole.Namespace}, foundClusterRole)
	if err != nil {
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
	err = controllerutil.SetControllerReference(instance, serviceAccount, scheme)
	if err != nil {
		return err
	}

	foundServiceAccount := &corev1.ServiceAccount{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil {
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
	err = controllerutil.SetControllerReference(instance, clusterRoleBinding, scheme)
	if err != nil {
		return err
	}

	foundClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBinding.Name, Namespace: clusterRoleBinding.Namespace}, foundClusterRoleBinding)
	if err != nil {
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
	err = controllerutil.SetControllerReference(instance, deployment, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundDeployment := &extensionsv1beta1.Deployment{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating Component Operator Deployment")
			err = client.Create(context.TODO(), deployment)
			if err != nil {
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
		err = client.Update(context.TODO(), foundDeployment)
		if err != nil {
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
			rbacv1.Subject{
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
			rbacv1.PolicyRule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			rbacv1.PolicyRule{
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

func newDeployment(instance *multicloudv1beta1.Endpoint) (*extensionsv1beta1.Deployment, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	var deploymentImage string
	if instance.Spec.ComponentOperatorImage != "" {
		deploymentImage = instance.Spec.ComponentOperatorImage
	} else {
		image, err := instance.GetImage("component-operator")
		if err != nil {
			return nil, err
		}
		deploymentImage = image.Repository + ":" + image.Tag
	}

	deployment := &extensionsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-component-operator",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: extensionsv1beta1.DeploymentSpec{
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
						corev1.Container{
							Name:            "klusterlet-component-opeator",
							Image:           deploymentImage,
							ImagePullPolicy: instance.Spec.ImagePullPolicy,
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "WATCH_NAMESPACE",
									Value: os.Getenv("WATCH_NAMESPACE"),
								},
								corev1.EnvVar{
									Name:  "OPERATOR_NAME",
									Value: "klusterlet-component-opeator",
								},
								corev1.EnvVar{
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
								"--zap-devel",
							},
						},
					},
				},
			},
		},
	}

	if instance.Spec.ImagePullSecret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			corev1.LocalObjectReference{
				Name: instance.Spec.ImagePullSecret,
			},
		}
	}

	return deployment, nil
}

func watchesFile(instance *multicloudv1beta1.Endpoint) string {
	versionSplit := strings.Split(instance.Spec.Version, "-")
	return "/opt/helm/versions/" + versionSplit[0] + "/watches.yaml"
}
