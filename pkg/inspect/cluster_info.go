// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package inspect provide information and utils about the cluster itself
package inspect

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mcmv1alpha1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/mcm/v1alpha1"
)

var log = logf.Log.WithName("inspect")

// DeployedOnHub checks that: Is this cluster a Hub Cluster?
func DeployedOnHub(c client.Client) bool {
	clusterStatusList := &mcmv1alpha1.ClusterStatusList{}
	err := c.List(context.TODO(), clusterStatusList, &client.ListOptions{})
	return err == nil
}

// OpenshiftPrometheusService check: Is the cluster have the openshift prometheus service?
func OpenshiftPrometheusService(client client.Client) bool {
	foundOpenshiftPrometheusService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "prometheus-k8s", Namespace: "openshift-monitoring"}, foundOpenshiftPrometheusService)
	return err == nil
}

// ICPPrometheusService check: Is the cluster have the openshift prometheus service?
func ICPPrometheusService(client client.Client) bool {
	foundICPPrometheusService := &corev1.Service{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "monitoring-prometheus", Namespace: "kube-system"}, foundICPPrometheusService)
	return err == nil
}

// IBMCloudClusterInfoConfigMapExist check if the cluster have ibmcloud-cluster-info configmap
func IBMCloudClusterInfoConfigMapExist(client client.Client) bool {
	foundConfigMap := &corev1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: "ibmcloud-cluster-info", Namespace: "kube-public"}, foundConfigMap)
	return err == nil
}

// InitClusterInfo initialize the global variable Info in the inspec package
func InitClusterInfo(cfg *rest.Config) error {
	// Initialize RESTClient
	oldNegotiatedSerializer := cfg.NegotiatedSerializer
	cfg.NegotiatedSerializer = unstructuredscheme.NewUnstructuredNegotiatedSerializer()
	kubeRESTClient, err := rest.UnversionedRESTClientFor(cfg)
	// restore cfg before leaving
	defer func(cfg *rest.Config) { cfg.NegotiatedSerializer = oldNegotiatedSerializer }(cfg)
	if err != nil {
		log.Error(err, "Fail to initialize RESTClient")
		Info.KubeVendor = KubeVendorOther
		Info.CloudVendor = CloudVendorOther
		return err
	}

	// Set Kubernetes Version info
	Info.KubeVersion = getKubeVersion(kubeRESTClient)

	// Set KubeVendor base on KubeVersion
	gitVersion := strings.ToUpper(Info.KubeVersion.GitVersion)
	if strings.Contains(gitVersion, string(KubeVendorIKS)) {
		Info.KubeVendor = KubeVendorIKS
		Info.CloudVendor = CloudVendorIBM
	} else if strings.Contains(gitVersion, string(KubeVendorEKS)) {
		Info.KubeVendor = KubeVendorEKS
		Info.CloudVendor = CloudVendorAWS
	} else if strings.Contains(gitVersion, string(KubeVendorICP)) {
		Info.KubeVendor = KubeVendorICP
	} else if strings.Contains(gitVersion, string(KubeVendorGKE)) {
		Info.KubeVendor = KubeVendorGKE
	} else if isOpenshift(kubeRESTClient) {
		Info.KubeVendor = KubeVendorOpenShift
	} else {
		Info.KubeVendor = KubeVendorOther
	}

	// Set CloudVendor from KubeVendor
	switch kubeVendor := Info.KubeVendor; kubeVendor {
	case KubeVendorEKS:
		Info.CloudVendor = CloudVendorAWS
	case KubeVendorGKE:
		Info.CloudVendor = CloudVendorGoogle
	case KubeVendorIKS:
		Info.CloudVendor = CloudVendorIBM
	default:
		Info.CloudVendor = cloudVendorFromNodeProviderID(kubeRESTClient)
	}

	if Info.CloudVendor == CloudVendorAzure && Info.KubeVendor == KubeVendorOther {
		Info.KubeVendor = KubeVendorAKS
	}

	log.Info("", "Info.KubeVendor", Info.KubeVendor)
	log.Info("", "Info.CloudVendor", Info.CloudVendor)
	return nil
}

func cloudVendorFromNodeProviderID(client *rest.RESTClient) CloudVendor {
	nodeList := &corev1.NodeList{}
	nodesListBody, err := client.Get().AbsPath("/api/v1/nodes").Do().Raw()
	if err != nil {
		log.Error(err, "fail to GET /api/v1/nodes")
		return CloudVendorOther
	}

	err = json.Unmarshal(nodesListBody, nodeList)
	if err != nil {
		log.Error(fmt.Errorf("fail to Unmarshel, got '%s': %v", string(nodesListBody), err), "")
		return CloudVendorOther
	}

	if len(nodeList.Items) == 0 {
		return CloudVendorOther
	}

	if strings.Contains(nodeList.Items[0].Spec.ProviderID, "ibm") {
		return CloudVendorIBM
	} else if strings.Contains(nodeList.Items[0].Spec.ProviderID, "azure") {
		return CloudVendorAzure
	} else if strings.Contains(nodeList.Items[0].Spec.ProviderID, "aws") {
		return CloudVendorAWS
	} else if strings.Contains(nodeList.Items[0].Spec.ProviderID, "gce") {
		return CloudVendorGoogle
	}

	return CloudVendorOther
}

func getKubeVersion(client *rest.RESTClient) version.Info {
	kubeVersion := version.Info{}

	versionBody, err := client.Get().AbsPath("/version").Do().Raw()
	if err != nil {
		log.Error(err, "fail to GET /version")
		return version.Info{}
	}

	err = json.Unmarshal(versionBody, &kubeVersion)
	if err != nil {
		log.Error(fmt.Errorf("fail to Unmarshal, got '%s': %v", string(versionBody), err), "")
		return version.Info{}
	}

	return kubeVersion
}

func isOpenshift(client *rest.RESTClient) bool {
	//check whether the cluster is openshift or not for openshift version 3.11 and before
	_, err := client.Get().AbsPath("/version/openshift").Do().Raw()
	if err == nil {
		log.Info("Found openshift version from /version/openshift")
		return true
	}

	//check whether the cluster is openshift or not for openshift version 4.1
	_, err = client.Get().AbsPath("/apis/config.openshift.io/v1/clusterversions").Do().Raw()
	if err == nil {
		log.Info("Found openshift version from /apis/config.openshift.io/v1/clusterversions")
		return true
	}

	log.Error(err, "fail to GET openshift version, assuming not OpenShift")
	return false
}
