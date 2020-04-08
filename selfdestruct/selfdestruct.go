package main

import (
	"flag"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var operatorNamespace string

const ENDPOINT_DELETION_TIMEOUT int = 60
const CR_DELETION_TIMEOUT int = 15

var componentCRDs []string = []string{
	"applicationmanagers.multicloud.ibm.com",
	"certpoliciescontroller.multicloud.ibm.com",
	"ciscontrollers.multicloud.ibm.com",
	"connectionmanagers.multicloud.ibm.com",
	"iampoliciescontroller.multicloud.ibm.com",
	"policycontrollers.multicloud.ibm.com",
	"searchcollectors.multicloud.ibm.com",
	"serviceregistries.multicloud.ibm.com",
	"workmanagers.multicloud.ibm.com",
	"endpoints.multicloud.ibm.com",
}

func main() {

	klog.InitFlags(nil)

	flag.StringVar(&operatorNamespace, "operator-namespace", "", "The namespace where the operator was installed")

	flag.Parse()

	if operatorNamespace == "" {
		operatorNamespace = os.Getenv("OPERATOR_NAMESPACE")
		if operatorNamespace == "" {
			operatorNamespace = "multicluster-endpoint"
		}
	}

	cfg, err := config.GetConfig()
	if err != nil {
		klog.Error(err, "")
		os.Exit(1)
	}

	klog.Info("Destruction requested")
	doDestruction(cfg)
	klog.Info("Destruction Completed")
	os.Exit(0)

}

func doDestruction(cfg *rest.Config) {
	clientDynamic := dynamic.NewForConfigOrDie(cfg)
	deleteEndpoints(clientDynamic)

	clientKube := kubernetes.NewForConfigOrDie(cfg)
	deleteEndpointOperator(clientKube)

	deleteEndpointComponentOperator(clientKube)

	clientAPIExtensionV1beta1 := apiextensionsclientset.NewForConfigOrDie(cfg)
	forceDeleteCRDS(clientAPIExtensionV1beta1, clientDynamic)

	deleteNamespace(clientKube)

}

func deleteEndpoints(clientDynamic dynamic.Interface) {
	gvr := schema.GroupVersionResource{Group: "multicloud.ibm.com", Version: "v1beta1", Resource: "endpoints"}
	klog.V(1).Infof("Retrieving resources %v", gvr)
	endpoints, err := clientDynamic.Resource(gvr).Namespace(operatorNamespace).List(metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		return
	}
	_ = endpoints.EachListItem(func(item runtime.Object) error {
		castItem := item.(*unstructured.Unstructured)
		klog.V(1).Infof("Deleting %s/%s", castItem.GetName(), castItem.GetNamespace())
		err := clientDynamic.Resource(gvr).Namespace(castItem.GetNamespace()).Delete(castItem.GetName(), &metav1.DeleteOptions{})
		gps := ENDPOINT_DELETION_TIMEOUT
		for gps != 0 {
			_, err := clientDynamic.Resource(gvr).Namespace(castItem.GetNamespace()).Get(castItem.GetName(), metav1.GetOptions{})
			if err != nil {
				gps = 0
			} else {
				klog.V(1).Infof("Wait endpoints %s/%s deletion", castItem.GetName(), castItem.GetNamespace())
				time.Sleep(1 * time.Second)
				gps -= 1
			}
		}
		if gps == 0 {
			klog.Error("endpoints deletions times out")
		}
		if err != nil {
			klog.Error(err)
		}
		return nil
	})
}

func deleteEndpointOperator(clientKube kubernetes.Interface) {
	klog.V(1).Info("Deleting endpoint-operator deployment")
	err := clientKube.AppsV1().Deployments(operatorNamespace).Delete("endpoint-operator", &metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
	}
}

func deleteEndpointComponentOperator(clientKube kubernetes.Interface) {
	klog.V(1).Info("Deleting endpoint-component-operator deployment")
	err := clientKube.AppsV1().Deployments(operatorNamespace).Delete("endpoint-component-operator", &metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
	}
}

func forceDeleteCRDS(clientAPIExtensionV1beta1 apiextensionsclientset.Interface, clientDynamic dynamic.Interface) {
	klog.V(1).Info("Retrieving CRDs")
	for _, crdName := range componentCRDs {
		crd, err := clientAPIExtensionV1beta1.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
		if err != nil {
			klog.Error(err)
			continue
		}
		forceDeleteCRD(crd, clientAPIExtensionV1beta1, clientDynamic)
	}
}

func forceDeleteCRD(crd *apiextensionsv1beta1.CustomResourceDefinition, clientAPIExtensionV1beta1 apiextensionsclientset.Interface, clientDynamic dynamic.Interface) {
	gvr := schema.GroupVersionResource{Group: crd.Spec.Group, Version: crd.Spec.Version, Resource: crd.Spec.Names.Plural}
	klog.V(1).Infof("Retrieving resources %v", gvr)
	crs, err := clientDynamic.Resource(gvr).Namespace(operatorNamespace).List(metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		return
	}
	_ = crs.EachListItem(func(item runtime.Object) error {
		castItem := item.(*unstructured.Unstructured)
		klog.V(1).Infof("Deleting %s/%s", castItem.GetName(), castItem.GetNamespace())
		resourceInterface := clientDynamic.Resource(gvr).Namespace(castItem.GetNamespace())
		gps := CR_DELETION_TIMEOUT
		_ = resourceInterface.Delete(castItem.GetName(), &metav1.DeleteOptions{})
		for gps != 0 {
			_, err := resourceInterface.Get(castItem.GetName(), metav1.GetOptions{})
			if err != nil {
				gps = 0
			} else {
				klog.V(1).Infof("Wait cr %s/%s deletion", castItem.GetName(), castItem.GetNamespace())
				time.Sleep(1 * time.Second)
				gps -= 1
			}
		}
		if gps == 0 {
			klog.V(1).Infof("Patching %s/%s", castItem.GetName(), castItem.GetNamespace())
			_, err = resourceInterface.Patch(castItem.GetName(), types.JSONPatchType, []byte("[{\"op\": \"remove\", \"path\":\"/metadata/finalizers\"}]"), metav1.PatchOptions{})
			if err != nil {
				klog.Error(err)
			}
		}
		return nil
	})
	klog.V(1).Infof("Deleting CRD %s", crd.GetName())
	err = clientAPIExtensionV1beta1.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.GetName(), &metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
	}
}

func deleteNamespace(clientKube kubernetes.Interface) {
	klog.V(1).Infof("Deleting %s namespace", operatorNamespace)
	err := clientKube.CoreV1().Namespaces().Delete(operatorNamespace, &metav1.DeleteOptions{})
	if err != nil {
		klog.Error(err)
	}
}
