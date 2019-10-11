// Package v1beta1 of metering provides a reconciler for the Metering
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
)

// Reconcile Resolves differences in the running state of the Metering services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	err := reconcileMongoDB(instance, client, scheme)
	if err != nil {
		return err
	}

	err = reconcileMetering(instance, client, scheme)
	if err != nil {
		return err
	}

	return nil
}
