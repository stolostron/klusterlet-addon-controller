//Package component Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package component

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconcile Resolves differences in the running state of the klusterlet-component-operator deployment
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("KlusterletService.Namespace", instance.Namespace, "KlusterletService.Name", instance.Name)
	reqLogger.Info("Reconciling ComponentOperator")

	var err error

	// Create or Update Component Operator ClusteRole
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

	// Create or Update Component Operator ServiceAccount
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

	// Create or Update Component Operator ClusterRoleBinding
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
	deployment := newDeployment(instance)
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
	}

	reqLogger.Info("Successfully Reconciled ComponentOperator")
	return nil
}

func newClusterRoleBinding(instance *klusterletv1alpha1.KlusterletService) *rbacv1.ClusterRoleBinding {
	clusteRoleBindingFile := "/opt/component-operator/deploy/" + instance.Spec.Version + "/cluster_role_binding.yaml"

	clusteRoleBindingYAML, err := ioutil.ReadFile(clusteRoleBindingFile)
	if err != nil {
		log.Error(err, "Fail to Read cluster_role_binding.yaml", "filename", clusteRoleBindingYAML)
		return nil
	}

	clusteRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = yaml.Unmarshal(clusteRoleBindingYAML, clusteRoleBinding)
	if err != nil {
		log.Error(err, "Fail to Unmarshal cluster_role_binding.yaml", "content", clusteRoleBindingYAML)
		return nil
	}

	return clusteRoleBinding
}

func newClusterRole(instance *klusterletv1alpha1.KlusterletService) *rbacv1.ClusterRole {
	clusteRoleFile := "/opt/component-operator/deploy/" + instance.Spec.Version + "/cluster_role.yaml"

	clusteRoleYAML, err := ioutil.ReadFile(clusteRoleFile)
	if err != nil {
		log.Error(err, "Fail to Read cluster_role_binding.yaml", "filename", clusteRoleYAML)
		return nil
	}

	clusteRole := &rbacv1.ClusterRole{}
	err = yaml.Unmarshal(clusteRoleYAML, clusteRole)
	if err != nil {
		log.Error(err, "Fail to Unmarshal cluster_role_binding.yaml", "content", clusteRoleYAML)
		return nil
	}

	return clusteRole
}

func newServiceAccount(instance *klusterletv1alpha1.KlusterletService) *corev1.ServiceAccount {
	serviceAccountFile := "/opt/component-operator/deploy/" + instance.Spec.Version + "/service_account.yaml"

	serviceAccountYAML, err := ioutil.ReadFile(serviceAccountFile)
	if err != nil {
		log.Error(err, "Fail to Read service_account.yaml", "filename", serviceAccountYAML)
		return nil
	}

	serviceAccount := &corev1.ServiceAccount{}
	err = yaml.Unmarshal(serviceAccountYAML, serviceAccount)
	if err != nil {
		log.Error(err, "Fail to Unmarshal", "content", serviceAccountYAML)
		return nil
	}

	serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{Name: instance.Spec.ImagePullSecret})

	return serviceAccount
}

func newDeployment(instance *klusterletv1alpha1.KlusterletService) *extensionsv1beta1.Deployment {
	deploymentFile := "/opt/component-operator/deploy/" + instance.Spec.Version + "/operator.yaml"

	deploymentYAML, err := ioutil.ReadFile(deploymentFile)
	if err != nil {
		log.Error(err, "Fail to Read operator.yaml", "filename", deploymentFile)
		return nil
	}

	deployment := &extensionsv1beta1.Deployment{}
	err = yaml.Unmarshal(deploymentYAML, deployment)
	if err != nil {
		log.Error(err, "Fail to Unmarshal", "content", deploymentYAML)
		return nil
	}

	deployment.Name = instance.Name + "-component-operator"
	deployment.Namespace = instance.Namespace
	deployment.Labels = map[string]string{"app": instance.Name}
	deployment.Spec.Selector.MatchLabels = map[string]string{"name": deployment.Name}
	deployment.Spec.Template.Labels = deployment.Spec.Selector.MatchLabels
	deployment.Spec.Template.Spec.ServiceAccountName = deployment.Name

	container := deployment.Spec.Template.Spec.Containers[0]
	if container.Name == "klusterlet-component-operator" {
		container.Image = containerImage(container, instance)
		container.ImagePullPolicy = instance.Spec.ImagePullPolicy
		container.Args = append(container.Args, "--watches-file="+watchesFile(instance))
		container.Args = append(container.Args, "--zap-devel")
		for _, env := range container.Env {
			switch name := env.Name; name {
			case "WATCH_NAMESPACE":
				env.Value = os.Getenv("WATCH_NAMESPACE")
			case "OPERATOR_NAME":
				env.Value = deployment.Name
			}
		}
	}

	deployment.Spec.Template.Spec.Containers[0] = container

	return deployment
}

func containerImage(container corev1.Container, instance *klusterletv1alpha1.KlusterletService) string {
	if instance.Spec.ImageRegistry == "" {
		return container.Image
	}
	return strings.Join([]string{instance.Spec.ImageRegistry, container.Image}, "/")
}

func watchesFile(instance *klusterletv1alpha1.KlusterletService) string {
	return "/opt/helm/versions/" + instance.Spec.Version + "/watches.yaml"
}
