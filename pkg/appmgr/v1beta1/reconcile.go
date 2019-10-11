// Package v1beta1 of appmgr provides a reconciler for the ApplicationManager
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	multicloudv1beta1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/multicloud/v1beta1"
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/inspect"
	tiller "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/tiller/v1beta1"
)

// Reconcile Resolves differences in the running state of the cert-manager services and CRDs.
func Reconcile(instance *multicloudv1beta1.Endpoint, client client.Client, scheme *runtime.Scheme) (bool, error) {
	reqLogger := log.WithValues("Endpoint.Namespace", instance.Namespace, "Endpoint.Name", instance.Name)
	reqLogger.Info("Reconciling ApplicationManager")

	appMgrCR, err := newApplicationManagerCR(instance, client)
	if err != nil {
		log.Error(err, "Fail to generate desired ApplicationManager CR")
		return false, err
	}

	err = controllerutil.SetControllerReference(instance, appMgrCR, scheme)
	if err != nil {
		log.Error(err, "Unable to SetControllerReference")
		return false, err
	}

	foundAppMgrCR := &multicloudv1beta1.ApplicationManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: appMgrCR.Name, Namespace: appMgrCR.Namespace}, foundAppMgrCR)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.V(5).Info("ApplicationManager CR DOES NOT exist")
			if instance.GetDeletionTimestamp() == nil {
				log.V(5).Info("instance IS NOT in deletion state")
				if instance.Spec.ApplicationManagerConfig.Enabled {
					log.V(5).Info("ApplicationManager ENABLED")

					caBundle, err := checkAndGenerateSecret(instance, client)
					if err != nil {
						return false, err
					}
					appMgrCR.Spec.HelmCRDAdmissionControllerSpec.CABundle = caBundle
					if err = create(instance, appMgrCR, client); err != nil {
						log.Error(err, "fail to CREATE ApplicationManager CR")
						return false, err
					}
				} else {
					log.V(5).Info("ApplicationManager DISABLED")
					if err = finalize(instance, appMgrCR, client); err != nil {
						log.Error(err, "fail to FINALIZE ApplicationManager CR")
						return false, err
					}
				}
			} else {
				log.V(5).Info("instance IS in deletion state")
				if err = finalize(instance, appMgrCR, client); err != nil {
					log.Error(err, "fail to FINALIZE ApplicationManager CR")
					return false, err
				}
			}
		} else {
			log.Error(err, "Unexpected ERROR")
			return false, err
		}
	} else {
		log.V(5).Info("ApplicationManager CR DOES exist")
		if foundAppMgrCR.GetDeletionTimestamp() == nil {
			log.V(5).Info("ApplicationManager CR IS NOT in deletion state")
			if instance.GetDeletionTimestamp() == nil && instance.Spec.ApplicationManagerConfig.Enabled {
				log.Info("instance IS NOT in deletion state and ApplicationManager ENABLED")

				if !appMgrCR.Spec.TillerIntegration.Enabled {
					foundAppMgrDeploy := &appsv1.Deployment{}
					err = client.Get(context.TODO(), types.NamespacedName{Name: "endpoint-appmgr-helm-crd", Namespace: appMgrCR.Namespace}, foundAppMgrDeploy)
					if err != nil {
						if kerrors.IsNotFound(err) {
							log.Info("Application Manger deploy not found. Clean up CRs")

							foundHelmCRD := &crdv1beta1.CustomResourceDefinition{}
							err := client.Get(context.TODO(), types.NamespacedName{Name: "helmreleases.helm.bitnami.com", Namespace: ""}, foundHelmCRD)
							if err != nil {
								if kerrors.IsNotFound(err) {
									log.Info("HelmCRD not found, skipping delete")
								} else {
									log.Error(err, "Unexpected ERROR")
									return false, err
								}
							} else {
								err = cleanUpHelmCRs(client)
								if err != nil {
									log.Error(err, "Failed to clean up Helm CRs")
									return false, err
								}
							}

							err = cleanUpSecret(instance, client, appMgrCR)
							if err != nil {
								return false, err
							}
						} else {
							log.Error(err, "Unexpected ERROR")
							return false, err
						}
					} else {
						log.Info("Application Manager deploy still exists, trying clean up later")
					}
				} else {
					caBundle, err := checkAndGenerateSecret(instance, client)
					if err != nil {
						return false, err
					}
					appMgrCR.Spec.HelmCRDAdmissionControllerSpec.CABundle = caBundle
					err = tiller.CheckDependency(instance, client, foundAppMgrCR.Name)
					if err != nil {
						log.Error(err, "fail to check dependency for ApplicationManager CR")
						return false, err
					}
				}
				err = update(instance, appMgrCR, foundAppMgrCR, client)
				if err != nil {
					log.Error(err, "fail to UPDATE ApplicationManager CR")
					return false, err
				}
			} else {
				log.V(5).Info("instance IS in deletion state or ApplicationManager DISABLED")
				if err = delete(foundAppMgrCR, client); err != nil {
					log.Error(err, "Fail to DELETE ApplicationManager CR")
					return false, err
				}
				reqLogger.Info("Requeueing Reconcile for ApplicationManager")
				return true, nil
			}
		} else {
			reqLogger.Info("Requeueing Reconcile for ApplicationManager")
			return true, nil
		}
	}

	reqLogger.Info("Successfully Reconciled ApplicationManager")
	return false, nil
}

func newApplicationManagerTillerIntegration(cr *multicloudv1beta1.Endpoint, client client.Client) multicloudv1beta1.ApplicationManagerTillerIntegration {
	if cr.Spec.TillerIntegration.Enabled {
		// ICP Tiller
		icpTillerServiceEndpoint := tiller.GetICPTillerServiceEndpoint(client)
		if icpTillerServiceEndpoint != "" {
			return multicloudv1beta1.ApplicationManagerTillerIntegration{
				Enabled:       true,
				Endpoint:      icpTillerServiceEndpoint,
				CertIssuer:    "icp-ca-issuer",
				AutoGenSecret: true,
				User:          tiller.GetICPTillerDefaultAdminUser(client),
			}
		}

		// KlusterletOperator deployed Tiller
		return multicloudv1beta1.ApplicationManagerTillerIntegration{
			Enabled:       true,
			Endpoint:      cr.Name + "-tiller" + ":44134",
			CertIssuer:    cr.Name + "-tiller",
			AutoGenSecret: true,
			User:          cr.Name + "-admin",
		}
	}

	return multicloudv1beta1.ApplicationManagerTillerIntegration{
		Enabled: false,
	}
}

func newApplicationManagerCR(instance *multicloudv1beta1.Endpoint, client client.Client) (*multicloudv1beta1.ApplicationManager, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	deployableImage, err := instance.GetImage("deployable")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "deployable")
		return nil, err
	}

	subscriptionImage, err := instance.GetImage("subscription")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "subscription")
		return nil, err
	}

	helmCRDImage, err := instance.GetImage("helmcrd")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "helmcrd")
		return nil, err
	}

	helmCRDAdmissionControllerImage, err := instance.GetImage("helmcrd-admission-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "helmcrd-admission-controller")
		return nil, err
	}

	clusterCADomain := "mycluster.icp"
	ipStr := "0.0.0.0"
	if inspect.Info.KubeVendor == inspect.KubeVendorICP {
		clusterInfoCM := &corev1.ConfigMap{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: "ibmcloud-cluster-info", Namespace: "kube-public"}, clusterInfoCM)
		if err != nil {
			log.Error(err, "Unexpected ERROR")
			return nil, err
		}

		clusterCADomain = clusterInfoCM.Data["cluster_ca_domain"]
		clusterAddress := clusterInfoCM.Data["cluster_address"]

		ip := net.ParseIP(clusterAddress)
		ipStr = ip.String()
		if ip == nil {
			log.Info("cluster_address is a FQDN, looking up the IP")
			ipArr, err := net.LookupIP(clusterAddress)
			if err != nil {
				log.Error(err, "Failed to look up IP for cluster_address")
				return nil, err
			}
			if len(ipArr) > 0 {
				ipStr = ipArr[0].String()
			} else {
				log.Info("Could not resolve IPs for hostname, using default value")
			}
		}
	}

	return &multicloudv1beta1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-appmgr",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: multicloudv1beta1.ApplicationManagerSpec{
			FullNameOverride:  instance.Name + "-appmgr",
			ConnectionManager: instance.Name + "-connmgr",
			ClusterName:       instance.Spec.ClusterName,
			ClusterNamespace:  instance.Spec.ClusterNamespace,
			TillerIntegration: newApplicationManagerTillerIntegration(instance, client),
			DeployableSpec: multicloudv1beta1.ApplicationManagerDeployableSpec{
				Image: deployableImage,
			},
			SubscriptionSpec: multicloudv1beta1.ApplicationManagerSubscriptionSpec{
				Image: subscriptionImage,
			},
			HelmCRDSpec: multicloudv1beta1.ApplicationManagerHelmCRDSpec{
				Image:    helmCRDImage,
				Hostname: clusterCADomain,
				IP:       ipStr,
			},
			HelmCRDAdmissionControllerSpec: multicloudv1beta1.ApplicationManagerHelmCRDAdmissionControllerSpec{
				Image: helmCRDAdmissionControllerImage,
			},
			ImagePullSecret: instance.Spec.ImagePullSecret,
		},
	}, nil
}

func generateCert(instance *multicloudv1beta1.Endpoint, client client.Client) ([]byte, []byte, []byte, error) {
	log.Info("Generating CA")
	cn := "helm-crd-admission-controller-svc"
	// Create Certificate Authority
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization: []string{"helm-crd"},
			CommonName:   cn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error(err, "Failed to generate CA key")
		return nil, nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Error(err, "Failed to generate CA cert")
		return nil, nil, nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"helm-crd"},
			CommonName:   cn,
		},
		DNSNames: []string{
			cn,
			cn + "." + instance.Namespace,
			cn + "." + instance.Namespace + ".svc",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error(err, "Failed to generate server key")
		return nil, nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		log.Error(err, "Failed to generate server cert using ca")
		return nil, nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	return caPEM.Bytes(), certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

func secretExists(instance *multicloudv1beta1.Endpoint, client client.Client, secretName string) (*corev1.Secret, error) {
	log.Info("Check secret")
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, secret)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info("Secret not found")
			return nil, nil
		}
		log.Error(err, "Unexpected ERROR")
		return nil, err
	}

	log.Info("Found secret " + secret.Name)
	return secret, nil
}

func checkAndGenerateSecret(instance *multicloudv1beta1.Endpoint, client client.Client) (string, error) {
	foundSecret, err := secretExists(instance, client, "helm-crd-admission-controller-certs")
	if err != nil {
		log.Error(err, "Unexpected ERROR")
		return "", err
	}

	if foundSecret == nil {
		caCertBytes, serverCertBytes, serverKeyBytes, err := generateCert(instance, client)
		if err != nil {
			return "", err
		}

		newSecret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "helm-crd-admission-controller-certs",
				Namespace: instance.Namespace,
			},
			Data: map[string][]byte{
				"ca.crt":  caCertBytes,
				"tls.crt": serverCertBytes,
				"tls.key": serverKeyBytes,
			},
			Type: corev1.SecretTypeOpaque,
		}

		err = client.Create(context.TODO(), newSecret)
		if err != nil {
			return "", err
		}

		caCertBase64 := base64.StdEncoding.EncodeToString(caCertBytes)
		return caCertBase64, nil
	}

	foundCACertBase64 := base64.StdEncoding.EncodeToString(foundSecret.Data["ca.crt"])
	return foundCACertBase64, nil
}

func cleanUpSecret(instance *multicloudv1beta1.Endpoint, client client.Client, cr *multicloudv1beta1.ApplicationManager) error {
	foundSecret, err := secretExists(instance, client, "helm-crd-admission-controller-certs")
	if err != nil {
		log.Error(err, "Unexpected ERROR")
		return err
	}

	if foundSecret != nil {
		log.Info("Deleting secret " + foundSecret.Name)
		err = client.Delete(context.TODO(), foundSecret)
		if err != nil {
			log.Error(err, "Failed to DELETE secret")
			return err
		}
	}

	foundSecret, err = secretExists(instance, client, cr.Spec.FullNameOverride+"-tiller-client-certs")
	if err != nil {
		log.Error(err, "Unexpected ERROR")
		return err
	}

	if foundSecret != nil {
		log.Info("Deleting secret " + foundSecret.Name)
		err = client.Delete(context.TODO(), foundSecret)
		if err != nil {
			log.Error(err, "Failed to DELETE secret")
			return err
		}
	}

	return nil
}
