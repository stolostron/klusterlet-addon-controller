// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("managedClusterAddon test", func() {
	var managedClusterName string
	BeforeEach(func() {
		managedClusterName = fmt.Sprintf("cluster-test-%s", rand.String(6))

		By(fmt.Sprintf("Create managed cluster %s", managedClusterName), func() {
			_, err := createManagedCluster(managedClusterName, map[string]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		deleteManagedCluster(managedClusterName)
	})

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
				if len(addonList.Items) != 6 {
					return fmt.Errorf("expected 6 addons, but got %v", len(addonList.Items))
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
				if len(addonList.Items) != 3 {
					return fmt.Errorf("expected 3 addons, but got %v", len(addonList.Items))
				}
				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())

		})

		By("delete klusterletAddonConfig", func() {
			err := kubeClient.Delete(context.TODO(), addonConfig)
			Expect(err).ToNot(HaveOccurred())
		})

		By("check if all addons are deleted", func() {
			Eventually(func() error {
				addonList := &addonv1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
				if err != nil {
					return err
				}
				if len(addonList.Items) != 0 {
					return fmt.Errorf("expected 0 addons, but got %v", len(addonList.Items))
				}
				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

	})
	It("test managedCluster delete", func() {
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
				if len(addonList.Items) != 6 {
					return fmt.Errorf("expected 6 addons, but got %v", len(addonList.Items))
				}
				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})

		By("delete managedCluster", func() {
			err := hubClusterClient.ClusterV1().ManagedClusters().Delete(context.TODO(), managedClusterName, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())
		})

		By("check if all addons are deleted", func() {
			Eventually(func() error {
				addonList := &addonv1alpha1.ManagedClusterAddOnList{}
				err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
				if err != nil {
					return err
				}
				if len(addonList.Items) != 0 {
					return fmt.Errorf("expected 0 addons, but got %v", len(addonList.Items))
				}
				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})
	})
})
