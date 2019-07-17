package klusterletservice

import (
	"context"

	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	openshiftsecurityv1 "github.com/openshift/api/security/v1"
	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_klusterletservice")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new KlusterletService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKlusterletService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("klusterletservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource KlusterletService
	err = c.Watch(&source.Kind{Type: &klusterletv1alpha1.KlusterletService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner KlusterletService
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &klusterletv1alpha1.KlusterletService{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileKlusterletService{}

// ReconcileKlusterletService reconciles a KlusterletService object
type ReconcileKlusterletService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a KlusterletService object and makes changes based on the state read
// and what is in the KlusterletService.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKlusterletService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling KlusterletService")

	// Fetch the KlusterletService instance
	instance := &klusterletv1alpha1.KlusterletService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// CertManager
	certMgr := newCertManagerCR(instance)
	if err := controllerutil.SetControllerReference(instance, certMgr, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	foundCertManager := &klusterletv1alpha1.CertManager{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: certMgr.Name, Namespace: certMgr.Namespace}, foundCertManager)

	if err != nil && errors.IsNotFound(err) {
		if err := installCertificateCRD(r); err != nil {
			return reconcile.Result{}, err
		}
		if err := installIssuerCRD(r); err != nil {
			return reconcile.Result{}, err
		}
		if err := installClusterIssuerCRD(r); err != nil {
			return reconcile.Result{}, err
		}

		certmgrSA, err := getOrCreateServiceAccount(r, certMgr.Spec.ServiceAccount.Name, instance.Namespace)
		if err != nil {
			return reconcile.Result{}, err
		}
		privilegedSCC := &openshiftsecurityv1.SecurityContextConstraints{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Name: "privileged", Namespace: ""}, privilegedSCC); err != nil {
			return reconcile.Result{}, err
		}
		if err := addServiceAccountToSCC(r, certmgrSA, privilegedSCC); err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Creating a new CertManager", "CertManager.Namespace", certMgr.Namespace, "CertManager.Name", certMgr.Name)
		if err := r.client.Create(context.TODO(), certMgr); err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	if err := createSelfSignIssuer(r, "self-signed", instance.Namespace); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func createSelfSignIssuer(r *ReconcileKlusterletService, name string, namespace string) error {
	clusterIssuer := &certmanagerv1alpha1.ClusterIssuer{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, clusterIssuer)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating self signed cluster issuer", "Name", name, "Namespace", namespace)
		clusterIssuer = &certmanagerv1alpha1.ClusterIssuer{
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
		if err := r.client.Create(context.TODO(), clusterIssuer); err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateServiceAccount(r *ReconcileKlusterletService, name string, namespace string) (*corev1.ServiceAccount, error) {
	log.Info("Get or create service account", "Name", name, "Namespace", namespace)

	serviceAccount := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating service account", "Name", name, "Namespace", namespace)
		serviceAccount = &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		if err := r.client.Create(context.TODO(), serviceAccount); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	log.Info("Get or create service account", "Name", serviceAccount.Name, "Namespace", serviceAccount.Namespace)
	return serviceAccount, nil
}

func addServiceAccountToSCC(r *ReconcileKlusterletService, sa *corev1.ServiceAccount, scc *openshiftsecurityv1.SecurityContextConstraints) error {
	user := "system:serviceaccount:" + sa.Namespace + ":" + sa.Name
	log.Info("Add ServiceAccount to SecurityContextConstraints", "user", user, "scc.Name", scc.Name)
	scc.Users = append(scc.Users, user)
	return r.client.Update(context.TODO(), scc)
}

func addImagePullSecretToServiceAccount() {}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *klusterletv1alpha1.KlusterletService) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

func newCertManagerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.CertManager {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.CertManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.CertManagerSpec{
			ClusterResourceNamespace: cr.Namespace,
			ServiceAccount: klusterletv1alpha1.CertManagerServiceAccount{
				Create: false,
				Name:   "cert-manager",
			},
		},
	}
}

func installCertificateCRD(r *ReconcileKlusterletService) error {
	log.Info("Installing certificates.certmanager.k8s.io CRD")

	certificatesCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "certificates.certmanager.k8s.io", Namespace: ""}, certificatesCRD)
	if err != nil && errors.IsNotFound(err) {
		//create certificates.certmanager.k8s.io CRDs
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
		err = r.client.Create(context.TODO(), certificatesCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installIssuerCRD(r *ReconcileKlusterletService) error {
	log.Info("Installing issuers.certmanager.k8s.io CRD")

	issuerCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "issuers.certmanager.k8s.io", Namespace: ""}, issuerCRD)
	if err != nil && errors.IsNotFound(err) {
		//create certificates.certmanager.k8s.io CRDs
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
		err = r.client.Create(context.TODO(), issuerCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func installClusterIssuerCRD(r *ReconcileKlusterletService) error {
	log.Info("Installing clusterissuers.certmanager.k8s.io CRD")

	clusterIssuerCRD := &apiextensionv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "clusterissuers.certmanager.k8s.io", Namespace: ""}, clusterIssuerCRD)
	if err != nil && errors.IsNotFound(err) {
		//create certificates.certmanager.k8s.io CRDs
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
		err = r.client.Create(context.TODO(), clusterIssuerCRD)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
