// +build functional

// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package endpoint_operator_test

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	endpointOperator           = "endpoint-operator"
	endpointOperatorContainer  = "endpoint-operator"
	defaultImageRegistry       = "quay.io/open-cluster-management"
	defaultImagePullSecretName = "multicloud-image-pull-secret"
	testEndpointName           = "endpoint"
	testNamespace              = "multicluster-endpoint"
)

var (
	useSha               bool
	tagPostfix           string
	clientCluster        kubernetes.Interface
	clientClusterDynamic dynamic.Interface

	gvrEndpoint               schema.GroupVersionResource
	gvrApplicationmanager     schema.GroupVersionResource
	gvrCertpoliciescontroller schema.GroupVersionResource
	gvrCiscontroller          schema.GroupVersionResource
	gvrConnectionmanager      schema.GroupVersionResource
	gvrIampoliciescontroller  schema.GroupVersionResource
	gvrPolicycontroller       schema.GroupVersionResource
	gvrSearchcollector        schema.GroupVersionResource
	gvrWorkmanagers           schema.GroupVersionResource

	optionsFile         string
	baseDomain          string
	kubeadminUser       string
	kubeadminCredential string
	kubeconfig          string
)

func newEndpoint(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "multicloud.ibm.com/v1beta1",
			"kind":       "Endpoint",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"applicationManager": map[string]interface{}{
					"enabled": true,
				},
				"certPolicyController": map[string]interface{}{
					"enabled": true,
				},
				"bootstrapConfig": map[string]interface{}{
					"hubSecret": "multicluster-endpoint/klusterlet-bootstrap",
				},
				"clusterLabels": map[string]interface{}{
					"cloud":  "auto-detect",
					"vendor": "auto-detect",
				},
				"clusterName":      "testCluster",
				"clusterNamespace": "testCluster",
				"connectionManager": map[string]interface{}{
					"enabledGlobalView": false,
				},
				"imageRegistry":   defaultImageRegistry,
				"imagePullSecret": "multicloud-image-pull-secret",
				"policyController": map[string]interface{}{
					"enabled": true,
				},
				"searchCollector": map[string]interface{}{
					"enabled": true,
				},
				"cisController": map[string]interface{}{
					"enabled": false,
				},
				"iamPolicyController": map[string]interface{}{
					"enabled": false,
				},
				"version": "1.0.0",
			},
		},
	}
}

// createNewUnstructured creates resources by using gvr & obj
func createNewUnstructured(
	clientClusterDynamic dynamic.Interface,
	gvr schema.GroupVersionResource,
	obj *unstructured.Unstructured,
	name, namespace string,
) {
	ns := clientClusterDynamic.Resource(gvr).Namespace(namespace)
	Expect(ns.Create(obj, metav1.CreateOptions{})).NotTo(BeNil())
	Expect(ns.Get(name, metav1.GetOptions{})).NotTo(BeNil())
}

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	flag.StringVar(&kubeadminUser, "kubeadmin-user", "kubeadmin", "Provide the kubeadmin credential for the cluster under test (e.g. -kubeadmin-user=\"xxxxx\").")
	flag.StringVar(&kubeadminCredential, "kubeadmin-credential", "", "Provide the kubeadmin credential for the cluster under test (e.g. -kubeadmin-credential=\"xxxxx-xxxxx-xxxxx-xxxxx\").")
	flag.StringVar(&baseDomain, "base-domain", "", "Provide the base domain for the cluster under test (e.g. -base-domain=\"demo.red-chesterfield.com\").")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Location of the kubeconfig to use; defaults to KUBECONFIG if not set")

	flag.StringVar(&optionsFile, "options", "", "Location of an \"options.yaml\" file to provide input for various tests")

}
func TestEndpointOperator(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "EndpointOperator Suite")
}

var _ = BeforeSuite(func() {
	By("Setup Kube client")
	gvrEndpoint = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "endpoints"}
	gvrApplicationmanager = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "applicationmanagers"}
	gvrCertpoliciescontroller = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "certpoliciescontroller"}
	gvrCiscontroller = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "ciscontrollers"}
	gvrConnectionmanager = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "connectionmanagers"}
	gvrIampoliciescontroller = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "iampoliciescontroller"}
	gvrPolicycontroller = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "policycontrollers"}
	gvrSearchcollector = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "searchcollectors"}
	gvrWorkmanagers = schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "workmanagers"}

	clientCluster = NewKubeClient("", "", "")
	clientClusterDynamic = NewKubeClientDynamic("", "", "")
	By("Create Namesapce if needed")
	namespaces := clientCluster.CoreV1().Namespaces()
	if _, err := namespaces.Get(testNamespace, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
		Expect(namespaces.Create(&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		})).NotTo(BeNil())
	}
	d, err := clientCluster.AppsV1().Deployments(testNamespace).Get(endpointOperator, metav1.GetOptions{})
	if err != nil {
		klog.V(5).Infof("endpoint-operator:\n%#v", d)
	}
	Expect(err).To(BeNil())
	useSha = false
	tagPostfix = ""
	for _, c := range d.Spec.Template.Spec.Containers {
		if c.Name == endpointOperatorContainer {
			for _, e := range c.Env {
				if e.Name == "USE_SHA_MANIFEST" {
					useSha = e.Value == strings.ToLower("true")
					klog.V(5).Infof("useSha=%t", useSha)
				}
				if e.Name == "IMAGE_TAG_POSTFIX" {
					tagPostfix = e.Value
					klog.V(5).Infof("tagPostFix=%s", tagPostfix)
				}
			}
			break
		}
	}
	Expect(namespaces.Get(testNamespace, metav1.GetOptions{})).NotTo(BeNil())

})

func NewKubeClient(url, kubeconfig, context string) kubernetes.Interface {
	klog.V(5).Infof("Create kubeclient for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := LoadConfig(url, kubeconfig, context)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func NewKubeClientDynamic(url, kubeconfig, context string) dynamic.Interface {
	klog.V(5).Infof("Create kubeclient dynamic for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := LoadConfig(url, kubeconfig, context)
	if err != nil {
		panic(err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func LoadConfig(url, kubeconfig, context string) (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	klog.V(5).Infof("Kubeconfig path %s\n", kubeconfig)
	// If we have an explicit indication of where the kubernetes config lives, read that.
	if kubeconfig != "" {
		if context == "" {
			return clientcmd.BuildConfigFromFlags(url, kubeconfig)
		}
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
			&clientcmd.ConfigOverrides{
				CurrentContext: context,
			}).ClientConfig()
	}
	// If not, try the in-cluster config.
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory.
	if usr, err := user.Current(); err == nil {
		klog.V(5).Infof("clientcmd.BuildConfigFromFlags for url %s using %s\n", url, filepath.Join(usr.HomeDir, ".kube/config"))
		if c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube/config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")

}
