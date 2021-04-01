// +build functional

// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package klusterlet_addon_controller_test

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

const (
	testKlusterletAddonConfigName = testNamespace
)

const (
	applicationManager      = testKlusterletAddonConfigName + "-klusterlet-addon-appmgr"
	certPolicyController    = testKlusterletAddonConfigName + "-klusterlet-addon-certpolicyctrl"
	iamPolicyController     = testKlusterletAddonConfigName + "-klusterlet-addon-iampolicyctrl"
	policyController        = testKlusterletAddonConfigName + "-klusterlet-addon-policyctrl"
	searchCollector         = testKlusterletAddonConfigName + "-klusterlet-addon-search"
	workManager             = testKlusterletAddonConfigName + "-klusterlet-addon-workmgr"
	allCRDs                 = testKlusterletAddonConfigName + "-klusterlet-addon-crds"
	klusterletAddonOperator = testKlusterletAddonConfigName + "-klusterlet-addon-operator"
	validVersion            = "2.0.0"
)

const (
	klusterletAddonFinalizer = "agent.open-cluster-management.io/klusterletaddonconfig-cleanup"
	manifestWorkFinalizer    = "cluster.open-cluster-management.io/manifest-work-cleanup"
)

var deletePatchStrings = map[string]string{
	applicationManager: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/applicationManager/enabled", false,
	),
	certPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/certPolicyController/enabled", false,
	),
	iamPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/iamPolicyController/enabled", false,
	),
	policyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/policyController/enabled", false,
	),
	searchCollector: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/searchCollector/enabled", false,
	),
}

var addPatchStrings = map[string]string{
	applicationManager: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/applicationManager/enabled", true,
	),
	certPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/certPolicyController/enabled", true,
	),
	iamPolicyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/iamPolicyController/enabled", true,
	),
	policyController: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/policyController/enabled", true,
	),
	searchCollector: fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%t}]",
		"replace", "/spec/searchCollector/enabled", true,
	),
}

// list of manifestwork name for addon crs
var addonCRs = []string{
	applicationManager,
	certPolicyController,
	iamPolicyController,
	policyController,
	searchCollector,
	workManager,
}

// list of regex we will use to validate json from the manifestwork
var validations = map[string][]string{
	applicationManager: []string{
		`"kind":"ApplicationManager"`,
		`"name":"klusterlet-addon-appmgr"`,
		`"kubeconfig":`,
		`"name":"appmgr-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	certPolicyController: []string{
		`"kind":"CertPolicyController"`,
		`"name":"klusterlet-addon-certpolicyctrl"`,
		`"kubeconfig":`,
		`"name":"certpolicyctrl-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	iamPolicyController: []string{
		`"kind":"IAMPolicyController"`,
		`"name":"klusterlet-addon-iampolicyctrl"`,
		`"kubeconfig":`,
		`"name":"iampolicyctrl-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	policyController: []string{
		`"kind":"PolicyController"`,
		`"name":"klusterlet-addon-policyctrl"`,
		`"kubeconfig":`,
		`"name":"policyctrl-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	searchCollector: []string{
		`"kind":"SearchCollector"`,
		`"name":"klusterlet-addon-search"`,
		`"kubeconfig":`,
		`"name":"search-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	workManager: []string{
		`"kind":"WorkManager"`,
		`"name":"klusterlet-addon-workmgr"`,
		`"kubeconfig":`,
		`"name":"workmgr-hub-kubeconfig"`,
		"klusterlet-addon-lease-controller@sha256",
	},
	klusterletAddonOperator: []string{
		`"kind":"Deployment"`,
		`"name":"klusterlet-addon-operator"`,
		`"kind":"Secret"`,
		`"kubernetes.io/dockerconfigjson"`,
	},
	allCRDs: []string{
		`"name":"applicationmanagers.agent.open-cluster-management.io"`,
		`"name":"certpolicycontrollers.agent.open-cluster-management.io"`,
		`"name":"iampolicycontrollers.agent.open-cluster-management.io"`,
		`"name":"policycontrollers.agent.open-cluster-management.io"`,
		`"name":"searchcollectors.agent.open-cluster-management.io"`,
		`"name":"workmanagers.agent.open-cluster-management.io"`,
		`"rbac.authorization.k8s.io/aggregate-to-admin":"true"`,
	},
}

// list of regex we need to check
var _ = Describe("Creating KlusterletAddonConfig", func() {
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
		})
	})
	Context("Cluster online", func() {
		BeforeEach(func() {
			setClusterOnline(clientClusterDynamic, testKlusterletAddonConfigName)
		})
		It("Should always add finalizer on ManagedCluster & KlusterletAddonConfig no matter offline or online", func() {
			By("Checking finalizers set when online")
			checkFinalizerIsSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testKlusterletAddonConfigName, klusterletAddonFinalizer)
			By("Setting managedcluster offline")
			setClusterOffline(clientClusterDynamic, testKlusterletAddonConfigName)
			By("Checking finalizer on ManagedCluster and KlusterletAddonConfig are still kept")
			checkFinalizerIsSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testKlusterletAddonConfigName, klusterletAddonFinalizer)
			By("Setting managedcluster online & checking finalizer added")
			setClusterOnline(clientClusterDynamic, testKlusterletAddonConfigName)
			checkFinalizerIsSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testKlusterletAddonConfigName, klusterletAddonFinalizer)
		})
		It("Should create ManifestWorks for CRDs, Addon Operator, and CRs; should set OwnerRef=klusterletaddonconfig for created Manifestworks", func() {
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
			})
			time.Sleep(30 * time.Second)
			By("Checking manifestwork of Addon Operator is created", func() {
				var addon *unstructured.Unstructured
				Eventually(func() error {
					addon, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), klusterletAddonOperator, metav1.GetOptions{})
					return err
				}, 5, 1).Should(BeNil())
				validateUnstructured(addon, validations[klusterletAddonOperator])
				Expect(isOwner(ownerKlusterletAddonConfig, addon)).Should(BeTrue(), "OwnerRef of "+klusterletAddonOperator+" should be set correctly")
			})

			By("Checking manifestwork of all CRs are created", func() {
				for _, crName := range addonCRs {
					By("Checking " + crName)
					var cr *unstructured.Unstructured
					Eventually(func() error {
						cr, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), crName, metav1.GetOptions{})
						return err
					}, 5, 1).Should(BeNil())
					By("Validating " + crName)
					validateUnstructured(cr, validations[crName])
					Expect(isOwner(ownerKlusterletAddonConfig, cr)).Should(BeTrue(), "OwnerRef of "+crName+" should be set correctly")
				}
			})
		})
	})
	Context("Cluster offline", func() {
		BeforeEach(func() {
			setClusterOffline(clientClusterDynamic, testKlusterletAddonConfigName)
		})
		It("Should also create Manifestworks for CRDs, Addon Operator; should set OwnerRef=klusterletaddonconfig for created Manifestworks", func() {
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

			By("Checking manifestwork of Addon Operator is created", func() {
				var addon *unstructured.Unstructured
				Eventually(func() error {
					addon, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), klusterletAddonOperator, metav1.GetOptions{})
					return err
				}, 5, 1).Should(BeNil())
				validateUnstructured(addon, validations[klusterletAddonOperator])
				Expect(isOwner(ownerKlusterletAddonConfig, addon)).Should(BeTrue(), "OwnerRef of "+klusterletAddonOperator+" should be set correctly")
			})
		})
	})
})

var _ = Describe("Disabling & Enabling Addons", func() {
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
	})
	It("Should update Addons' Manifestwork when KlusterletAddonConfig changed", func() {
		By("Shaffling addon orders")
		tmpAddonCRs := append([]string{}, addonCRs...)
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(tmpAddonCRs), func(i, j int) { tmpAddonCRs[i], tmpAddonCRs[j] = tmpAddonCRs[j], tmpAddonCRs[i] })
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
		})
		time.Sleep(30 * time.Second)
		By("Disabling all addons one by one", func() {
			for _, addon := range tmpAddonCRs {
				// workmgr is always enabled
				if addon == workManager {
					continue
				}
				By("Checking the Manifestwork "+addon+" exists", func() {
					Eventually(func() error {
						_, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), addon, metav1.GetOptions{})
						return err
					}, 5, 1).Should(BeNil())
				})
				By("Disabling " + addon)
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Patch(context.TODO(), testKlusterletAddonConfigName, types.JSONPatchType, []byte(deletePatchStrings[addon]), metav1.PatchOptions{})
				Expect(err).To(BeNil())
				By("Checking the Manifestwork "+addon+" is removed", func() {
					eventuallyNotFound(clientClusterDynamic, gvrManifestwork, addon, testNamespace)
				})
			}
		})

		By("Checking the CRDs are not changed", func() {
			Consistently(func() error {
				crds, err := clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), allCRDs, metav1.GetOptions{})
				if err != nil {
					return err
				}
				validateUnstructured(crds, validations[allCRDs])
				return nil
			}, 3, 1).Should(BeNil())
		})
		By("Shaffling addon orders")
		rand.Shuffle(len(tmpAddonCRs), func(i, j int) { tmpAddonCRs[i], tmpAddonCRs[j] = tmpAddonCRs[j], tmpAddonCRs[i] })
		By("Enabling all Addons one by one", func() {
			for _, addon := range tmpAddonCRs {
				// workmgr is always enabled
				if addon == workManager {
					continue
				}
				By("Enabling " + addon)
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Patch(context.TODO(), testKlusterletAddonConfigName, types.JSONPatchType, []byte(addPatchStrings[addon]), metav1.PatchOptions{})
				Expect(err).To(BeNil())
				By("Checking the Manifestwork "+addon+" is created", func() {
					Eventually(func() error {
						_, err = clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), addon, metav1.GetOptions{})
						return err
					}, 5, 1).Should(BeNil())
				})
			}
		})
	})
})

var _ = Describe("Deleting Managedcluster Which Has Never Been Online", func() {
	var managedCluster, klusterletAddonConfig *unstructured.Unstructured
	BeforeEach(func() {
		By("Cleanup old test data", func() {
			cleanUpTestData(clientClusterDynamic)
		})
		By("Creating KlusterletAddonConfig & ManagedCluster (offline)", func() {
			managedCluster = newManagedCluster(testKlusterletAddonConfigName, testNamespace)
			createNewUnstructured(clientClusterDynamic, gvrManagedCluster,
				managedCluster, testKlusterletAddonConfigName, "")
			setClusterOffline(clientClusterDynamic, testKlusterletAddonConfigName)
			klusterletAddonConfig = newKlusterletAddonConfig(testKlusterletAddonConfigName, testNamespace, validVersion)
			createNewUnstructured(clientClusterDynamic, gvrKlusterletAddonConfig,
				klusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
		})
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
		})
		By("Deleting ManagedCluster", func() {
			Expect(func() error {
				return clientClusterDynamic.Resource(gvrManagedCluster).Namespace("").Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
			}()).Should(BeNil())
		})
	})
	It("Should not leave any ManifestWorks; should remove KlusterletAddonConfig; should remove finalizer on ManagedCluster", func() {
		By("Checking CR/CRD/Addon Operator's Manifestworks are deleted", func() {
			for _, crName := range addonCRs {
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, crName, testNamespace)
			}
			eventuallyNotFound(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace)
			eventuallyNotFound(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace)
		})

		By("Checking KlusterletAddonConfig & Managedcluster are deleted", func() {
			eventuallyNotFound(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "")
			eventuallyNotFound(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
		})
	})
})

var _ = Describe("Deleting A Joined ManagedCluster", func() {
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
		})
		time.Sleep(30 * time.Second)
		// wait for ManagedCluster to have finalizer
		By("Waiting for reconcile", func() {
			checkFinalizerIsSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testKlusterletAddonConfigName, klusterletAddonFinalizer)
		})

		// adding finalizers to manifestworks
		By("Adding finalizers to manifestworks", func() {
			addFinalizerToManifestWork(clientClusterDynamic, allCRDs, testNamespace)
			addFinalizerToManifestWork(clientClusterDynamic, klusterletAddonOperator, testNamespace)
			for _, crName := range addonCRs {
				addFinalizerToManifestWork(clientClusterDynamic, crName, testNamespace)
			}
		})

	})
	Context("Cluster Online", func() {
		BeforeEach(func() {
			setClusterOnline(clientClusterDynamic, testKlusterletAddonConfigName)
			// delete Managedcluster
			By("Deleting ManagedCluster", func() {
				Expect(func() error {
					return clientClusterDynamic.Resource(gvrManagedCluster).Namespace("").Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
				}()).Should(BeNil())
			})
		})
		It("Should remove all Manifestworks for Addon CRs before removing Manifestworks for CRDs and Addon Operator", func() {
			By("Checking deletion timestamp are set for all CRs", func() {
				Eventually(func() error {
					for _, crName := range addonCRs {
						if err := checkDeletionTimestampIsSet(clientClusterDynamic, gvrManifestwork, crName, testNamespace); err != nil {
							return err
						}
					}
					return nil
				}, 5, 1).Should(BeNil())
			})

			By("Checking deletion timestamp empty for klusterlet addon operator", func() {
				Consistently(func() error {
					return checkDeletionTimestampIsNotSet(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace)
				}, 5, 1).Should(BeNil())
			})

			By("Checking deletion timestamp empty for CRDs", func() {
				Consistently(func() error {
					return checkDeletionTimestampIsNotSet(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace)
				}, 5, 1).Should(BeNil())
			})
		})
		It("Should remove all Manifestworks", func() {
			By("Force removing all finalizers in Manifestworks", func() {
				for _, crName := range addonCRs {
					deleteFinalizerOfManifestWork(clientClusterDynamic, crName, testNamespace)
				}
				deleteFinalizerOfManifestWork(clientClusterDynamic, klusterletAddonOperator, testNamespace)
				deleteFinalizerOfManifestWork(clientClusterDynamic, allCRDs, testNamespace)
			})

			By("Checking CR/CRD/Addon Operator's Manifestworks are deleted", func() {
				for _, crName := range addonCRs {
					eventuallyNotFound(clientClusterDynamic, gvrManifestwork, crName, testNamespace)
				}
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace)
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace)
			})
		})
	})
	Context("Cluster Offline", func() {
		BeforeEach(func() {
			setClusterOffline(clientClusterDynamic, testKlusterletAddonConfigName)
			// delete Managedcluster
			By("Deleting ManagedCluster", func() {
				Expect(func() error {
					return clientClusterDynamic.Resource(gvrManagedCluster).Namespace("").Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
				}()).Should(BeNil())
			})
		})
		It("Should force remove all Manifestworks; should remove finalizers on KlusterletAddonConfig &  ManagedCluster", func() {
			By("Checking all Manifestworks force deleted", func() {
				for _, crName := range addonCRs {
					eventuallyNotFound(clientClusterDynamic, gvrManifestwork, crName, testNamespace)
				}
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace)
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace)
			})

			By("Checking klusterletAddonConfig & ManagedCluster are deleted", func() {
				eventuallyNotFound(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "")
				eventuallyNotFound(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
			})

		})
	})
})

var _ = Describe("Deleting KlusterletAddonConfig Only", func() {
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
		// wait for ManagedCluster to have finalizer
		By("Waiting for reconcile", func() {
			checkFinalizerIsSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testKlusterletAddonConfigName, klusterletAddonFinalizer)
		})

		// adding finalizers to manifestworks
		By("Adding finalizers to manifestworks", func() {
			addFinalizerToManifestWork(clientClusterDynamic, allCRDs, testNamespace)
			addFinalizerToManifestWork(clientClusterDynamic, klusterletAddonOperator, testNamespace)
			for _, crName := range addonCRs {
				addFinalizerToManifestWork(clientClusterDynamic, crName, testNamespace)
			}
			//check finalizers are set
			checkFinalizerIsSet(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace, manifestWorkFinalizer)
			checkFinalizerIsSet(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace, manifestWorkFinalizer)
			for _, crName := range addonCRs {
				checkFinalizerIsSet(clientClusterDynamic, gvrManifestwork, crName, testNamespace, manifestWorkFinalizer)
			}

		})

		By("Delete KlusterletAddonConfig", func() {
			Expect(func() error {
				return clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testNamespace).Delete(context.TODO(), testKlusterletAddonConfigName, metav1.DeleteOptions{})
			}()).Should(BeNil())
		})

	})
	It("Should remove all Manifestworks and remove finalizer on ManagedCluster", func() {
		By("Force removing all finalizers in Manifestworks", func() {
			for _, crName := range addonCRs {
				deleteFinalizerOfManifestWork(clientClusterDynamic, crName, testNamespace)
			}
			deleteFinalizerOfManifestWork(clientClusterDynamic, klusterletAddonOperator, testNamespace)
			deleteFinalizerOfManifestWork(clientClusterDynamic, allCRDs, testNamespace)
		})
		By("Checking CR/CRD/klusterlet Addon Operator's Manifestworks removed", func() {
			for _, crName := range addonCRs {
				eventuallyNotFound(clientClusterDynamic, gvrManifestwork, crName, testNamespace)
			}
			eventuallyNotFound(clientClusterDynamic, gvrManifestwork, klusterletAddonOperator, testNamespace)
			eventuallyNotFound(clientClusterDynamic, gvrManifestwork, allCRDs, testNamespace)
		})
		By("Checking KlusterletAddonConfig is deleted", func() {
			eventuallyNotFound(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
		})
		By("Checking managedcluster finalizer is removed", func() {
			checkFinalizerIsNotSet(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "", klusterletAddonFinalizer)
		})
	})
})

func checkFinalizerIsSet(clientClusterDynamic dynamic.Interface, gvr schema.GroupVersionResource, name string, namespace string, value string) {
	Eventually(func() error {
		obj, err := clientClusterDynamic.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		finalizers := obj.GetFinalizers()
		for _, finalizer := range finalizers {
			if finalizer == value {
				return nil
			}
		}
		return fmt.Errorf("Failed to get finalizer %s in %s %s", value, gvr.Resource, name)
	}, 10, 1).Should(BeNil())
}
func checkFinalizerIsNotSet(clientClusterDynamic dynamic.Interface, gvr schema.GroupVersionResource, name string, namespace string, value string) {
	Eventually(func() error {
		obj, err := clientClusterDynamic.Resource(gvr).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		finalizers := obj.GetFinalizers()
		for _, finalizer := range finalizers {
			if finalizer == value {
				return fmt.Errorf("Should not get finalizer %s in %s %s", value, gvr.Resource, name)
			}
		}
		return nil
	}, 10, 1).Should(BeNil())
}

func cleanUpTestData(clientClusterDynamic dynamic.Interface) {
	klog.V(1).Info("Deleting KlusterletAddonConfig & ManagedCluster")
	deleteIfExists(clientClusterDynamic, gvrKlusterletAddonConfig, testKlusterletAddonConfigName, testNamespace)
	deleteIfExists(clientClusterDynamic, gvrManagedCluster, testKlusterletAddonConfigName, "")

	klog.V(1).Info("Deleting all Manifestworks")

	ns := clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace)
	if items, err := ns.List(context.TODO(), metav1.ListOptions{}); err == nil {
		for _, item := range items.Items {
			klog.V(2).Info("Deleting " + item.GetName())
			deleteIfExists(clientClusterDynamic, gvrManifestwork, item.GetName(), testNamespace)
		}
	}

}

func setClusterOnline(clientHubDynamic dynamic.Interface, name string) {
	// patch and set online condition
	patchString := `{"status":` +
		`{"conditions":[` +
		`{"type":"ManagedClusterJoined","lastTransitionTime":"2020-01-01T01:01:01Z","message":"Managed Cluster joined","status":"True","reason":"ManagedClusterJoined"}` + `,` +
		`{"type":"ManagedClusterConditionAvailable","lastTransitionTime":"2020-01-01T01:01:01Z","message":"Managed Cluster Available","status":"True","reason":"ManagedClusterConditionAvailable"}` + `,` +
		`{"type":"HubAcceptedManaged","lastTransitionTime":"2020-01-01T01:01:01Z","message":"Accepted by hub cluster admin","status":"True","reason":"HubClusterAdminAccepted"}` +
		`]}}`
	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManagedCluster).Namespace("").Patch(context.TODO(), name, types.MergePatchType, []byte(patchString), metav1.PatchOptions{}, "status")
		return err
	}()).Should(BeNil())
}

func setClusterOffline(clientHubDynamic dynamic.Interface, name string) {
	// add place holder
	patchString := `{"status":` +
		`{"conditions":[` +
		`{"type":"placeholder","lastTransitionTime":"2020-01-01T01:01:01Z","message":"placeholder","status":"False","reason":"placeholder"}` +
		`]}}`
	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManagedCluster).Namespace("").Patch(context.TODO(), name, types.MergePatchType, []byte(patchString), metav1.PatchOptions{}, "status")
		return err
	}()).Should(BeNil())
	// patch and set no online condition
	patchString = fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%s}]",
		"replace", "/status/conditions", `[{"type":"placeholder","lastTransitionTime":"2020-01-01T01:01:01Z","message":"placeholder","status":"False","reason":"placeholder"}]`,
	)
	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManagedCluster).Namespace("").Patch(context.TODO(), name, types.JSONPatchType, []byte(patchString), metav1.PatchOptions{}, "status")
		return err
	}()).Should(BeNil())
}

func addFinalizerToManifestWork(clientHubDynamic dynamic.Interface, name string, namespace string) {
	// wait to see it
	Eventually(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		return err
	}, 5, 1).Should(BeNil())

	// patch and add finalizer
	patchString := fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%s}]",
		"add", "/metadata/finalizers", "[\"cluster.open-cluster-management.io/manifest-work-cleanup\"]",
	)
	// expect patch to work
	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Patch(context.TODO(), name, types.JSONPatchType, []byte(patchString), metav1.PatchOptions{})
		return err
	}()).Should(BeNil())
}

func deleteFinalizerOfManifestWork(clientHubDynamic dynamic.Interface, name string, namespace string) {
	// wait to see it
	Eventually(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		return err
	}, 5, 1).Should(BeNil())

	// patch and add finalizer
	patchString := fmt.Sprintf(
		"[{\"op\":\"%s\",\"path\":\"%s\",\"value\":%s}]",
		"replace", "/metadata/finalizers", "[]",
	)
	// expect patch to work
	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Patch(context.TODO(), name, types.JSONPatchType, []byte(patchString), metav1.PatchOptions{})
		return err
	}()).Should(BeNil())
}

func validateUnstructured(obj *unstructured.Unstructured, regexps []string) {
	resources, err := obj.MarshalJSON()
	Expect(err).To(BeNil())
	for _, r := range regexps {
		Expect(string(resources)).To(MatchRegexp(r))
	}
}

func checkDeletionTimestampIsSet(clientHubDynamic dynamic.Interface, gvr schema.GroupVersionResource, name string, namespace string) error {
	mw, err := clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if mw.GetDeletionTimestamp() == nil {
		return fmt.Errorf("Deletion timestamp of %s is not set", name)
	}
	return nil
}
func checkDeletionTimestampIsNotSet(clientHubDynamic dynamic.Interface, gvr schema.GroupVersionResource, name string, namespace string) error {
	mw, err := clientClusterDynamic.Resource(gvrManifestwork).Namespace(testNamespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if mw.GetDeletionTimestamp() != nil {
		return fmt.Errorf("Deletion timestamp of %s is set", name)
	}
	return nil
}

func eventuallyNotFound(clientHubDynamic dynamic.Interface, gvr schema.GroupVersionResource, name string, namespace string) {
	Eventually(func() error {
		_, err := clientClusterDynamic.Resource(gvrManifestwork).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err == nil {
			return fmt.Errorf("%s is not deleted", name)
		} else if !errors.IsNotFound(err) {
			return err
		}
		return nil
	}, 10, 1).Should(BeNil())
}

// isOwner checks if obj is owned by owner, obj can either be unstructured or ObjectMeta
func isOwner(owner *unstructured.Unstructured, obj interface{}) bool {
	if obj == nil {
		return false
	}
	var owners []metav1.OwnerReference
	objMeta, ok := obj.(*metav1.ObjectMeta)
	if ok {
		owners = objMeta.GetOwnerReferences()
	} else {
		if objUnstructured, ok := obj.(*unstructured.Unstructured); ok {
			owners = objUnstructured.GetOwnerReferences()
		} else {
			klog.Error("Failed to get owners")
			return false
		}
	}

	for _, ownerRef := range owners {
		if _, ok := owner.Object["metadata"]; !ok {
			klog.Error("no meta")
			continue
		}
		meta, ok := owner.Object["metadata"].(map[string]interface{})
		if !ok || meta == nil {
			klog.Error("no meta map")
			continue
		}
		name, ok := meta["name"].(string)
		if !ok || name == "" {
			klog.Error("failed to get name")
			continue
		}
		if ownerRef.Kind == owner.Object["kind"] && ownerRef.Name == name {
			return true
		}
	}
	return false
}

func setManifestWorkStatusAvailable(clientHubDynamic dynamic.Interface, name, namespace string) {
	patchString := `{"status":{"conditions":[{"lastTransitionTime":"2021-03-31T14:46:27Z","type":"Available","status":"True","message":"All resources are available","reason":"ResourcesAvailable"`
	//	patchString = patchString + `"lastTransitionTime":` + metav1.Time{Time: time.Now()}
	patchString = patchString + `}]}}`

	Expect(func() error {
		_, err := clientHubDynamic.Resource(gvrManifestwork).Namespace(namespace).Patch(context.TODO(), name, types.MergePatchType, []byte(patchString), metav1.PatchOptions{}, "status")
		return err
	}()).Should(BeNil())
}
