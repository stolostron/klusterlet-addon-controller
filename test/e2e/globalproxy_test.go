// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/helpers"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
)

var _ = Describe("globalProxy test", func() {
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

	It("test managedClusterAddon with OCPGlobalProxy ", func() {
		assertManagedClusterNamespace(managedClusterName)

		addonConfig := newKlusterletAddonConfig(managedClusterName)
		By("create klusterletAddonConfig with App addon enable OCPGlobalProxy", func() {
			addonConfig.Spec.ApplicationManagerConfig.ProxyPolicy = agentv1.ProxyPolicyOCPGlobalProxy
			err := kubeClient.Create(context.TODO(), addonConfig)
			Expect(err).ToNot(HaveOccurred())
		})

		By("create install-config secret", func() {
			secret := helpers.NewInstallConfigSecret(fmt.Sprintf("%s-install-config", managedClusterName), managedClusterName, helpers.InstallConfigYaml)
			_, err := hubClient.CoreV1().Secrets(managedClusterName).Create(context.TODO(), secret, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
		})
		By("check if klusterletAddonConfig status has OCPGlobalProxy", func() {
			addonConfig := &agentv1.KlusterletAddonConfig{}
			Eventually(func() error {
				err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: managedClusterName, Namespace: managedClusterName}, addonConfig)
				if err != nil {
					return err
				}
				if addonConfig.Status.OCPGlobalProxy.HTTPProxy == "" ||
					addonConfig.Status.OCPGlobalProxy.NoProxy == "" ||
					addonConfig.Status.OCPGlobalProxy.HTTPSProxy == "" {
					return fmt.Errorf("expected addonConfig has OCPGlobalProxy status,but got %v", addonConfig.Status.OCPGlobalProxy)
				}
				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
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

		By("check if app addon has annotation", func() {
			Eventually(func() error {
				appAddon := &addonv1alpha1.ManagedClusterAddOn{}
				err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: agentv1.ApplicationAddonName, Namespace: managedClusterName}, appAddon)
				if err != nil {
					return err
				}
				if len(appAddon.Annotations) == 0 {
					return fmt.Errorf("expected app addon has annation, but got empty")
				}
				if _, ok := appAddon.Annotations["addon.open-cluster-management.io/values"]; !ok {
					return fmt.Errorf("expected app addon has value annation, but got empty")
				}

				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})
	})
	It("test managedClusterAddon with CustomGlobalProxy ", func() {
		assertManagedClusterNamespace(managedClusterName)

		addonConfig := newKlusterletAddonConfig(managedClusterName)
		By("create klusterletAddonConfig with App addon enable CustomProxy", func() {
			addonConfig.Spec.ProxyConfig = agentv1.ProxyConfig{
				HTTPProxy:  "https://username:password@proxy.example.com:123/",
				HTTPSProxy: "https://username:password@proxy.example.com:123/",
				NoProxy:    ".cluster.local,.svc,10.128.0.0/14,123.example.com,10.88.0.0/16,127.0.0.1,169.254.169.254,172.30.0.0/16,192.168.124.0/24,api-int.cluster.test.redhat.com,localhost",
			}
			addonConfig.Spec.ApplicationManagerConfig.ProxyPolicy = agentv1.ProxyPolicyCustomProxy
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

		By("check if app addon has annotation", func() {
			Eventually(func() error {
				appAddon := &addonv1alpha1.ManagedClusterAddOn{}
				err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: agentv1.ApplicationAddonName, Namespace: managedClusterName}, appAddon)
				if err != nil {
					return err
				}
				if len(appAddon.Annotations) == 0 {
					return fmt.Errorf("expected app addon has annation, but got empty")
				}
				if _, ok := appAddon.Annotations["addon.open-cluster-management.io/values"]; !ok {
					return fmt.Errorf("expected app addon has value annation, but got empty")
				}

				return nil
			}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		})
	})
})
