// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

var _ = Describe("Loopback test", func() {
	It("create and delete all addons", func() {
		testKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: agentv1.SchemeGroupVersion.String(),
				Kind:       "KlusterletAddonConfig",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      managedclusterName,
				Namespace: managedclusterName,
			},
			Spec: agentv1.KlusterletAddonConfigSpec{
				ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
					Enabled: true,
				},
				CertPolicyControllerConfig: agentv1.KlusterletAddonAgentConfigSpec{
					Enabled: true,
				},
				IAMPolicyControllerConfig: agentv1.KlusterletAddonAgentConfigSpec{
					Enabled: true,
				},
				PolicyController: agentv1.KlusterletAddonAgentConfigSpec{
					Enabled: true,
				},
				SearchCollectorConfig: agentv1.KlusterletAddonAgentConfigSpec{
					Enabled: true,
				},
			},
		}

		raw, err := json.Marshal(testKlusterletAddonConfig)
		Expect(err).ToNot(HaveOccurred())
		obj := &unstructured.Unstructured{}
		err = json.Unmarshal(raw, obj)
		Expect(err).ToNot(HaveOccurred())

		By("Create klusterletaddonconfigs")
		err = kubeClient.Create(context.TODO(), testKlusterletAddonConfig)
		if err != nil && !errors.IsAlreadyExists(err) {
			logf.Log.Info("Create klusterletAddonConfig error", "klusterletAddonConfig", testKlusterletAddonConfig.Name, "error", err)
			Expect(BeFalse()).To(BeTrue())
		}
		klusterletAddonConfig := agentv1.KlusterletAddonConfig{}
		Eventually(func() bool {
			err = kubeClient.Get(context.TODO(), client.ObjectKey{
				Namespace: testKlusterletAddonConfig.Namespace,
				Name:      testKlusterletAddonConfig.Name,
			}, &klusterletAddonConfig)
			if err != nil {
				logf.Log.Info("Get klusterletAddonConfig error", "klusterletAddonConfig", testKlusterletAddonConfig.Name, "error", err)
				return false
			}

			logf.Log.Info("Klusterlet addon config created", "klusterletAddonConfig", klusterletAddonConfig)
			return true
		}, 3*time.Second, 500*time.Millisecond).Should(BeTrue())

		By("Check addons are installed")
		for addonName := range agentv1.KlusterletAddons {
			if addonName == agentv1.WorkManagerAddonName || addonName == agentv1.PolicyAddonName {
				continue
			}
			Eventually(func() error {
				addon := &addonv1alpha1.ManagedClusterAddOn{}
				return kubeClient.Get(context.TODO(), types.NamespacedName{Name: addonName, Namespace: testKlusterletAddonConfig.Namespace}, addon)
			}, 300*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		}

		By("Delete klusterletaddonconfigs")
		err = kubeClient.Delete(context.TODO(), &klusterletAddonConfig)
		Expect(err).ToNot(HaveOccurred())

		By("Check addons are deleted")
		for addonName := range agentv1.KlusterletAddons {
			if addonName == agentv1.WorkManagerAddonName || addonName == agentv1.PolicyAddonName {
				continue
			}
			Eventually(func() error {
				addon := &addonv1alpha1.ManagedClusterAddOn{}
				err := kubeClient.Get(context.TODO(), types.NamespacedName{Name: addonName, Namespace: testKlusterletAddonConfig.Namespace}, addon)
				if errors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("failed to get addon %v,%v", addonName, err)
			}, 300*time.Second, 5*time.Second).ShouldNot(HaveOccurred())
		}

	})
})
