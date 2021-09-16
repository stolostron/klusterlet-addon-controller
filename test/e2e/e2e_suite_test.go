// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	hubClient           kubernetes.Interface
	spokeClient         kubernetes.Interface
	hubDynamicClient    dynamic.Interface
	agentAddonNamespace string
	managedclusterName  string
	clusterCfg          *rest.Config
)

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	managedclusterName = "cluster1"
	agentAddonNamespace = "open-cluster-management-agent-addon"
	kubeconfig := os.Getenv("KUBECONFIG")
	err := func() error {
		var err error
		clusterCfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return err
		}

		hubClient, err = kubernetes.NewForConfig(clusterCfg)
		if err != nil {
			return err
		}

		spokeClient = hubClient

		hubDynamicClient, err = dynamic.NewForConfig(clusterCfg)
		if err != nil {
			return err
		}

		return err
	}()
	Expect(err).ToNot(HaveOccurred())
})
