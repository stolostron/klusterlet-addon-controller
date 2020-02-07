// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1beta1 of certmgr Defines the Reconciliation logic and required setup for component operator.
package v1beta1

import (
	"context"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	openshiftsecurityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/utils"
)

var log = logf.Log.WithName("certmgr")

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling CertManager")

	var err error

	// ICP CertManager
	log.V(5).Info("Looking for ICP CertManager Deployment", "Deployment.Name", "cert-manager-ibm-cert-manager", "Deployment.Namespace", "cert-manager")
	findICPCertMgr := &appsv1.Deployment{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: "cert-manager-ibm-cert-manager", Namespace: "cert-manager"}, findICPCertMgr)
	if err == nil {
		err = createSelfSignClusterIssuer(client, scheme, instance)
		if err != nil {
			log.Error(err, "Unable to CREATE SelfSigned ClusterIssuer.")
			return false, err
		}

		log.V(1).Info("Found ICP CertManager, skip CertManagerCR Reconcile.")
		return false, nil
	}

	// No ICP CertManager
	certMgr, err := newCertManagerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired CertManager CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, certMgr, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundCertManager := &multicloudv1beta1.CertManager{}
	log.V(5).Info("Looking for CertManager CR", "CertManager.Name", certMgr.Name, "CertManager.Namespace", certMgr.Namespace)
	err = client.Get(context.TODO(), types.NamespacedName{Name: certMgr.Name, Namespace: certMgr.Namespace}, foundCertManager)
	if err != nil {
		if errors.IsNotFound(err) {
			// CertManager CR does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				err = createOrUpdateServiceAccount(client, scheme, instance, certMgr)
				if err != nil {
					log.Error(err, "Fail to CREATE ServiceAccount")
					return false, err
				}

				log.Info("Creating a new CertManager CR", "CertManager.Namespace", certMgr.Namespace, "CertManager.Name", certMgr.Name)
				err = client.Create(context.TODO(), certMgr)
				if err != nil {
					log.Error(err, "Fail to CREATE CertManager CR")
					return false, err
				}

				// Create SelfSigned ClusterIssuer
				err = createSelfSignClusterIssuer(client, scheme, instance)
				if err != nil {
					log.Error(err, "Fail to CREATE SelfSigned ClusterIssuer")
					return false, err
				}

				// Adding Finalizer to KlusterletService instance
				instance.Finalizers = append(instance.Finalizers, certMgr.Name)
			} else {
				// Remove finalizer
				for i, finalizer := range instance.Finalizers {
					if finalizer == certMgr.Name {
						// Delete cert-manager-controller ConfigMap
						foundConfigMap := &corev1.ConfigMap{}
						err = client.Get(context.TODO(), types.NamespacedName{Name: "cert-manager-controller", Namespace: certMgr.Namespace}, foundConfigMap)
						if err == nil {
							err = client.Delete(context.TODO(), foundConfigMap)
							if err != nil {
								log.Error(err, "Fail to DELETE ConnectionManager Secret", "Secret.Name", foundConfigMap)
								return false, err
							}
						}

						// Delete SelfSigned ClusterIssuer
						err = deleteSelfSignClusterIssuer(client, scheme, instance)
						if err != nil {
							log.Error(err, "Fail to DELETE SelfSigned ClusterIssuer")
							return false, err
						}

						instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
						return true, err
					}
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		if foundCertManager.GetDeletionTimestamp() == nil {
			// CertManager CR does exist
			if instance.GetDeletionTimestamp() == nil {
				// Endpoint NOT in deletion state
				err = createOrUpdateServiceAccount(client, scheme, instance, certMgr)
				if err != nil {
					log.Error(err, "Fail to Update ServiceAccount")
					return false, err
				}
				foundCertManager.Spec = certMgr.Spec
				err = client.Update(context.TODO(), foundCertManager)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE CertManager CR")
					return false, err
				}
				// Create SelfSigned ClusterIssuer
				// TODO(liuhao): figure out why selfsigned cluster issuer is being deleted
				err = createSelfSignClusterIssuer(client, scheme, instance)
				if err != nil {
					log.Error(err, "Fail to CREATE SelfSigned ClusterIssuer")
					return false, err
				}
			} else {
				// Endpoint in deletion state
				// Delete CertManager CR
				err = client.Delete(context.TODO(), foundCertManager)
				if err != nil {
					log.Error(err, "Fail to DELETE CertManager CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for CertManager")
				return true, err
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for CertManager")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled CertManager")
	return false, nil
}

func createSelfSignClusterIssuer(client client.Client, scheme *runtime.Scheme, cr *multicloudv1beta1.Endpoint) error {
	clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name + "-self-signed",
		},
		Spec: certmanagerv1alpha1.IssuerSpec{
			IssuerConfig: certmanagerv1alpha1.IssuerConfig{
				SelfSigned: &certmanagerv1alpha1.SelfSignedIssuer{},
			},
		},
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating SelfSigned ClusterIssuer")
		return client.Create(context.TODO(), clusterIssuer)
	}

	return err
}

func deleteSelfSignClusterIssuer(client client.Client, scheme *runtime.Scheme, cr *multicloudv1beta1.Endpoint) error {
	clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: cr.Name + "-self-signed",
		},
		Spec: certmanagerv1alpha1.IssuerSpec{
			IssuerConfig: certmanagerv1alpha1.IssuerConfig{
				SelfSigned: &certmanagerv1alpha1.SelfSignedIssuer{},
			},
		},
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer); err == nil {
		log.Info("Deleting SelfSigned ClusterIssuer")
		return client.Delete(context.TODO(), foundClusterIssuer)
	}

	return nil
}

func createOrUpdateServiceAccount(client client.Client, scheme *runtime.Scheme, instance *multicloudv1beta1.Endpoint, certmgr *multicloudv1beta1.CertManager) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certmgr.Spec.ServiceAccount.Name,
			Namespace: certmgr.Namespace,
		},
	}

	imagePullSecret := corev1.LocalObjectReference{Name: instance.Spec.ImagePullSecret}

	err := controllerutil.SetControllerReference(instance, serviceAccount, scheme)
	if err != nil {
		return err
	}

	foundServiceAccount := &corev1.ServiceAccount{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		log.Info("Creating ServiceAccount", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
		if instance.Spec.ImagePullSecret != "" {
			serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, imagePullSecret)
		}
		if err := client.Create(context.TODO(), serviceAccount); err != nil {
			return err
		}

		// if client.Get return error that means no privileged SCC in that case skip adding user to scc and ignore error
		foundPrivilegedSCC := &openshiftsecurityv1.SecurityContextConstraints{}
		if err := client.Get(context.TODO(), types.NamespacedName{Name: "privileged", Namespace: ""}, foundPrivilegedSCC); err == nil {
			user := "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name
			log.Info("Adding User to SCC", "User", user, "SCC", foundPrivilegedSCC.Name)
			foundPrivilegedSCC.Users = append(foundPrivilegedSCC.Users, user)
			foundPrivilegedSCC.Users = utils.UniqueStringSlice(foundPrivilegedSCC.Users)
			if err := client.Update(context.TODO(), foundPrivilegedSCC); err != nil {
				return err
			}
		}
	} else {
		if instance.Spec.ImagePullSecret != "" {
			foundImagePullSecret := false
			for _, secret := range foundServiceAccount.ImagePullSecrets {
				if secret.Name == imagePullSecret.Name {
					foundImagePullSecret = true
					break
				}
			}

			if !foundImagePullSecret {
				foundServiceAccount.ImagePullSecrets = append(foundServiceAccount.ImagePullSecrets, imagePullSecret)
				log.Info("Updating ServiceAccount", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
				err = client.Update(context.TODO(), foundServiceAccount)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func newCertManagerCR(cr *multicloudv1beta1.Endpoint) (*multicloudv1beta1.CertManager, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	image, err := cr.GetImage("cert-manager-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cert-manager-controller")
		return nil, err
	}

	return &multicloudv1beta1.CertManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-certmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.CertManagerSpec{
			FullNameOverride:         cr.Name + "-certmgr",
			ClusterResourceNamespace: cr.Namespace,
			ServiceAccount: multicloudv1beta1.CertManagerServiceAccount{
				Name: cr.Name + "-certmgr",
			},
			Image: image,
			PolicyController: multicloudv1beta1.CertManagerPolicyControllerSpec{
				Enabled: false,
			},
		},
	}, nil
}
