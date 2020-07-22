// +build functional

// Copyright (c) 2020 Red Hat, Inc.

package klusterlet_addon_controller_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

const (
	invalidSemanticVersion    = "2.0.0.12"
	unavailableVersion        = "1.0.0"
	klusterletaddonconfigName = "webhook-test"
	versionList               = "2.0.0 2.1.0"
	admissionName             = "klusterletaddonconfig.validating-webhook.open-cluster-management.io"
)

// list of regex we need to check
var _ = Describe("KlusterletAddonConfig admission webhook", func() {
	BeforeEach(func() {
		Eventually(func() error {
			_, err := clientCluster.CoreV1().Services(klusterletAddonNamespace).Get(context.TODO(), webhookserviceName, metav1.GetOptions{})
			return err
		}, 20*time.Second, 1*time.Second).Should(BeNil())

		Eventually(func() error {
			_, err := clientCluster.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(context.TODO(), validatingCfgName, metav1.GetOptions{})

			return err
		}, 20*time.Second, 1*time.Second).Should(BeNil())
	})

	Context("Creating a klusterletaddonconfig", func() {
		It("Should respond bad request when creating a klusterletaddonconfig with invalid semantic version", func() {
			By("Creating KlusterletAddonConfig with invalid semantic version", func() {
				klusterletAddonConfig := newKlusterletAddonConfig(klusterletaddonconfigName, testWebhookNamespace, invalidSemanticVersion)

				_, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Create(context.TODO(), klusterletAddonConfig, metav1.CreateOptions{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(fmt.Sprintf(
					"admission webhook \"%s\" denied the request: Version \"%s\" is invalid semantic version",
					admissionName,
					invalidSemanticVersion,
				)))
			})
		})

		It("Should respond bad request when creating a klusterletaddonconfig with unavailable version", func() {
			By("Creating KlusterletAddonConfig with unavailable version", func() {
				klusterletAddonConfig := newKlusterletAddonConfig(klusterletaddonconfigName, testWebhookNamespace, unavailableVersion)

				_, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Create(context.TODO(), klusterletAddonConfig, metav1.CreateOptions{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(fmt.Sprintf(
					"admission webhook \"%s\" denied the request: Version %s is not available. Available Versions are: [%s]",
					admissionName,
					unavailableVersion,
					versionList,
				)))
			})
		})

	})

	Context("Updating a klusterletaddonconfig", func() {
		BeforeEach(func() {
			By("Cleanup old test data", func() {
				cleanUpWebhookTestData(clientClusterDynamic)
			})
			By("Creating KlusterletAddonConfig", func() {
				klusterletAddonConfig := newKlusterletAddonConfig(klusterletaddonconfigName, testWebhookNamespace, validVersion)

				createNewUnstructured(clientClusterDynamic, gvrKlusterletAddonConfig,
					klusterletAddonConfig, klusterletaddonconfigName, testWebhookNamespace)
			})
		})

		It("Should respond bad request when creating a klusterletaddonconfig with invalid semantic version", func() {
			By("Updating klusterletaddonconfig with invalid semantic version", func() {
				klusterletAddonConfig, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Get(context.TODO(), klusterletaddonconfigName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				klusterletAddonConfig.Object["spec"].(map[string]interface{})["version"] = invalidSemanticVersion
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Update(context.TODO(), klusterletAddonConfig, metav1.UpdateOptions{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(fmt.Sprintf(
					"admission webhook \"%s\" denied the request: Version \"%s\" is invalid semantic version",
					admissionName,
					invalidSemanticVersion,
				)))
			})
		})

		It("Should respond bad request when creating a klusterletaddonconfig with unavailable version", func() {
			By("Updating klusterletaddonconfig with unavailable version", func() {
				klusterletAddonConfig, err := clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Get(context.TODO(), klusterletaddonconfigName, metav1.GetOptions{})
				Expect(err).NotTo(HaveOccurred())

				klusterletAddonConfig.Object["spec"].(map[string]interface{})["version"] = unavailableVersion
				_, err = clientClusterDynamic.Resource(gvrKlusterletAddonConfig).Namespace(testWebhookNamespace).Update(context.TODO(), klusterletAddonConfig, metav1.UpdateOptions{})
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).Should(Equal(fmt.Sprintf(
					"admission webhook \"%s\" denied the request: Version %s is not available. Available Versions are: [%s]",
					admissionName,
					unavailableVersion,
					versionList,
				)))
			})
		})
	})

})

func cleanUpWebhookTestData(clientClusterDynamic dynamic.Interface) {
	klog.V(1).Info("Deleting KlusterletAddonConfig")
	deleteIfExists(clientClusterDynamic, gvrKlusterletAddonConfig, klusterletaddonconfigName, testWebhookNamespace)
}
