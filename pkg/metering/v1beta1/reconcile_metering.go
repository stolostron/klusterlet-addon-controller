// Package v1beta1 of metering provides a reconciler for the Metering
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
	//"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileSender Resolves differences in the running state of the MeteringSender services and CRDs.
func reconcileMetering(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Metering")

	var err error
	var meteringSenderCR *multicloudv1beta1.Metering

	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		meteringSenderCR, err = newMeteringSenderCRForICP(instance)
		//TODO(liuhao): add ICP on OpenShift
	} else {
		meteringSenderCR, err = newMeteringSenderCR(instance)
	}

	if err != nil {
		log.Error(err, "Fail to generate desired Metering CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, meteringSenderCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundMeteringSenderCR := &multicloudv1beta1.Metering{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: meteringSenderCR.Name, Namespace: meteringSenderCR.Namespace}, foundMeteringSenderCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("Metering CR for Sender DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("Instance IS NOT in deletion state")
				if instance.Spec.EndpointMeteringConfig.Enabled {
					log.V(5).Info("Metering ENABLED")
					err := create(instance, meteringSenderCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE Metering CR for Sender")
						return err
					}
				} else {
					log.V(5).Info("Metering DISABLED")
					err := finalize(instance, meteringSenderCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE Metering CR for Sender")
						return err
					}
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				err := finalize(instance, meteringSenderCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE Metering CR for Sender")
					return err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		log.V(5).Info("Metering CR for Sender DOES exist")
		if foundMeteringSenderCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("Metering CR for Sender IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.EndpointMeteringConfig.Enabled {
				log.Info("Instance IS NOT in deletion state and Metering ENABLED")
				err := update(instance, meteringSenderCR, foundMeteringSenderCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE Metering CR for Sender")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state or Metering DISABLED")
				if foundMeteringSenderCR.GetDeletionTimestamp() == nil {
					err := delete(instance, foundMeteringSenderCR, client)
					if err != nil {
						log.Error(err, "Fail to DELETE Metering CR for Sender")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled Metering Sender")
	return nil
}

func newMeteringSenderCR(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.Metering, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	senderImage, err := instance.GetImage("metering-sender")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "metering-sender")
		return nil, err
	}

	dmImage, err := instance.GetImage("metering-dm")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "metering-dm")
		return nil, err
	}

	readerImage, err := instance.GetImage("metering-reader")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "metering-reader")
		return nil, err
	}

	return &multicloudv1beta1.Metering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-metering",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.MeteringSpec{
			FullNameOverride: instance.Name + "-metering",
			API:              multicloudv1beta1.MeteringSpecAPI{Enabled: false},
			UI:               multicloudv1beta1.MeteringSpecUI{Enabled: false},
			McmUI:            multicloudv1beta1.MeteringSpecMcmUI{Enabled: false},

			DataManager: multicloudv1beta1.MeteringSpecDataManager{
				Enabled:                  true,
				Image:                    dmImage,
				NodeSelectorEnabled:      false,
				PriorityClassNameEnabled: false,
			},

			Reader: multicloudv1beta1.MeteringSpecReader{
				Enabled: true,
				Image:   readerImage,
			},

			Sender: multicloudv1beta1.MeteringSpecSender{
				Enabled:                  true,
				ClusterName:              instance.Spec.ClusterName,
				ClusterNamespace:         instance.Spec.ClusterNamespace,
				HubKubeConfigSecret:      instance.Name + "-connmgr-cert-store",
				Image:                    senderImage,
				NodeSelectorEnabled:      false,
				PriorityClassNameEnabled: false,
			},

			ImagePullSecrets: []string{instance.Spec.ImagePullSecret},
			Mongo: multicloudv1beta1.MeteringSpecMongo{
				ClusterCertsSecret: instance.Name + "-metering-ca-cert",
				ClientCertsSecret:  instance.Name + "-metering-mongodb-client-cert",
				Username: multicloudv1beta1.MeteringSpecMongoUsername{
					Secret: instance.Name + "-metering-mongodb-admin",
					Key:    "user",
				},
				Password: multicloudv1beta1.MeteringSpecMongoPassword{
					Secret: instance.Name + "-metering-mongodb-admin",
					Key:    "password",
				},
			},
			ServiceAccountNameEnabled: true,
			ClusterRoleEnabled:        true,
		},
	}, nil
}

func newMeteringSenderCRForICP(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.Metering, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	senderImage, err := instance.GetImage("metering-sender")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "metering-sender")
		return nil, err
	}

	return &multicloudv1beta1.Metering{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-metering",
			Namespace: "kube-system",
			Labels:    labels,
		},
		Spec: multicloudv1beta1.MeteringSpec{
			FullNameOverride: instance.Name + "-metering",
			API:              multicloudv1beta1.MeteringSpecAPI{Enabled: false},
			UI:               multicloudv1beta1.MeteringSpecUI{Enabled: false},
			McmUI:            multicloudv1beta1.MeteringSpecMcmUI{Enabled: false},
			DataManager:      multicloudv1beta1.MeteringSpecDataManager{Enabled: false},
			Reader:           multicloudv1beta1.MeteringSpecReader{Enabled: false},

			Sender: multicloudv1beta1.MeteringSpecSender{
				Enabled:                  true,
				ClusterName:              instance.Spec.ClusterName,
				ClusterNamespace:         instance.Spec.ClusterNamespace,
				HubKubeConfigSecret:      instance.Namespace + "/" + instance.Name + "-connmgr-hub-kubeconfig",
				Image:                    senderImage,
				NodeSelectorEnabled:      true,
				PriorityClassNameEnabled: true,
			},

			ImagePullSecrets: []string{instance.Spec.ImagePullSecret},
			Mongo: multicloudv1beta1.MeteringSpecMongo{
				ClusterCertsSecret: "cluster-ca-cert",
				ClientCertsSecret:  "icp-mongodb-client-cert",
				Username: multicloudv1beta1.MeteringSpecMongoUsername{
					Secret: "icp-mongodb-admin",
					Key:    "user",
				},
				Password: multicloudv1beta1.MeteringSpecMongoPassword{
					Secret: "icp-mongodb-admin",
					Key:    "password",
				},
			},
			ServiceAccountNameEnabled: false,
			ClusterRoleEnabled:        false,
		},
	}, nil
}
