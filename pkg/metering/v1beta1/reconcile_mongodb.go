// Package v1beta1 of metering provides a reconciler for the Metering
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
	mongodb "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/mongodb/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileMongoDB Resolves differences in the running state of the MongoDB services for metering.
func reconcileMongoDB(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling Metering MongoDB")

	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		//TODO(liuhao): add ICP on OpenShift
		return nil
	}

	// Not on ICP
	meteringMongoDBCR, err := newMeteringMongoDBCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired Metering CR for Collector")
		return err
	}

	err = controllerutil.SetControllerReference(instance, meteringMongoDBCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundMeteringMongoDBCR := &multicloudv1beta1.MongoDB{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: meteringMongoDBCR.Name, Namespace: meteringMongoDBCR.Namespace}, foundMeteringMongoDBCR)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(5).Info("MongoDB CR for Metering DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("Instance IS NOT in deletion state")
				if instance.Spec.EndpointMeteringConfig.Enabled {
					log.V(5).Info("Metering ENABLED")
					err = createRootCACert(instance, client, scheme)
					if err != nil {
						return err
					}
					err = createClusterIssuer(instance, client, scheme)
					if err != nil {
						return err
					}
					err := mongodb.Create(instance, meteringMongoDBCR, client)
					if err != nil {
						log.Error(err, "fail to CREATE MongoDB CR for Metering")
						return err
					}
				} else {
					log.V(5).Info("Metering DISABLED")
					err := deleteSecrets(instance, client)
					if err != nil {
						log.Error(err, "fail to delete MongoDB Secrets for Metering")
						return err
					}
					err = mongodb.Finalize(instance, meteringMongoDBCR, client)
					if err != nil {
						log.Error(err, "fail to FINALIZE MongoDB CR for Metering")
						return err
					}
				}
			} else {
				log.V(5).Info("Instance IS in deletion state")
				err := deleteSecrets(instance, client)
				if err != nil {
					log.Error(err, "fail to delete MongoDB Secrets for Metering")
					return err
				}
				err = mongodb.Finalize(instance, meteringMongoDBCR, client)
				if err != nil {
					log.Error(err, "fail to FINALIZE MongoDB CR for Metering")
					return err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		log.V(5).Info("MongoDB CR for Metering DOES exist")
		if foundMeteringMongoDBCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("MongoDB CR for Metering IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.EndpointMeteringConfig.Enabled {
				log.Info("Instance IS NOT in deletion state and Metering ENABLED")
				err := mongodb.Update(instance, meteringMongoDBCR, foundMeteringMongoDBCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE MongoDB CR for Metering")
					return err
				}
			} else {
				log.V(5).Info("Instance IS in deletion state or Metering DISABLED")
				if foundMeteringMongoDBCR.GetDeletionTimestamp() == nil {
					err := deleteClusterCertificate(instance, client)
					if err != nil {
						log.Error(err, "Fail to DELETE Certificate for Metering")
						return err
					}
					err = deleteClusterIssuer(instance, client)
					if err != nil {
						log.Error(err, "Fail to DELETE Cluster Issuer for Metering")
						return err
					}
					err = mongodb.Delete(foundMeteringMongoDBCR, client)
					if err != nil {
						log.Error(err, "Fail to DELETE MongoDB CR for Metering")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled Metering MongoDB")
	return nil
}

func createRootCACert(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	certificate := &certmanagerv1alpha1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-metering-ca-cert",
			Namespace: instance.Namespace,
		},
		Spec: certmanagerv1alpha1.CertificateSpec{
			CommonName: instance.Name + "-metering",
			IssuerRef: certmanagerv1alpha1.ObjectReference{
				Name: instance.Name + "-self-signed",
				Kind: "ClusterIssuer",
			},
			SecretName:   instance.Name + "-metering-ca-cert",
			IsCA:         true,
			Organization: []string{"IBM"},
		},
	}
	err := controllerutil.SetControllerReference(instance, certificate, scheme)
	if err != nil {
		return err
	}

	foundCertificate := &certmanagerv1alpha1.Certificate{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: certificate.Name, Namespace: certificate.Namespace}, foundCertificate)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Metering CA Certificate")
		return client.Create(context.TODO(), certificate)
	}

	return err
}

func createClusterIssuer(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: instance.Name + "-metering",
		},
		Spec: certmanagerv1alpha1.IssuerSpec{
			IssuerConfig: certmanagerv1alpha1.IssuerConfig{
				CA: &certmanagerv1alpha1.CAIssuer{
					SecretName: instance.Name + "-metering-ca-cert",
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(instance, clusterIssuer, scheme)
	if err != nil {
		return err
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Metering ClusterIssuer")
		return client.Create(context.TODO(), clusterIssuer)
	}

	return err
}

func newMeteringMongoDBCR(instance *multicloudv1beta1.Endpoint) (*multicloudv1beta1.MongoDB, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	curlImage, err := instance.GetImage("curl")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "curl")
		return nil, err
	}
	mongodbImage, err := instance.GetImage("mongodb")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "mongodb")
		return nil, err
	}
	installImage, err := instance.GetImage("mongodb-install")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "mongodb-install")
		return nil, err
	}

	return &multicloudv1beta1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-metering-mongodb",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.MongoDBSpec{
			FullNameOverride: instance.Name + "-metering-mongodb",
			Auth:             multicloudv1beta1.MongoDBSpecAuth{Enabled: true},
			Curl:             multicloudv1beta1.MongoDBSpecCurl{Image: curlImage},
			Image:            mongodbImage,
			InstallImage:     installImage,
			Metrics:          multicloudv1beta1.MongoDBSpecMetrics{Enabled: false},
			PersistentVolume: multicloudv1beta1.MongoDBSpecPersistentVolume{
				AccessModes: []string{"ReadWriteOnce"},
				Size:        "10Gi",
				Enabled:     true,
			},
			Replicas: 1,
			TLS: multicloudv1beta1.MongoDBSpecTLS{
				CASecret:   instance.Name + "-metering-ca-cert",
				Issuer:     instance.Name + "-metering",
				IssuerKind: "ClusterIssuer",
				Enabled:    true,
			},
			ImagePullSecrets: []string{instance.Spec.ImagePullSecret},
			LivenessProbe: multicloudv1beta1.MongoDBSpecProbe{
				FailureThreshold:    60,
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      5,
			},
			ReadinessProbe: multicloudv1beta1.MongoDBSpecProbe{
				FailureThreshold:    60,
				InitialDelaySeconds: 30,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      5,
			},
			Resources: multicloudv1beta1.MongoDBSpecResources{
				Limits: multicloudv1beta1.MongoDBSpecResourcesLimit{
					Memory: "4Gi",
				},
				Requests: multicloudv1beta1.MongoDBSpecResourcesRequest{
					Memory: "4Gi",
				},
			},
			NodeSelectorEnabled:       false,
			PriorityClassNameEnabled:  false,
			ServiceAccountNameEnabled: true,
			ClusterRoleEnabled:        true,
		},
	}, nil
}

func deleteClusterCertificate(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	foundClusterIssuer := &certmanagerv1alpha1.Certificate{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-metering-ca-cert", Namespace: instance.Namespace}, foundClusterIssuer)
	if err == nil {
		log.Info("Deleting Metering Certificate")
		return client.Delete(context.TODO(), foundClusterIssuer)
	}

	return err
}

func deleteClusterIssuer(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-metering"}, foundClusterIssuer)
	if err == nil {
		log.Info("Deleting Metering ClusterIssuer")
		return client.Delete(context.TODO(), foundClusterIssuer)
	}

	return err
}

func deleteSecrets(instance *multicloudv1beta1.Endpoint, client client.Client) error {
	secretsToDelete := []string{
		instance.Name + "-metering-mongodb-client-cert",
		instance.Name + "-metering-ca-cert",
	}

	for _, secretToDelete := range secretsToDelete {
		foundSecretToDelete := &corev1.Secret{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: secretToDelete, Namespace: instance.Namespace}, foundSecretToDelete)
		if err == nil {
			err := client.Delete(context.TODO(), foundSecretToDelete)
			if err != nil {
				log.Error(err, "Fail to DELETE MongoDB Secrets for Metering", "Secret.Name", secretToDelete)
				return err
			}
		}
	}
	return nil
}
