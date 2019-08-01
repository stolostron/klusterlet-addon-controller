// Package inspect provide information and utils about the cluster itself
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package inspect

import (
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("inspect")

// InitClusterInfo initialize the global variable Info in the inspec package
func InitClusterInfo(cfg *rest.Config) error {
	// Initialize RESTClient
	cfg.NegotiatedSerializer = unstructuredscheme.NewUnstructuredNegotiatedSerializer()
	kubeRESTClient, err := rest.UnversionedRESTClientFor(cfg)
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
	} else if strings.Contains(gitVersion, string(KubeVendorEKS)) {
		Info.KubeVendor = KubeVendorEKS
	} else if strings.Contains(gitVersion, string(KubeVendorICP)) {
		Info.KubeVendor = KubeVendorICP
	} else if strings.Contains(gitVersion, string(KubeVendorGKE)) {
		Info.KubeVendor = KubeVendorGKE
	} else {
		Info.OpenShiftVersion = getOpenShiftVersion(kubeRESTClient)
		if Info.OpenShiftVersion.GitVersion != "" {
			Info.KubeVendor = KubeVendorOpenShift
		} else {
			Info.KubeVendor = KubeVendorOther
		}
	}

	// Set CloudVendor from KubeVendor
	switch kubeVendor := Info.KubeVendor; kubeVendor {
	case KubeVendorAKS:
		Info.CloudVendor = CloudVendorAzure
	case KubeVendorEKS:
		Info.CloudVendor = CloudVendorAWS
	case KubeVendorGKE:
		Info.CloudVendor = CloudVendorGoogle
	case KubeVendorIKS:
		Info.CloudVendor = CloudVendorIBM
	default:
		Info.CloudVendor = cloudVendorFromNodeProviderID(kubeRESTClient)
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

	if len(nodeList.Items) != 0 {
		if strings.Contains(nodeList.Items[0].Spec.ProviderID, "ibm") {
			return CloudVendorIBM
		}
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

func getOpenShiftVersion(client *rest.RESTClient) version.Info {
	openshiftVersion := version.Info{}

	openShiftVersionBody, err := client.Get().AbsPath("/version/openshift").Do().Raw()
	if err != nil {
		log.Error(err, "fail to GET /version/openshift, assuming not OpenShift")
		return version.Info{}
	}

	err = json.Unmarshal(openShiftVersionBody, &openshiftVersion)
	if err != nil {
		log.Error(fmt.Errorf("fail to Unmarshal, got '%s': %v", string(openShiftVersionBody), err), "")
		return version.Info{}
	}

	return openshiftVersion
}
