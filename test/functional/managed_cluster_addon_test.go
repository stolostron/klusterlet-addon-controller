// +build functional

// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package klusterlet_addon_controller_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var mcaMaps = map[string]string{
	applicationManager:   "application-manager",
	certPolicyController: "cert-policy-controller",
	iamPolicyController:  "iam-policy-controller",
	policyController:     "policy-controller",
	searchCollector:      "search-collector",
	workManager:          "work-manager",
}
var mcaValidations = map[string][]string{
	applicationManager: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.ApplicationManager].DisplayName + `"`,
	},
	certPolicyController: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.CertPolicyController].DisplayName + `"`,
	},
	iamPolicyController: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.IamPolicyController].DisplayName + `"`,
	},
	policyController: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.PolicyController].DisplayName + `"`,
	},
	searchCollector: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.SearchCollector].DisplayName + `"`,
	},
	workManager: []string{
		`"resource":"klusterletaddonconfigs"`,
		`"group":"agent.open-cluster-management.io"`,
		`"name":"` + testKlusterletAddonConfigName + `"`,
		`"crName":"` + testKlusterletAddonConfigName + `"`,
		`"crdName":"klusterletaddonconfigs.agent.open-cluster-management.io"`,
		`"displayName":"` + clustermanagementaddon.ClusterManagementAddOnMap[clustermanagementaddon.WorkManager].DisplayName + `"`,
	},
}

var _ = Describe("ManagedClusterAddOns", func() {
	var managedCluster, klusterletAddonConfig *unstructured.Unstructured
	BeforeEach(func() {
		By("Cleanup old test data", func() {
			cleanUpTestData(clientClusterDynamic)
		})
		By("Creating KlusterletAddonConfig & ManagedCluster", func() {
			managedCluster = newManagedCluster(testKlusterletAddonConfigName, testNamespace)
			createNewUnstructured(clientClusterDynamic, gvrManagedCluster,
				managedCluster, testKlusterletAddonConfigName, "")
			klusterletAddonConfig = newKlusterletAddonConfig(testKlusterletAddonConfigName, testNamespace, validVersion)
			createNewUnstructured(clientClusterDynamic, gvrKlusterletAddonConfig,
				klusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
			setClusterOnline(clientClusterDynamic, testKlusterletAddonConfigName)
		})
		var err error
		ownerKlusterletAddonConfig, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testKlusterletAddonConfigName).Get(context.TODO(), testKlusterletAddonConfigName, metav1.GetOptions{})
		Expect(err).Should(BeNil())
		By("Checking manifestwork of CRDs is created", func() {
			var crds *unstructured.Unstructured
			Eventually(func() error {
				crds, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), allCRDs, metav1.GetOptions{})
				return err
			}, 5, 1).Should(BeNil())
			validateUnstructured(crds, validations[allCRDs])
			Expect(isOwner(ownerKlusterletAddonConfig, crds)).Should(BeTrue(), "OwnerRef of "+allCRDs+" should be set correctly")
		})
		By("Updating manifestwork of CRDs with all applied.", func() {
			setManifestWorkStatusAvailable(clientClusterDynamic, allCRDs, testNamespace)
			time.Sleep(time.Second * 30)
		})
	})
	It("Should create ManagedClusterAddOn Operator for enabled addons, delete when disabled, delete all when klusterletaddonconfig is deleted", func() {
		By("Check all addons has managedclusteraddons", func() {
			var err error
			ownerKlusterletAddonConfig, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testKlusterletAddonConfigName).Get(context.TODO(), testKlusterletAddonConfigName, metav1.GetOptions{})
			Expect(err).Should(BeNil())
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				By("Checking " + mcaName)
				var mca *unstructured.Unstructured
				Eventually(func() error {
					mca, err = clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), mcaName, metav1.GetOptions{})
					return err
				}, 10, 2).Should(BeNil())
				By("Validating " + mcaName)
				validateUnstructured(mca, mcaValidations[mcaName])
				Expect(isOwner(ownerKlusterletAddonConfig, mca)).Should(BeTrue(), "OwnerRef of "+mcaName+" should be set correctly")
			}
		})
		By("If managedclusteraddons are deleted accidentally, should recreate", func() {
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				By("Deleting " + mcaName)
				Expect(func() error {
					return clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Delete(context.TODO(), mcaName, metav1.DeleteOptions{})
				}()).Should(BeNil())
			}
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				By("Checking " + mcaName)
				Eventually(func() error {
					_, err := clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), mcaName, metav1.GetOptions{})
					return err
				}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
			}
		})
		By("Disabling one by one, and managedclusteraddons should be deleted", func() {
			var err error
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				// workmgr is always enabled
				if addon == workManager {
					continue
				}
				By("Checking the managedclusteraddon "+mcaName+" exists", func() {
					Eventually(func() error {
						_, err = clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), mcaName, metav1.GetOptions{})
						return err
					}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
				})
				By("Disabling " + addon)
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Patch(context.TODO(), testKlusterletAddonConfigName, types.JSONPatchType, []byte(deletePatchStrings[addon]), metav1.PatchOptions{})
				Expect(err).To(BeNil())
				By("Checking the Managedclusteraddon "+mcaName+" is created", func() {
					eventuallyNotFound(clientClusterDynamic, gvrManagedClusterAddOn, mcaName, testNamespace)
				})
			}
		})
		By("Enabling one by one, and managedclusteraddons should be removed", func() {
			var err error
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				// workmgr is always enabled
				if addon == workManager {
					continue
				}
				By("Enabling " + addon)
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Patch(context.TODO(), testKlusterletAddonConfigName, types.JSONPatchType, []byte(addPatchStrings[addon]), metav1.PatchOptions{})
				Expect(err).To(BeNil())
				By("Checking the Managedclusteraddon "+mcaName+" is created", func() {
					Eventually(func() error {
						_, err = clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), mcaName, metav1.GetOptions{})
						return err
					}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
				})

				By("Checking the Managedclusteraddon "+mcaName+" has install namespace and registration config set", func() {
					Eventually(func() bool {
						managedClusterAddOn, err := clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), mcaName, metav1.GetOptions{})
						if err != nil {
							return false
						}
						installNamespace, _, _ := unstructured.NestedString(managedClusterAddOn.Object, "spec", "installNamespace")
						if installNamespace != addonoperator.KlusterletAddonNamespace {
							return false
						}
						registrationConfigs, _, err := unstructured.NestedSlice(managedClusterAddOn.Object, "status", "registrations")
						return len(registrationConfigs) == 1
					}, eventuallyTimeout, eventuallyInterval).Should(BeTrue())
				})
			}
		})
		By("Deleting klusterletaddonconfig, and all managedclusteraddons should be deleted", func() {
			Expect(func() error {
				return clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
			}()).Should(BeNil())
			for _, addon := range addonCRs {
				mcaName := mcaMaps[addon]
				By("Checking the Managedclusteraddon "+mcaName+" is removed", func() {
					eventuallyNotFound(clientClusterDynamic, gvrManagedClusterAddOn, mcaName, testNamespace)
				})
			}
		})
	})

	It("Should show correct Progressing condition status", func() {
		// add finalizers to manifestworks
		for _, crName := range addonCRs {
			addFinalizerToManifestWork(clientClusterDynamic, crName, testNamespace)
		}
		By("Checking Progressing=True when manifestwork is installing", func() {
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				checkStatusCondition(clientClusterDynamic, mcaName, testNamespace, "Progressing", "True")
			}
		})
		By("Updating manifestwork with 1 failed, other succeeded. Checking Progressing=False when manifestwork not finished applying", func() {
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				setManifestWorkAppliedStatus(clientClusterDynamic, crName, testNamespace, 4, 1)
				checkStatusCondition(clientClusterDynamic, mcaName, testNamespace, "Progressing", "False")
			}
		})
		By("Updating manifestwork with all applied. Checking Progressing=False when manifestwork finished applying", func() {
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				setManifestWorkAppliedStatus(clientClusterDynamic, crName, testNamespace, 5, 0)
				checkStatusCondition(clientClusterDynamic, mcaName, testNamespace, "Progressing", "False")
			}
		})

		By("Deleting klusterletaddonconfig. Checking Progressing=True when manifestwork not Deleted", func() {
			Expect(func() error {
				return clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
			}()).Should(BeNil())
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				checkStatusCondition(clientClusterDynamic, mcaName, testNamespace, "Progressing", "True")
			}
		})
		// removing finalizers
		for _, crName := range addonCRs {
			deleteFinalizerOfManifestWork(clientClusterDynamic, crName, testNamespace)
		}
	})
	It("Should show correct Degraded condition status", func() {
		By("Checking Degraded=true when manifestwork failed to install", func() {
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				setManifestWorkAppliedStatus(clientClusterDynamic, crName, testNamespace, 4, 1)
				checkStatusCondition(clientClusterDynamic, mcaName, testNamespace, "Degraded", "True")
			}
		})
		By("Checking Degraded removed when everything looks fine", func() {
			for _, crName := range addonCRs {
				mcaName := mcaMaps[crName]
				setManifestWorkAppliedStatus(clientClusterDynamic, crName, testNamespace, 5, 0)
				checkStatusConditionNotFound(clientClusterDynamic, mcaName, testNamespace, "Degraded", "True")
			}
		})
	})
})

func generateManifestStatus(ordinal int, applied string, message string, reason string) string {
	return fmt.Sprintf(`{"conditions":[{"lastTransitionTime":"2021-03-31T14:46:27Z","type":"Applied","status":"%s", "message":"%s", "reason":"%s"}],"resourceMeta":{"ordinal":%d}}`, applied, message, reason, ordinal)
}
func setManifestWorkAppliedStatus(clientHubDynamic dynamic.Interface, name, namespace string, succeed int, failed int) {
	ordinal := 0
	patchString := `{"status":{"resourceStatus":{"manifests":[`

	for i := 0; i < failed; i++ {
		ordinal = i
		s := generateManifestStatus(ordinal, "False", "Failed to apply manifestwork", "ManifestDegraded")
		if i < failed-1 || succeed > 0 {
			s = s + ","
		}
		patchString = patchString + s
	}
	for i := 0; i < succeed; i++ {
		ordinal = i + failed
		s := generateManifestStatus(ordinal, "True", "Manifestwork Applied", "ManifestApplied")
		if i < succeed-1 {
			s = s + ","
		}
		patchString = patchString + s
	}

	patchString = patchString + `]}}}`

	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Patch(context.TODO(), name, types.MergePatchType, []byte(patchString), metav1.PatchOptions{}, "status")
		return err
	}()).Should(BeNil())
}
func hasStatusHelper(lientHubDynamic dynamic.Interface, name, namespace, condType, condStatus string) error {
	mca, err := clientClusterDynamic.Resource(gvrManagedClusterAddOn).Namespace(testNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	retErr := fmt.Errorf("failed to get expected status. expect %s only 1 type %s and status %s in %v",
		name, condType, condStatus, mca.Object["status"])
	status, ok := mca.Object["status"]
	if !ok {
		return retErr
	}
	s, ok := status.(map[string]interface{})
	if !ok {
		return retErr
	}
	conditions, ok := s["conditions"]
	if !ok {
		return retErr
	}
	clist, ok := conditions.([]interface{})
	if !ok {
		return retErr
	}
	countType := 0
	countMatch := 0
	for _, c := range clist {
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}
		checkType := `"type":"` + condType + `"`
		checkStatus := `"status":"` + condStatus + `"`
		if strings.Contains(string(b), checkType) {
			countType++
			if strings.Contains(string(b), checkStatus) {
				countMatch++
			}
		}
	}
	if countType == 1 && countMatch == 1 {
		return nil
	}
	return retErr
}

func checkStatusCondition(clientHubDynamic dynamic.Interface, name, namespace, condType, condStatus string) {
	Eventually(func() error { return hasStatusHelper(clientHubDynamic, name, namespace, condType, condStatus) }, eventuallyTimeout, eventuallyInterval).Should(BeNil())
}

func checkStatusConditionNotFound(clientHubDynamic dynamic.Interface, name, namespace, condType, condStatus string) {
	Eventually(func() error {
		if err := hasStatusHelper(clientHubDynamic, name, namespace, condType, condStatus); err == nil {
			return fmt.Errorf("Expected to not found any condition %s=%s", condType, condStatus)
		}
		return nil
	}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
}
