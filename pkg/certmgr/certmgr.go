package certmgr

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	openshiftsecurityv1 "github.com/openshift/api/security/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("certmgr")

func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	certMgr := newCertManagerCR(instance)
	if err := controllerutil.SetControllerReference(instance, certMgr, scheme); err != nil {
		return err
	}

	foundCertManager := &klusterletv1alpha1.CertManager{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: certMgr.Name, Namespace: certMgr.Namespace}, foundCertManager)
	if err != nil && errors.IsNotFound(err) {
		preCreate(client, instance, certMgr)
		log.Info("Creating a new CertManager", "CertManager.Namespace", certMgr.Namespace, "CertManager.Name", certMgr.Name)
		if err := client.Create(context.TODO(), certMgr); err != nil {
			return err
		}
		if err := createSelfSignIssuer(client, "self-signed", ""); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func createSelfSignIssuer(client client.Client, name string, namespace string) error {
	if namespace == "" {
		clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ""}, clusterIssuer)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating self signed cluster issuer", "Name", name)
			clusterIssuer = &certmanagerv1alpha1.ClusterIssuer{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: certmanagerv1alpha1.IssuerSpec{
					IssuerConfig: certmanagerv1alpha1.IssuerConfig{
						SelfSigned: &certmanagerv1alpha1.SelfSignedIssuer{},
					},
				},
			}
			if err := client.Create(context.TODO(), clusterIssuer); err != nil {
				return err
			}
		}
	} else {
		issuer := &certmanagerv1alpha1.Issuer{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, issuer)
		if err != nil && errors.IsNotFound(err) {
			log.Info("Creating self signed cluster issuer", "Name", name)
			issuer = &certmanagerv1alpha1.Issuer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: certmanagerv1alpha1.IssuerSpec{
					IssuerConfig: certmanagerv1alpha1.IssuerConfig{
						SelfSigned: &certmanagerv1alpha1.SelfSignedIssuer{},
					},
				},
			}
			if err := client.Create(context.TODO(), issuer); err != nil {
				return err
			}
		}
	}
	return nil
}

func preCreate(client client.Client, instance *klusterletv1alpha1.KlusterletService, certMgr *klusterletv1alpha1.CertManager) error {
	if err := installCertificateCRD(client); err != nil {
		return err
	}
	if err := installIssuerCRD(client); err != nil {
		return err
	}
	if err := installClusterIssuerCRD(client); err != nil {
		return err
	}
	if err := installOrderCRD(client); err != nil {
		return err
	}
	if err := installChallengeCRD(client); err != nil {
		return err
	}
	certMgrSA, err := getOrCreateServiceAccount(client, certMgr.Spec.ServiceAccount.Name, instance.Namespace)
	if err != nil {
		return err
	}
	privilegedSCC := &openshiftsecurityv1.SecurityContextConstraints{}
	if err := client.Get(context.TODO(), types.NamespacedName{Name: "privileged", Namespace: ""}, privilegedSCC); err != nil {
		return err
	}
	if err := addServiceAccountToSCC(client, certMgrSA, privilegedSCC); err != nil {
		return err
	}
	return nil
}

func newCertManagerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.CertManager {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.CertManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-cert-manager",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.CertManagerSpec{
			ClusterResourceNamespace: cr.Namespace,
			ServiceAccount: klusterletv1alpha1.CertManagerServiceAccount{
				Create: false,
				Name:   cr.Name + "-cert-manager",
			},
			FullNameOverride: cr.Name + "-cert-manager",
		},
	}
}

func installCertificateCRD(client client.Client) error {
	log.Info("Installing certificates.certmanager.k8s.io CRD")

	certificatesCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "certificates.certmanager.k8s.io", Namespace: ""}, certificatesCRD)
	if err != nil && errors.IsNotFound(err) {
		certificatesCRD = &apiextensionv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "certificates.certmanager.k8s.io",
			},
			Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
				Scope:   "Namespaced",
				Group:   "certmanager.k8s.io",
				Version: "v1alpha1",
				Names: apiextensionv1beta1.CustomResourceDefinitionNames{
					Kind:       "Certificate",
					Plural:     "certificates",
					ShortNames: []string{"cert", "certs"},
				},
				AdditionalPrinterColumns: []apiextensionv1beta1.CustomResourceColumnDefinition{
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Ready",
						Type:     "string",
						JSONPath: ".status.conditions[?(@.type==\"Ready\")].status",
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Secret",
						Type:     "string",
						JSONPath: ".spec.secretName",
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Issuer",
						Type:     "string",
						JSONPath: ".spec.issuerRef.name",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Status",
						Type:     "string",
						JSONPath: ".status.conditions[?(@.type==\"Ready\")].message",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Age",
						Type:     "date",
						JSONPath: ".metadata.creationTimestamp",
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Expiration",
						Type:     "string",
						JSONPath: ".status.notAfter",
					},
				},
			},
		}
		err = client.Create(context.TODO(), certificatesCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installIssuerCRD(client client.Client) error {
	log.Info("Installing issuers.certmanager.k8s.io CRD")

	issuerCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "issuers.certmanager.k8s.io", Namespace: ""}, issuerCRD)
	if err != nil && errors.IsNotFound(err) {
		issuerCRD = &apiextensionv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "issuers.certmanager.k8s.io",
			},
			Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
				Scope:   "Namespaced",
				Group:   "certmanager.k8s.io",
				Version: "v1alpha1",
				Names: apiextensionv1beta1.CustomResourceDefinitionNames{
					Kind:   "Issuer",
					Plural: "issuers",
				},
			},
		}
		err = client.Create(context.TODO(), issuerCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installClusterIssuerCRD(client client.Client) error {
	log.Info("Installing clusterissuers.certmanager.k8s.io CRD")

	clusterIssuerCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "clusterissuers.certmanager.k8s.io", Namespace: ""}, clusterIssuerCRD)
	if err != nil && errors.IsNotFound(err) {
		clusterIssuerCRD = &apiextensionv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "clusterissuers.certmanager.k8s.io",
			},
			Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
				Scope:   "Cluster",
				Group:   "certmanager.k8s.io",
				Version: "v1alpha1",
				Names: apiextensionv1beta1.CustomResourceDefinitionNames{
					Kind:   "ClusterIssuer",
					Plural: "clusterissuers",
				},
			},
		}
		err = client.Create(context.TODO(), clusterIssuerCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installOrderCRD(client client.Client) error {
	log.Info("Installing orders.certmanager.k8s.io CRD")

	orderCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "orders.certmanager.k8s.io", Namespace: ""}, orderCRD)
	if err != nil && errors.IsNotFound(err) {
		orderCRD = &apiextensionv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "orders.certmanager.k8s.io",
			},
			Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
				Scope:   "Namespaced",
				Group:   "certmanager.k8s.io",
				Version: "v1alpha1",
				Names: apiextensionv1beta1.CustomResourceDefinitionNames{
					Kind:   "Order",
					Plural: "orders",
				},
				AdditionalPrinterColumns: []apiextensionv1beta1.CustomResourceColumnDefinition{
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "State",
						Type:     "string",
						JSONPath: ".status.state",
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Issuer",
						Type:     "string",
						JSONPath: ".spec.issuerRef.name",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Reason",
						Type:     "string",
						JSONPath: ".status.reason",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Age",
						Type:     "date",
						JSONPath: ".metadata.creationTimestamp",
					},
				},
			},
		}
		err = client.Create(context.TODO(), orderCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installChallengeCRD(client client.Client) error {
	log.Info("Installing challenges.certmanager.k8s.io CRD")

	challengeCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "challenges.certmanager.k8s.io", Namespace: ""}, challengeCRD)
	if err != nil && errors.IsNotFound(err) {
		challengeCRD = &apiextensionv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "challenges.certmanager.k8s.io",
			},
			Spec: apiextensionv1beta1.CustomResourceDefinitionSpec{
				Scope:   "Namespaced",
				Group:   "certmanager.k8s.io",
				Version: "v1alpha1",
				Names: apiextensionv1beta1.CustomResourceDefinitionNames{
					Kind:   "Challenge",
					Plural: "challenges",
				},
				AdditionalPrinterColumns: []apiextensionv1beta1.CustomResourceColumnDefinition{
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "State",
						Type:     "string",
						JSONPath: ".status.state",
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Domain",
						Type:     "string",
						JSONPath: "..spec.dnsName",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Reason",
						Type:     "string",
						JSONPath: ".status.reason",
						Priority: 1,
					},
					apiextensionv1beta1.CustomResourceColumnDefinition{
						Name:     "Age",
						Type:     "date",
						JSONPath: ".metadata.creationTimestamp",
					},
				},
			},
		}
		err = client.Create(context.TODO(), challengeCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

//NOTE: service account related methods may need to be refactored to another package
func getOrCreateServiceAccount(client client.Client, name string, namespace string) (*corev1.ServiceAccount, error) {
	log.Info("Get or create service account", "Name", name, "Namespace", namespace)

	serviceAccount := &corev1.ServiceAccount{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating service account", "Name", name, "Namespace", namespace)
		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		if err := client.Create(context.TODO(), serviceAccount); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	log.Info("Get or create service account", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
	return serviceAccount, nil
}

func addServiceAccountToSCC(client client.Client, sa *corev1.ServiceAccount, scc *openshiftsecurityv1.SecurityContextConstraints) error {
	user := "system:serviceaccount:" + sa.Namespace + ":" + sa.Name
	log.Info("Add ServiceAccount to SecurityContextConstraints", "user", user, "scc.Name", scc.Name)
	scc.Users = append(scc.Users, user)
	return client.Update(context.TODO(), scc)
}
