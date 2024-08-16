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

var _ = Describe("klusterletAddonConfig test", func() {
	Context("klusterletAddonConfig test for hypershift cluster", func() {
		var managedClusterName string
		BeforeEach(func() {
			managedClusterName = fmt.Sprintf("cluster-test-%s", rand.String(6))

			By(fmt.Sprintf("Create managed cluster %s", managedClusterName), func() {
				_, err := createManagedCluster(managedClusterName, map[string]string{"cluster.open-cluster-management.io/provisioner": "test.test.HypershiftDeployment.cluster.open-cluster-management.io"})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		AfterEach(func() {
			deleteManagedCluster(managedClusterName)
		})

		It("test klusterletAddonConfig create", func() {
			testKACCreate(managedClusterName, 3)
		})
	})

	Context("klusterletAddonConfig test for claim cluster", func() {
		var managedClusterName string
		BeforeEach(func() {
			managedClusterName = fmt.Sprintf("cluster-test-%s", rand.String(6))

			By(fmt.Sprintf("Create managed cluster %s", managedClusterName), func() {
				_, err := createManagedCluster(managedClusterName,
					map[string]string{"cluster.open-cluster-management.io/provisioner": fmt.Sprintf("%s.%s.ClusterClaim.hive.openshift.io/v1", managedClusterName, managedClusterName)})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		AfterEach(func() {
			deleteManagedCluster(managedClusterName)
		})

		It("test klusterletAddonConfig create", func() {
			testKACCreate(managedClusterName, 5)
		})
	})

	Context("klusterletAddonConfig test for normal clusters with annotation", func() {
		var managedClusterName string
		BeforeEach(func() {
			managedClusterName = fmt.Sprintf("cluster-test-%s", rand.String(6))

			By(fmt.Sprintf("Create managed cluster %s", managedClusterName), func() {
				_, err := createManagedCluster(managedClusterName,
					map[string]string{common.AnnotationCreateWithDefaultKlusterletAddonConfig: "true"})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		AfterEach(func() {
			deleteManagedCluster(managedClusterName)
		})

		It("test klusterletAddonConfig create", func() {
			testKACCreate(managedClusterName, 4)
		})
	})
})

func testKACCreate(managedClusterName string, addonNumber int) {
	addonConfig := &agentv1.KlusterletAddonConfig{}
	By("check if klusterletAddonConfig is created", func() {
		Eventually(func() error {
			return kubeClient.Get(context.TODO(), types.NamespacedName{Name: managedClusterName, Namespace: managedClusterName}, addonConfig)
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
	})

	By("check if all addons are installed", func() {
		Eventually(func() error {
			addonList := &addonv1alpha1.ManagedClusterAddOnList{}
			err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
			if err != nil {
				return err
			}
			if len(addonList.Items) != addonNumber {
				return fmt.Errorf("expected %d addons, but got %v", addonNumber, len(addonList.Items))
			}
			return nil
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
	})

	By("delete klusterletAddonConfig", func() {
		err := kubeClient.Delete(context.TODO(), addonConfig)
		Expect(err).ToNot(HaveOccurred())
	})

	By("check if klusterletAddonConfig is created again", func() {
		Eventually(func() error {
			return kubeClient.Get(context.TODO(), types.NamespacedName{Name: managedClusterName, Namespace: managedClusterName}, addonConfig)
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
	})

	By("check if all addons are installed", func() {
		Eventually(func() error {
			addonList := &addonv1alpha1.ManagedClusterAddOnList{}
			err := kubeClient.List(context.TODO(), addonList, &client.ListOptions{Namespace: managedClusterName})
			if err != nil {
				return err
			}
			if len(addonList.Items) != addonNumber {
				return fmt.Errorf("expected %d addons, but got %v", addonNumber, len(addonList.Items))
			}
			return nil
		}, 60*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
	})
}
