// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/common"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("managedClusterAddon test", func() {
	var managedClusterName string
	var managedClusterAnnotations map[string]string

	AfterEach(func() {
		deleteManagedCluster(managedClusterName)
	})

	BeforeEach(func() {
		managedClusterName = fmt.Sprintf("cluster-test-%s", rand.String(6))
	})

	JustBeforeEach(func() {
		By(fmt.Sprintf("Create managed cluster %s", managedClusterName), func() {
			_, err := createManagedCluster(managedClusterName, managedClusterAnnotations)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Cluser is imported in default mode", func() {
		It("test managedClusterAddon create/update/delete", func() {
			assertManagedClusterNamespace(managedClusterName)

			addonConfig := newKlusterletAddonConfig(managedClusterName)
			By("create klusterletAddonConfig", func() {
				err := kubeClient.Create(context.TODO(), addonConfig)
				Expect(err).ToNot(HaveOccurred())
			})

			By("check if all addons are installed", func() {
				Eventually(func() error {
					addonList := &addonv1alpha1.ManagedClusterAddOnList{}
					err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
					if err != nil {
						return err
					}
					if len(addonList.Items) != 5 {
						return fmt.Errorf("expected 5 addons, but got %v", len(addonList.Items))
					}
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			})

			By("disable addons", func() {
				Eventually(func() error {
					newAddonConfig := &agentv1.KlusterletAddonConfig{}
					err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: managedClusterName, Namespace: managedClusterName}, newAddonConfig)
					if err != nil {
						return err
					}
					newAddonConfig.Spec.SearchCollectorConfig.Enabled = false
					newAddonConfig.Spec.PolicyController.Enabled = false
					err = kubeClient.Update(context.TODO(), newAddonConfig, &client.UpdateOptions{})
					if err != nil {
						return err
					}
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			})

			By("check if the enabled addons are installed", func() {
				Eventually(func() error {
					addonList := &addonv1alpha1.ManagedClusterAddOnList{}
					err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
					if err != nil {
						return err
					}
					if len(addonList.Items) != 2 {
						return fmt.Errorf("expected 2 addons, but got %v", len(addonList.Items))
					}
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			})

			var changedAddonName string
			By("change the installNamespace of the addon", func() {
				Eventually(func() error {
					addonList := &addonv1alpha1.ManagedClusterAddOnList{}
					err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
					if err != nil {
						return err
					}
					addon := addonList.Items[0]
					addon.Spec.InstallNamespace = "default"
					err = kubeClient.Update(context.TODO(), &addon, &client.UpdateOptions{})
					if err != nil {
						return err
					}
					changedAddonName = addon.Name
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			})

			By("check if the changed addon is installed", func() {
				time.Sleep(60 * time.Second) // Wait for controller reconcile.
				Eventually(func() error {
					addon := &addonv1alpha1.ManagedClusterAddOn{}
					err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: changedAddonName, Namespace: managedClusterName}, addon)
					if err != nil {
						return err
					}
					if addon.Spec.InstallNamespace != "default" {
						return fmt.Errorf("expected the addon to be installed in default namespace, but got %s", addon.Spec.InstallNamespace)
					}
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
			})
		})
	})

	Context("Cluster is imported in hosted mode with hosted add-ons enabled", func() {
		BeforeEach(func() {
			managedClusterAnnotations = map[string]string{
				common.AnnotationKlusterletDeployMode:         "Hosted",
				common.AnnotationKlusterletHostingClusterName: "cluster1",
				common.AnnotationEnableHostedModeAddons:       "true",
			}
		})

		It("should create the managedClusterAddons in hosted mode", func() {
			assertManagedClusterNamespace(managedClusterName)
			By("check if klusterletAddonConfig is created", func() {
				var addonConfig *agentv1.KlusterletAddonConfig
				Eventually(func() error {
					config := &agentv1.KlusterletAddonConfig{}
					err := kubeClient.Get(context.TODO(), types.NamespacedName{
						Name:      managedClusterName,
						Namespace: managedClusterName,
					}, config)
					if err != nil {
						return err
					}
					addonConfig = config
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				Expect(addonConfig).ShouldNot(BeNil())
				Expect(addonConfig.Spec.PolicyController.Enabled).Should(BeTrue())
				Expect(addonConfig.Spec.ApplicationManagerConfig.Enabled).Should(BeFalse())
				Expect(addonConfig.Spec.CertPolicyControllerConfig.Enabled).Should(BeFalse())
				Expect(addonConfig.Spec.SearchCollectorConfig.Enabled).Should(BeFalse())
			})

			By("check if the desired addons are installed", func() {
				var addons []addonv1alpha1.ManagedClusterAddOn
				Eventually(func() error {
					addonList := &addonv1alpha1.ManagedClusterAddOnList{}
					err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
					if err != nil {
						return err
					}
					if len(addonList.Items) != 2 {
						return fmt.Errorf("expected 2 addons, but got %v", len(addonList.Items))
					}

					addons = addonList.Items
					return nil
				}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

				for _, addon := range addons {
					Expect(addon.Annotations[common.AnnotationAddOnHostingClusterName]).Should(Equal("cluster1"))
					Expect(addon.Spec.InstallNamespace).Should(Equal(fmt.Sprintf("klusterlet-%s", managedClusterName)))
				}
			})
		})
	})
})
