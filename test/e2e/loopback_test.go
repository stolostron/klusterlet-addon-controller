// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e_test

import (
	"context"
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addoncontroller "github.com/open-cluster-management/klusterlet-addon-controller/pkg/controller/klusterletaddon"
)

var _ = Describe("Loopback test", func() {
	It("create and delete all addons", func() {
		testNamespace := "cluster1-test"
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
				ApplicationManagerConfig: agentv1.KlusterletAddonConfigApplicationManagerSpec{
					Enabled: true,
				},
				CertPolicyControllerConfig: agentv1.KlusterletAddonConfigCertPolicyControllerSpec{
					Enabled: true,
				},
				IAMPolicyControllerConfig: agentv1.KlusterletAddonConfigIAMPolicyControllerSpec{
					Enabled: true,
				},
				PolicyController: agentv1.KlusterletAddonConfigPolicyControllerSpec{
					Enabled: true,
				},
				SearchCollectorConfig: agentv1.KlusterletAddonConfigSearchCollectorSpec{
					Enabled: true,
				},

				ClusterName:      managedclusterName,
				ClusterNamespace: testNamespace,
				ClusterLabels: map[string]string{
					"author": "tester",
				},
				Version: "2.0.0",
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

		By("Check agent addon deployments, will not check the pod Status as the image pull secret do not provide and Github action resource limitation")
		deployments := []string{
			"klusterlet-addon-operator",
			"klusterlet-addon-appmgr",
			"klusterlet-addon-certpolicyctrl",
			"klusterlet-addon-iampolicyctrl",
			"klusterlet-addon-policyctrl-config-policy",
			"klusterlet-addon-policyctrl-framework",
			"klusterlet-addon-search",
			"klusterlet-addon-workmgr",
		}

		for _, deploy := range deployments {
			Eventually(func() bool {
				deployment, err := spokeClient.AppsV1().Deployments(agentAddonNamespace).Get(context.TODO(), deploy, metav1.GetOptions{})
				if err != nil {
					logf.Log.Info("Get deployment error", "namespace", agentAddonNamespace, "name", deploy, "error", err)
					return false
				}

				logf.Log.Info("Deployment created", "name", deployment.Name)
				return true
			}, 300*time.Second, 3*time.Second).Should(BeTrue())
		}

		By("Delete klusterletaddonconfigs")
		err = kubeClient.Delete(context.TODO(), &klusterletAddonConfig)
		Expect(err).ToNot(HaveOccurred())

		By("Check manifestworks are deleted")
		manifestWorks := []string{
			"cluster1-klusterlet-addon-appmgr",
			"cluster1-klusterlet-addon-certpolicyctrl",
			"cluster1-klusterlet-addon-iampolicyctrl",
			"cluster1-klusterlet-addon-policyctrl",
			"cluster1-klusterlet-addon-search",
			"cluster1-klusterlet-addon-workmgr",
			"cluster1-klusterlet-addon-operator",
			"cluster1-klusterlet-addon-crds",
		}
		for _, mw := range manifestWorks {
			Eventually(func() bool {
				manifestwork := manifestworkv1.ManifestWork{}
				err = kubeClient.Get(context.TODO(), client.ObjectKey{
					Namespace: managedclusterName,
					Name:      mw,
				}, &manifestwork)
				if err == nil {
					return false
				}

				if errors.IsNotFound(err) {
					logf.Log.Info("Manifestwork deleted", "name", mw)
					return true
				}

				logf.Log.Info("Get manifestwork error", "name", mw, "error", err)
				return false
			}, 300*time.Second, 3*time.Second).Should(BeTrue())
		}

		By("Check namespaces are deleted")
		namespaces := []string{
			testNamespace,
			agentAddonNamespace,
		}
		for _, ns := range namespaces {
			Eventually(func() bool {
				_, err := spokeClient.CoreV1().Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
				if err == nil {
					return false
				}

				if errors.IsNotFound(err) {
					logf.Log.Info("Namespace deleted", "name", ns)
					return true
				}

				logf.Log.Info("Get namespace error", "name", ns, "error", err)
				return false
			}, 300*time.Second, 3*time.Second).Should(BeTrue())
		}

		By("Check klusterletaddonconfig is deleted")
		Eventually(func() bool {
			klusterletAddonConfig := agentv1.KlusterletAddonConfig{}
			err = kubeClient.Get(context.TODO(), client.ObjectKey{
				Namespace: testKlusterletAddonConfig.Namespace,
				Name:      testKlusterletAddonConfig.Name,
			}, &klusterletAddonConfig)
			if err == nil {
				return false
			}

			if errors.IsNotFound(err) {
				logf.Log.Info("KlusterletAddonConfig deleted", "name", klusterletAddonConfig.Name)
				return true
			}

			logf.Log.Info("Get klusterletAddonConfig error", "name", testKlusterletAddonConfig.Name, "error", err)
			return false
		}, 300*time.Second, 3*time.Second).Should(BeTrue())

		By("Check klusterletaddonconfig cleanup finalizer on managed cluster is removed")
		Eventually(func() bool {
			managedCluster := managedclusterv1.ManagedCluster{}
			err = kubeClient.Get(context.TODO(), client.ObjectKey{
				Name: managedclusterName,
			}, &managedCluster)
			if err != nil {
				logf.Log.Info("Get managedCluster error", "name", managedclusterName, "error", err)
				return false
			}

			for _, finalizer := range managedCluster.Finalizers {
				if finalizer == addoncontroller.KlusterletAddonFinalizer {
					logf.Log.Info("Klusterlet addon finalizer still exist", "name", managedclusterName, "finalizer", finalizer)
					return false
				}
			}
			return true
		}, 300*time.Second, 3*time.Second).Should(BeTrue())
	})
})
