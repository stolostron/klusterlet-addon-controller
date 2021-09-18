// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	hubClient           kubernetes.Interface
	spokeClient         kubernetes.Interface
	kubeClient          client.Client
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

		kubeClient, err = newRuntimeClient(clusterCfg)
		if err != nil {
			return err
		}
		return nil
	}()
	Expect(err).ToNot(HaveOccurred())
})

func newRuntimeClient(config *rest.Config) (client.Client, error) {
	if err := agentv1.SchemeBuilder.AddToScheme(kscheme.Scheme); err != nil {
		logf.Log.Info("add to scheme error", "error", err, "name", "klusterletaddon")
		return nil, err
	}
	if err := manifestworkv1.AddToScheme(kscheme.Scheme); err != nil {
		logf.Log.Info("add to scheme error", "error", err, "name", "manifestwork")
		return nil, err
	}
	if err := managedclusterv1.AddToScheme(kscheme.Scheme); err != nil {
		logf.Log.Info("add to scheme error", "error", err, "name", "managedcluster")
		return nil, err
	}

	c, err := client.New(clusterCfg, client.Options{
		Scheme: kscheme.Scheme,
	})
	if err != nil {
		logf.Log.Info("Failed to initialize a client connection to the cluster", "error", err.Error())
		return nil, err
	}
	return c, nil
}
