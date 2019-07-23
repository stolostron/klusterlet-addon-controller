/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package connmgr

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("connmgr")

// Reconcile Resolves differences in the running state of the connection manager services and CRDs.
func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	connMgrCR := newConnectionManagerCR(instance)
	err := controllerutil.SetControllerReference(instance, connMgrCR, scheme)
	if err != nil {
		return err
	}

	foundConnMgrCR := &klusterletv1alpha1.ConnectionManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: connMgrCR.Name, Namespace: connMgrCR.Namespace}, foundConnMgrCR)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new ConnectionManager", "ConnectionManager.Namespace", connMgrCR.Namespace, "ConnectionManager.Name", connMgrCR.Name)
		err := client.Create(context.TODO(), connMgrCR)
		if err != nil {
			return err
		}
	}
	return nil
}

func newConnectionManagerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.ConnectionManager {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.ConnectionManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-connmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.ConnectionManagerSpec{
			ClusterName:      cr.Spec.ClusterName,
			ClusterNamespace: cr.Spec.ClusterNamespace,
			BootStrapConfig:  cr.Spec.BootStrapConfig,
			FullNameOverride: cr.Name + "-connmgr",
		},
	}
}
