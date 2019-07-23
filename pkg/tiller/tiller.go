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

	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("tiller")

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	// No Tiller Integration
	if !instance.Spec.TillerIntegration.Enabled {
		log.Info("Tiller Integration disabled, skip Tiller Reconcile.")
		return nil
	}

	// ICP Tiller
	foundICPTillerService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "tiller-deploy", Namespace: "kube-system"}, foundICPTillerService)
	if err == nil {
		log.Info("Found ICP Tiller, skip TillerCR Reconcile.")
		return nil
	}

	// No ICP Tiller
	tillerCR := newTillerCR(instance)
	err = controllerutil.SetControllerReference(instance, tillerCR, scheme)
	if err != nil {
		return err
	}

	foundTillerCR := &klusterletv1alpha1.Tiller{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: tillerCR.Name, Namespace: tillerCR.Namespace}, foundTillerCR)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Tiller", "Tiller.Namespace", tillerCR.Namespace, "Tiller.Name", tillerCR.Name)
		err = client.Create(context.TODO(), tillerCR)
		if err != nil {
			return err
		}
	}
	return nil
}

func newTillerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.Tiller {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.Tiller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-tiller",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.TillerSpec{
			FullNameOverride: cr.Name + "-tiller",
			CACertIssuer:     cr.Name + "-self-signed",
			DefaultAdminUser: cr.Name + "-admin",
			Image: image.Image{
				Repository: "ibmcom/tiller",
				Tag:        "v2.12.3-icp-3.2.0",
				PullPolicy: "IfNotPresent",
			},
		},
	}
}

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
