// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
)

var _ = Describe("Loopback test", func() {
	gvr := schema.GroupVersionResource{
		Group:    "agent.open-cluster-management.io",
		Version:  "v1",
		Resource: "klusterletaddonconfigs",
	}

	It("enable all addons", func() {
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
				ClusterNamespace: "cluster1-test",
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
		createdKac, err := hubDynamicClient.Resource(gvr).Namespace(managedclusterName).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(BeFalse()).To(BeTrue())
		}
		logf.Log.Info("Klusterlet addon config created", "klusterletAddonConfig", createdKac)

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

				if errors.IsNotFound(err) {
					return false
				}

				By(fmt.Sprintf("check deployment for %s success", deployment.Name))
				return true
			}, 300*time.Second, 3*time.Second).Should(BeTrue())
		}

		By("Delete klusterletaddonconfigs")
		err = hubDynamicClient.Resource(gvr).Namespace(managedclusterName).Delete(context.TODO(), managedclusterName, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())
	})
})
