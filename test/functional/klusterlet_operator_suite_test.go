// +build functional

// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package klusterlet_addon_controller_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
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
	klusterletAddonController = "klusterlet-addon-controller"
	defaultImageRegistry      = "quay.io/open-cluster-management"
	testNamespace             = "test-klusterlet-addon-controller"

	klusterletAddonNamespace = "open-cluster-management"
)

var (
	//useSha               bool
	//tagPostfix           string
	clientCluster        kubernetes.Interface
	clientClusterDynamic dynamic.Interface

	gvrKlusterletAddonConfig  schema.GroupVersionResource
	gvrManifestwork           schema.GroupVersionResource
	gvrManagedCluster         schema.GroupVersionResource
	gvrManagedClusterAddOn    schema.GroupVersionResource
	gvrClusterManagementAddOn schema.GroupVersionResource
	gvrLease                  schema.GroupVersionResource

	kubeconfig    string
	imageRegistry string
)

func newLease(name, namespace string, renewTime string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "coordination.k8s.io/v1",
			"kind":       "Lease",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"leaseDurationSeconds": 60,
				"renewTime":            renewTime,
			},
		},
	}
}

func newManagedCluster(name, namespace string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.open-cluster-management.io/v1",
			"kind":       "ManagedCluster",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"hubAcceptsClient": true,
			},
			"status": map[string]interface{}{
				"conditions": []map[string]interface{}{
					map[string]interface{}{
						"type":               "placeholder",
						"lastTransitionTime": "2020-01-01T01:01:01Z",
						"reason":             "placeholder",
						"status":             "False",
					},
				},
			},
		},
	}
}
func newKlusterletAddonConfig(name, namespace, version string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "agent.open-cluster-management.io/v1",
			"kind":       "KlusterletAddonConfig",
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
				"clusterLabels": map[string]interface{}{
					"cloud":  "auto-detect",
					"vendor": "auto-detect",
				},
				"clusterName":      "testCluster",
				"clusterNamespace": "testCluster",
				"imageRegistry":    imageRegistry,
				"imagePullSecret":  "multicloud-image-pull-secret",
				"policyController": map[string]interface{}{
					"enabled": true,
				},
				"searchCollector": map[string]interface{}{
					"enabled": true,
				},
				"iamPolicyController": map[string]interface{}{
					"enabled": true,
				},
				"version": version,
			},
		},
	}
}

// deleteIfExists deletes resources by using gvr & name & namespace, will wait for deletion to complete by using eventually
func deleteIfExists(clientHubDynamic dynamic.Interface, gvr schema.GroupVersionResource, name, namespace string) {
	ns := clientHubDynamic.Resource(gvr).Namespace(namespace)
	if _, err := ns.Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
		Expect(errors.IsNotFound(err)).To(Equal(true))
		return
	}

	Expect(func() error {
		// possibly already got deleted
		err := ns.Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		return nil
	}()).To(BeNil())
	// remove finalizers if needed
	if obj, err := ns.Get(context.TODO(), name, metav1.GetOptions{}); err == nil {
		if len(obj.GetFinalizers()) > 0 {
			obj.SetFinalizers([]string{})
			_, _ = ns.Update(context.TODO(), obj, metav1.UpdateOptions{})
		}
	}

	klog.V(2).Info("Wait for deletion")
	Eventually(func() error {
		var err error
		_, err = ns.Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		if err == nil {
			return fmt.Errorf("found object %s in namespace %s after deletion", name, namespace)
		}
		return nil
	}, 10, 1).Should(BeNil())
}

// createNewUnstructured creates resources by using gvr & obj
func createNewUnstructured(
	clientClusterDynamic dynamic.Interface,
	gvr schema.GroupVersionResource,
	obj *unstructured.Unstructured,
	name, namespace string,
) {
	ns := clientClusterDynamic.Resource(gvr).Namespace(namespace)
	Expect(ns.Create(context.TODO(), obj, metav1.CreateOptions{})).NotTo(BeNil())
	Expect(ns.Get(context.TODO(), name, metav1.GetOptions{})).NotTo(BeNil())
}

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Location of the kubeconfig to use; defaults to KUBECONFIG if not set")
	flag.StringVar(&imageRegistry, "image-registry", defaultImageRegistry, fmt.Sprintf("URL if the image registry (ie: %s", defaultImageRegistry))

}
func TestKlusterletOperator(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "KlusterletOperator Suite")
}

var _ = BeforeSuite(func() {
	klog.V(1).Info("running before suite")
	By("Setup Kube client")
	gvrKlusterletAddonConfig = schema.GroupVersionResource{Group: "agent.open-cluster-management.io", Version: "v1", Resource: "klusterletaddonconfigs"}
	gvrManifestwork = schema.GroupVersionResource{Group: "work.open-cluster-management.io", Version: "v1", Resource: "manifestworks"}
	gvrManagedCluster = schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
	gvrClusterManagementAddOn = schema.GroupVersionResource{Group: "addon.open-cluster-management.io", Version: "v1alpha1", Resource: "clustermanagementaddons"}
	gvrManagedClusterAddOn = schema.GroupVersionResource{Group: "addon.open-cluster-management.io", Version: "v1alpha1", Resource: "managedclusteraddons"}
	gvrLease = schema.GroupVersionResource{Group: "coordination.k8s.io", Version: "v1", Resource: "leases"}

	clientCluster = NewKubeClient("", "", "")
	clientClusterDynamic = NewKubeClientDynamic("", "", "")
	By("Create Namesapce if needed")
	namespaces := clientCluster.CoreV1().Namespaces()
	if _, err := namespaces.Get(context.TODO(), testNamespace, metav1.GetOptions{}); err != nil && errors.IsNotFound(err) {
		Expect(namespaces.Create(context.TODO(), &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}, metav1.CreateOptions{})).NotTo(BeNil())
	}
	d, err := clientCluster.AppsV1().Deployments(klusterletAddonNamespace).Get(context.TODO(), klusterletAddonController, metav1.GetOptions{})
	if err != nil {
		klog.V(1).Infof("klusterlet-addon-controller:\n%#v", d)
	}
	Expect(err).To(BeNil())

})

func NewKubeClient(url, kubeconfig, context string) kubernetes.Interface {
	klog.V(1).Infof("Create kubeclient for url %s using kubeconfig path %s\n", url, kubeconfig)
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
	klog.V(1).Infof("Create kubeclient dynamic for url %s using kubeconfig path %s\n", url, kubeconfig)
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

	klog.V(1).Infof("Kubeconfig path %s\n", kubeconfig)
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
		klog.V(1).Infof("clientcmd.BuildConfigFromFlags for url %s using %s\n", url, filepath.Join(usr.HomeDir, ".kube/config"))
		if c, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube/config")); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not create a valid kubeconfig")

}
