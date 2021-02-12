// +build functional

package klusterlet_addon_controller_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

var cmaValidations = map[string][]string{
	clustermanagementaddon.ApplicationManager: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"application-manager"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.ApplicationManager].DisplayName + `"`,
	},
	clustermanagementaddon.CertPolicyController: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"cert-policy-controller"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.CertPolicyController].DisplayName + `"`,
	},
	clustermanagementaddon.IamPolicyController: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"iam-policy-controller"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.IamPolicyController].DisplayName + `"`,
	},
	clustermanagementaddon.PolicyController: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"policy-controller"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.PolicyController].DisplayName + `"`,
	},
	clustermanagementaddon.SearchCollector: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"search-collector"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.SearchCollector].DisplayName + `"`,
	},
	clustermanagementaddon.WorkManager: []string{
		`"kind":"ClusterManagementAddOn"`,
		`"name":"work-manager"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.WorkManager].DisplayName + `"`,
	},
}

var addOnPatchStrings = map[string]string{
	clustermanagementaddon.ApplicationManager: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Application",
	),
	clustermanagementaddon.CertPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Cert Policy",
	),
	clustermanagementaddon.IamPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Iam Policy",
	),
	clustermanagementaddon.PolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Policy",
	),
	clustermanagementaddon.SearchCollector: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Search",
	),
	clustermanagementaddon.WorkManager: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":\"%s\"}]",
		"replace", "/spec/addOnMeta/displayName", "Work",
	),
}

var _ = Describe("Creating ClusterManagementAddOn", func() {

	It("Should always create ClusterManagementAddOn for addons in klusterleraddonconfig", func() {
		By("Check all addons has clustermanagementaddons", func() {
			var err error
			Expect(err).Should(BeNil())
			for _, addonName := range clustermanagementaddon.ClusterManagementAddOnNames {
				var clusterManagementAddOn *unstructured.Unstructured
				Eventually(func() error {
					clusterManagementAddOn, err = clientClusterDynamic.Resource(gvrClusterManagementAddOn).Get(context.TODO(), addonName, metav1.GetOptions{})
					return err
				}, 5, 1).Should(BeNil())
				By("Validating " + addonName)
				validateUnstructured(clusterManagementAddOn, cmaValidations[addonName])
			}
		})

		By("Deleting one by one ClusterManagementAddOn", func() {
			var err error
			for _, addonName := range clustermanagementaddon.ClusterManagementAddOnNames {
				err = clientClusterDynamic.Resource(gvrClusterManagementAddOn).Delete(context.TODO(), addonName, metav1.DeleteOptions{})
				Expect(err).To(BeNil())
				By("Checking the ClusterManagementAddOn "+addonName+" is created back", func() {
					var clusterManagementAddOn *unstructured.Unstructured
					Eventually(func() error {
						clusterManagementAddOn, err = clientClusterDynamic.Resource(gvrClusterManagementAddOn).Get(context.TODO(), addonName, metav1.GetOptions{})
						return err
					}, 10, 1).Should(BeNil())
					By("Validating " + addonName)
					validateUnstructured(clusterManagementAddOn, cmaValidations[addonName])
				})
			}
		})

		By("Modifying one by one ClusterManagementAddOn", func() {
			var err error
			for _, addonName := range clustermanagementaddon.ClusterManagementAddOnNames {
				_, err = clientClusterDynamic.Resource(gvrClusterManagementAddOn).Patch(context.TODO(), addonName, types.JSONPatchType, []byte(addOnPatchStrings[addonName]), metav1.PatchOptions{})
				Expect(err).To(BeNil())
				time.Sleep(time.Second * 2)
				By("Checking the ClusterManagementAddOn "+addonName+" is reverted back to original Spec", func() {
					var clusterManagementAddOn *unstructured.Unstructured
					Eventually(func() error {
						clusterManagementAddOn, err = clientClusterDynamic.Resource(gvrClusterManagementAddOn).Get(context.TODO(), addonName, metav1.GetOptions{})
						return err
					}, 5, 1).Should(BeNil())
					By("Validating " + addonName)
					validateUnstructured(clusterManagementAddOn, cmaValidations[addonName])
				})
			}
		})
	})
})
