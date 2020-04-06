// +build functional

//copyright
package endpoint_operator_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	ApplicationManager        = "endpoint-appmgr"
	CertPolicyController      = "endpoint-certpolicyctrl"
	ConnectionManager         = "endpoint-connmgr"
	PolicyController          = "endpoint-policyctrl"
	SearchCollector           = "endpoint-search"
	WorkManager               = "endpoint-workmgr"
	ServiceRegistries         = "endpoint-svcreg"
	EndpointComponentOperator = "endpoint-component-operator"
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
		By("Creating endpoint")
		endpoint := newEndpoint(testEndpointName, testNamespace)
		createNewUnstructured(clientHubDynamic, gvrEndpoint,
			endpoint, testEndpointName, testNamespace)

		When("endpoint created, wait for all component CRs to be created", func() {
			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait endpoint component operator...")
				_, err = clientHub.AppsV1().Deployments(testNamespace).Get(EndpointComponentOperator, metav1.GetOptions{})
				return err
			}, 4, 0.2).Should(BeNil())
			klog.V(1).Info("endpoint component operator created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait cert policy controller...")
				_, err = clientHubDynamic.Resource(gvrCertpoliciescontroller).Namespace(testNamespace).Get(CertPolicyController, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("cert policy controller created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait search controller...")
				_, err = clientHubDynamic.Resource(gvrSearchcollector).Namespace(testNamespace).Get(SearchCollector, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("search controller created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait policy controller...")
				_, err = clientHubDynamic.Resource(gvrPolicycontroller).Namespace(testNamespace).Get(PolicyController, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("policy controller created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait application manager...")
				_, err = clientHubDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(ApplicationManager, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("application manager created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait connection manager...")
				_, err = clientHubDynamic.Resource(gvrConnectionmanager).Namespace(testNamespace).Get(ConnectionManager, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("connection manager created")

			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait service registries...")
				_, err = clientHubDynamic.Resource(gvrServiceregistries).Namespace(testNamespace).Get(ServiceRegistries, metav1.GetOptions{})
				return err
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("service registries created")

		})
	})

	It("Should delete corresponding component CR", func() {
		By("Updating endpoint")
		_, err := clientHubDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Patch(testEndpointName, types.JSONPatchType, []byte(deletePatchString), metav1.PatchOptions{})
		Expect(err).To(BeNil())

		When("endpoint update, wait for corresponding component to create/delete", func() {
			Eventually(func() *unstructured.Unstructured {
				var objAppmgr *unstructured.Unstructured
				klog.V(1).Info("Wait application manager component...")
				objAppmgr, err = clientHubDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(ApplicationManager, metav1.GetOptions{})
				return objAppmgr
			}, 2, 0.2).Should(BeNil())
			klog.V(1).Info("application manager deleted")
		})
	})

	It("Should add corresponding component CR", func() {
		By("Updating endpoint")
		_, err := clientHubDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Patch(testEndpointName, types.JSONPatchType, []byte(addPatchString), metav1.PatchOptions{})
		Expect(err).To(BeNil())

		When("endpoint update, wait for corresponding component to create/delete", func() {
			Eventually(func() error {
				var err error
				klog.V(1).Info("Wait application manager...")
				_, err = clientHubDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(ApplicationManager, metav1.GetOptions{})
				return err
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("application manager created")
		})
	})

	It("Should delete all component CRs", func() {
		By("Deleteing endpoint")
		err := clientHubDynamic.Resource(gvrEndpoint).Namespace(testNamespace).Delete(testEndpointName, &metav1.DeleteOptions{})
		Expect(err).To(BeNil())

		When("endpoint deleted, wait for all components CR to be deleted", func() {
			Eventually(func() *unstructured.Unstructured {
				var objCertPolicyCtl *unstructured.Unstructured
				klog.V(1).Info("Wait cert policy controller...")
				objCertPolicyCtl, err = clientHubDynamic.Resource(gvrCertpoliciescontroller).Namespace(testNamespace).Get(CertPolicyController, metav1.GetOptions{})
				return objCertPolicyCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("cert policy controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objSearchCtl *unstructured.Unstructured
				klog.V(1).Info("Wait search controller...")
				objSearchCtl, err = clientHubDynamic.Resource(gvrSearchcollector).Namespace(testNamespace).Get(SearchCollector, metav1.GetOptions{})
				return objSearchCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("search controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objPolicyCtl *unstructured.Unstructured
				klog.V(1).Info("Wait policy controller...")
				objPolicyCtl, err = clientHubDynamic.Resource(gvrPolicycontroller).Namespace(testNamespace).Get(PolicyController, metav1.GetOptions{})
				return objPolicyCtl
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("policy controller deleted")

			Eventually(func() *unstructured.Unstructured {
				var objAppMgr *unstructured.Unstructured
				klog.V(1).Info("Wait application manager...")
				objAppMgr, err = clientHubDynamic.Resource(gvrApplicationmanager).Namespace(testNamespace).Get(ApplicationManager, metav1.GetOptions{})
				return objAppMgr
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("application manager deleted")

			Eventually(func() *unstructured.Unstructured {
				var objConnMgr *unstructured.Unstructured
				klog.V(1).Info("Wait connection manager...")
				objConnMgr, err = clientHubDynamic.Resource(gvrConnectionmanager).Namespace(testNamespace).Get(ConnectionManager, metav1.GetOptions{})
				return objConnMgr
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("connection manager deletedd")

			Eventually(func() *unstructured.Unstructured {
				var objServiceReg *unstructured.Unstructured
				klog.V(1).Info("Wait service registries...")
				objServiceReg, err = clientHubDynamic.Resource(gvrServiceregistries).Namespace(testNamespace).Get(ServiceRegistries, metav1.GetOptions{})
				return objServiceReg
			}, 5, 1).Should(BeNil())
			klog.V(1).Info("service registries deleted")

			Eventually(func() error {
				//var objComponentOperator *appsv1.Deployment
				klog.V(1).Info("Wait endpoint component operator...")
				_, err = clientHub.AppsV1().Deployments(testNamespace).Get(EndpointComponentOperator, metav1.GetOptions{})
				return err
			}, 5, 1).ShouldNot(BeNil())
			klog.V(1).Info("endpoint component operator deleted")
		})
	})
})
