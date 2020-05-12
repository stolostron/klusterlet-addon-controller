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
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	applicationManager        = "endpoint-appmgr"
	certPolicyController      = "endpoint-certpolicyctrl"
	connectionManager         = "endpoint-connmgr"
	policyController          = "endpoint-policyctrl"
	searchCollector           = "endpoint-search"
	workManager               = "endpoint-workmgr"
	endpointComponentOperator = "endpoint-component-operator"
)

const (
	//We can not test on the sha value as the image manifest is overwriten by CICD
	endpointComponentOperatorContainer = "endpoint-component-operator"
	endpointComponentOperatorImage     = "endpoint-component-operator"
	endpointComponentOperatorSha       = "sha256:b3edec494a5c9f5a9bf65699d0592ca2e50c205132f5337e8df07a7808d03887"
	endpointComponentOperatorImagePath = defaultImageRegistry + "/" + endpointComponentOperatorImage

	certPolicyControllerImage  = "cert-policy-controller"
	certPolicyControllerShaKey = "cert_policy_controller"

	searchCollectorImage  = "search-collector"
	searchCollectorShaKey = "search_collector"

	policyControllerImage  = "mcm-compliance"
	policyControllerShaKey = "mcm_compliance"

	applicationManagerSubImage  = "multicluster-operators-subscription"
	applicationManagerSubShaKey = "multicluster_operators_subscription"

	applicationManagerDepImage  = "multicluster-operators-deployable"
	applicationManagerDepShaKey = "multicluster_operators_deployable"

	connectionManagerImage  = "multicloud-manager"
	connectionManagerShaKey = "multicloud_manager"

)

var deletePatchString = fmt.Sprintf(
	"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
	"replace", "/spec/applicationManager/enabled", false,
)

var addPatchString = fmt.Sprintf(
	"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
	"replace", "/spec/applicationManager/enabled", true,
)

var _ = Describe("Endpoint", func() {

	It("Should create all component CR", func() {
		endpoint := newEndpoint(testEndpointName, testNamespace)
		clientClusterDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Delete(testEndpointName, &metav1.DeleteOptions{})
		createNewUnstructured(clientClusterDynamic, gvrEndpoint,
			endpoint, testEndpointName, testNamespace)
		When("endpoint created, wait for all component CRs to be created", func() {
			var endpointComponentOperatorDeployment *appsv1.Deployment
			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait endpoint component operator...")
				endpointComponentOperatorDeployment, err = clientCluster.AppsV1().Deployments(testNamespace).Get(endpointComponentOperator, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("endpoint component operator created")
			var image string
			for _, c := range endpointComponentOperatorDeployment.Spec.Template.Spec.Containers {
				if c.Name == endpointComponentOperatorContainer {
					image = c.Image
					klog.Infof("image:%s", image)
					break
				}
			}
			if useSha {
				splits := strings.Split(image, "@")
				//We can not test the sha itself because manifest is overwriten in CICD
				Expect(len(splits)).To(Equal(2))
				Expect(len(splits[1]) > 0).To((BeTrue()))
			} else if tagPostfix != "" {
				//We can not test the tag itself because it is defined in CICD
				Expect(strings.Contains(image, tagPostfix)).To(BeTrue())
			} else {
				Expect(len(image) > len(endpointComponentOperatorImagePath)+1).To(BeTrue())
			}

			var cr *unstructured.Unstructured
			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait cert policy controller...")
				cr, err = clientClusterDynamic.Resource(gvrCertpoliciescontroller).Namespace(testNamespace).Get(certPolicyController, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("cert policy controller created")

			checkImageAttributes(cr, useSha, defaultImageRegistry, tagPostfix)

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait search controller...")
				cr, err = clientClusterDynamic.Resource(gvrSearchcollector).Namespace(testNamespace).Get(searchCollector, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("search controller created")

			checkImageAttributes(cr, useSha, defaultImageRegistry, tagPostfix)

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait policy controller...")
				cr, err = clientClusterDynamic.Resource(gvrPolicycontroller).Namespace(testNamespace).Get(policyController, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("policy controller created")

			checkImageAttributes(cr, useSha, defaultImageRegistry, tagPostfix)

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait application manager...")
				cr, err = clientClusterDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(applicationManager, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("application manager created")

			checkImageAttributes(cr, useSha, defaultImageRegistry, tagPostfix)

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait connection manager...")
				cr, err = clientClusterDynamic.Resource(gvrConnectionmanager).Namespace(testNamespace).Get(connectionManager, metav1.GetOptions{})
				return err
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("connection manager created")

			checkImageAttributes(cr, useSha, defaultImageRegistry, tagPostfix)
		})
	})

	It("Should delete corresponding component CR", func() {
		By("Updating endpoint")
		_, err := clientClusterDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Patch(testEndpointName, types.JSONPatchType, []byte(deletePatchString), metav1.PatchOptions{})
		Expect(err).To(BeNil())

		When("endpoint update, wait for corresponding component to create/delete", func() {
			Eventually(func() *unstructured.Unstructured {
				var objAppmgr *unstructured.Unstructured
				klog.V(1).Info("Wait application manager component...")
				objAppmgr, err = clientClusterDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(applicationManager, metav1.GetOptions{})
				return objAppmgr
			}, 10, 1).Should(BeNil())
			klog.V(1).Info("application manager deleted")
		})
	})

	It("Should add corresponding component CR", func() {
		By("Updating endpoint")
		_, err := clientClusterDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Patch(testEndpointName, types.JSONPatchType, []byte(addPatchString), metav1.PatchOptions{})
		Expect(err).To(BeNil())

		When("endpoint update, wait for corresponding component to create/delete", func() {
			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait application manager...")
				_, err = clientClusterDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(applicationManager, metav1.GetOptions{})
				return err
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("application manager created")
		})
	})

	It("Should delete all component CRs", func() {
		By("Deleteing endpoint")
		err := clientClusterDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Delete(testEndpointName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

		When("endpoint deleted, wait for all components CR to be deleted", func() {
			Eventually(func() *unstructured.Unstructured {
				var objCertPolicyCtl *unstructured.Unstructured
				klog.V(1).Info("Wait deletion cert policy controller...")
				objCertPolicyCtl, err = clientClusterDynamic.Resource(gvrCertpoliciescontroller).Namespace(testNamespace).Get(certPolicyController, metav1.GetOptions{})
				return objCertPolicyCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("cert policy controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objSearchCtl *unstructured.Unstructured
				klog.V(1).Info("Wait deletion search controller...")
				objSearchCtl, err = clientClusterDynamic.Resource(gvrSearchcollector).Namespace(testNamespace).Get(searchCollector, metav1.GetOptions{})
				return objSearchCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("search controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objPolicyCtl *unstructured.Unstructured
				klog.V(1).Info("Wait deletion policy controller...")
				objPolicyCtl, err = clientClusterDynamic.Resource(gvrPolicycontroller).Namespace(testNamespace).Get(policyController, metav1.GetOptions{})
				return objPolicyCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("policy controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objAppMgr *unstructured.Unstructured
				klog.V(1).Info("Wait deletion application manager...")
				objAppMgr, err = clientClusterDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(applicationManager, metav1.GetOptions{})
				return objAppMgr
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("application manager deleted")

			Eventually(func() *unstructured.Unstructured {
				var objConnMgr *unstructured.Unstructured
				klog.V(1).Info("Wait deletion connection manager...")
				objConnMgr, err = clientClusterDynamic.Resource(gvrConnectionmanager).Namespace(testNamespace).Get(connectionManager, metav1.GetOptions{})
				return objConnMgr
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("connection manager deletedd")

			Eventually(func() error {
				klog.V(1).Info("Wait deletion endpoint component operator...")
				_, err = clientCluster.AppsV1().Deployments(testNamespace).Get(endpointComponentOperator, metav1.GetOptions{})
				return err
			}, 5, 1).ShouldNot(BeNil())
			klog.V(1).Info("endpoint component operator deleted")
		})
	})
})

func checkImageAttributes(cr *unstructured.Unstructured, useSha bool, repository, tagPostfix string) {
	spec := cr.Object["spec"].(map[string]interface{})
	global := spec["global"].(map[string]interface{})
	imageMap, ok, err := unstructured.NestedStringMap(global, "imageOverrides")
	Expect(err).To(BeNil())
	Expect(ok).To(BeTrue())
	for _, repositoryCR := range imageMap {

		var splits []string
		if useSha {
			splits = strings.Split(repositoryCR, "@")
			Expect(len(splits)).To(Equal(2))
			Expect(strings.Contains(splits[1], "sha256:")).To(BeTrue())
		} else {
			splits = strings.Split(repositoryCR, ":")
			if tagPostfix != "" {
				Expect(strings.HasSuffix(splits[1], tagPostfix)).To(BeTrue())
			} else {
				Expect(splits[1]).NotTo(BeEmpty())
			}
		}
		//We can not test on the image name because we don't it in the CR.
		Expect(strings.HasPrefix(splits[0], repository+"/")).To(BeTrue())
	}
}
