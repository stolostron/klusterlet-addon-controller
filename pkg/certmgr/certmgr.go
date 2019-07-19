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
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("certmgr")

func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	// ICP CertManager
	findICPCertMgr := &extensionsv1beta1.Deployment{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "cert-manager-ibm-cert-manager", Namespace: "cert-manager"}, findICPCertMgr)
	if err == nil {
		err = createSelfSignClusterIssuer(client, scheme, instance)
		if err != nil {
			return nil
		}

		log.Info("Found ICP CertManager, skip CertManagerCR Reconcile.")
		return nil
	}

	// No ICP CertManager
	certMgr := newCertManagerCR(instance)
	err = controllerutil.SetControllerReference(instance, certMgr, scheme)
	if err != nil {
		return err
	}

	foundCertManager := &klusterletv1alpha1.CertManager{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: certMgr.Name, Namespace: certMgr.Namespace}, foundCertManager)
	if err != nil && errors.IsNotFound(err) {
		err := installCRDs(client)
		if err != nil {
			return err
		}

		err = createServiceAccount(client, scheme, instance, certMgr)
		if err != nil {
			return err
		}

		log.Info("Creating a new CertManager", "CertManager.Namespace", certMgr.Namespace, "CertManager.Name", certMgr.Name)
		err = client.Create(context.TODO(), certMgr)
		if err != nil {
			return err
		}

		err = createSelfSignClusterIssuer(client, scheme, instance)
		if err != nil {
			return err
		}

		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func createSelfSignClusterIssuer(client client.Client, scheme *runtime.Scheme, cr *klusterletv1alpha1.KlusterletService) error {
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

	return nil
}

func createServiceAccount(client client.Client, scheme *runtime.Scheme, instance *klusterletv1alpha1.KlusterletService, certmgr *klusterletv1alpha1.CertManager) error {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      certmgr.Spec.ServiceAccount.Name,
			Namespace: certmgr.Namespace,
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

func installCRDs(client client.Client) error {
	err := installCertificateCRD(client)
	if err != nil {
		return err
	}
	err = installIssuerCRD(client)
	if err != nil {
		return err
	}
	err = installClusterIssuerCRD(client)
	if err != nil {
		return err
	}
	err = installOrderCRD(client)
	if err != nil {
		return err
	}
	err = installChallengeCRD(client)
	if err != nil {
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
			Name:      cr.Name + "-certmgr",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.CertManagerSpec{
			ClusterResourceNamespace: cr.Namespace,
			ServiceAccount: klusterletv1alpha1.CertManagerServiceAccount{
				Create: false,
				Name:   cr.Name + "-certmgr",
			},
			FullNameOverride: cr.Name + "-certmgr",
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
		return client.Create(context.TODO(), certificatesCRD)
	}

	if err != nil {
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
		return client.Create(context.TODO(), issuerCRD)
	}

	if err != nil {
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
		return client.Create(context.TODO(), clusterIssuerCRD)
	}

	if err != nil {
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
		return client.Create(context.TODO(), orderCRD)
	}

	if err != nil {
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
		return client.Create(context.TODO(), challengeCRD)
	}

	if err != nil {
		return err
	}
	return nil
}
