//Package v1beta1 of certmgr Defines the Reconciliation logic and required setup for component operator.
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"context"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	openshiftsecurityv1 "github.com/openshift/api/security/v1"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// TODO(liuhao): switch from klusterletv1alpha1.CertManager to multicloudv1beta1.CertManager

var log = logf.Log.WithName("certmgr")

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) error {
	reqLogger := log.WithValues("KlusterletService.Namespace", instance.Namespace, "KlusterletService.Name", instance.Name)
	reqLogger.Info("Reconciling CertManager")

	var err error

	// ICP CertManager
	log.V(5).Info("Looking for ICP CertManager Deployment", "Deployment.Name", "cert-manager-ibm-cert-manager", "Deployment.Namespace", "cert-manager")
	findICPCertMgr := &extensionsv1beta1.Deployment{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: "cert-manager-ibm-cert-manager", Namespace: "cert-manager"}, findICPCertMgr)
	if err == nil {
		err = createSelfSignClusterIssuer(client, scheme, instance)
		if err != nil {
			log.Error(err, "Unable to CREATE SelfSigned ClusterIssuer.")
			return err
		}

		log.V(1).Info("Found ICP CertManager, skip CertManagerCR Reconcile.")
		return nil
	}

	// No ICP CertManager
	certMgr, err := newCertManagerCR(instance)
	if err != nil {
		log.Error(err, "Fail to generate desired CertManager CR")
		return err
	}

	err = controllerutil.SetControllerReference(instance, certMgr, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return err
	}

	foundCertManager := &klusterletv1alpha1.CertManager{}
	log.V(5).Info("Looking for CertManager CR", "CertManager.Name", certMgr.Name, "CertManager.Namespace", certMgr.Namespace)
	err = client.Get(context.TODO(), types.NamespacedName{Name: certMgr.Name, Namespace: certMgr.Namespace}, foundCertManager)
	if err != nil {
		if errors.IsNotFound(err) {
			// CertManager CR does NOT exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				err = createServiceAccount(client, scheme, instance, certMgr)
				if err != nil {
					log.Error(err, "Fail to CREATE ServiceAccount")
					return err
				}

				log.Info("Creating a new CertManager CR", "CertManager.Namespace", certMgr.Namespace, "CertManager.Name", certMgr.Name)
				err = client.Create(context.TODO(), certMgr)
				if err != nil {
					log.Error(err, "Fail to CREATE CertManager CR")
					return err
				}

				// Create SelfSigned ClusterIssuer
				err = createSelfSignClusterIssuer(client, scheme, instance)
				if err != nil {
					log.Error(err, "Fail to CREATE SelfSigned ClusterIssuer")
					return err
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
								return err
							}
						}

						// Delete SelfSigned ClusterIssuer
						err = deleteSelfSignClusterIssuer(client, scheme, instance)
						if err != nil {
							log.Error(err, "Fail to DELETE SelfSigned ClusterIssuer")
							return err
						}

						instance.Finalizers = append(instance.Finalizers[0:i], instance.Finalizers[i+1:]...)
						break
					}
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return err
		}
	} else {
		if foundCertManager.GetDeletionTimestamp() == nil {
			// CertManager CR does exist
			if instance.GetDeletionTimestamp() == nil {
				// KlusterletService NOT in deletion state
				foundCertManager.Spec = certMgr.Spec
				err = client.Update(context.TODO(), foundCertManager)
				if err != nil && !errors.IsConflict(err) {
					log.Error(err, "Fail to UPDATE CertManager CR")
					return err
				}
			} else {
				// KlusterletService in deletion state
				if foundCertManager.GetDeletionTimestamp() == nil {
					// Delete CertManager CR
					err = client.Delete(context.TODO(), foundCertManager)
					if err != nil {
						log.Error(err, "Fail to DELETE CertManager CR")
						return err
					}
				}
			}
		}
	}

	reqLogger.Info("Successfully Reconciled CertManager")
	return nil
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
	err := controllerutil.SetControllerReference(cr, clusterIssuer, scheme)
	if err != nil {
		return err
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer)

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
	err := controllerutil.SetControllerReference(cr, clusterIssuer, scheme)
	if err != nil {
		return err
	}

	foundClusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: clusterIssuer.Name}, foundClusterIssuer)

	if err == nil {
		log.Info("Deleting SelfSigned ClusterIssuer")
		return client.Delete(context.TODO(), foundClusterIssuer)
	}

	return nil
}

func createServiceAccount(client client.Client, scheme *runtime.Scheme, instance *multicloudv1beta1.Endpoint, certmgr *klusterletv1alpha1.CertManager) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certmgr.Spec.ServiceAccount.Name,
			Namespace: certmgr.Namespace,
		},
		ImagePullSecrets: []corev1.LocalObjectReference{
			corev1.LocalObjectReference{
				Name: instance.Spec.ImagePullSecret,
			},
		},
	}
	err := controllerutil.SetControllerReference(instance, serviceAccount, scheme)
	if err != nil {
		return err
	}

	foundServiceAccount := &corev1.ServiceAccount{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ServiceAccount", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
		err = client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	foundPrivilegedSCC := &openshiftsecurityv1.SecurityContextConstraints{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: "privileged", Namespace: ""}, foundPrivilegedSCC)
	// if client.Get return error that means no privileged SCC in that case skip adding user to scc and ignore error
	if err == nil {
		user := "system:serviceaccount:" + serviceAccount.Namespace + ":" + serviceAccount.Name
		log.Info("Adding User to SCC", "User", user, "SCC", foundPrivilegedSCC.Name)
		foundPrivilegedSCC.Users = append(foundPrivilegedSCC.Users, user)
		err = client.Update(context.TODO(), foundPrivilegedSCC)
		if err != nil {
			return err
		}
	}

	return nil
}

func newCertManagerCR(cr *multicloudv1beta1.Endpoint) (*klusterletv1alpha1.CertManager, error) {
	labels := map[string]string{
		"app": cr.Name,
	}

	image, err := cr.GetImage("cert-manager-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cert-manager-controller")
		return nil, err
	}

	return &klusterletv1alpha1.CertManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-certmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.CertManagerSpec{
			FullNameOverride:         cr.Name + "-certmgr",
			ClusterResourceNamespace: cr.Namespace,
			ServiceAccount: klusterletv1alpha1.CertManagerServiceAccount{
				Name: cr.Name + "-certmgr",
			},
			Image: image,
		},
	}, nil
}
