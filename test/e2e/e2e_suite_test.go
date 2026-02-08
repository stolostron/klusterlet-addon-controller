// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:revive // dot imports are idiomatic for ginkgo
	. "github.com/onsi/gomega"    //nolint:revive // dot imports are idiomatic for gomega
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var (
	hubClient        kubernetes.Interface
	hubClusterClient clusterclient.Interface
	kubeClient       client.Client
	clusterCfg       *rest.Config
)

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
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

		kubeClient, err = newRuntimeClient(clusterCfg)
		if err != nil {
			return err
		}
		if kubeClient == nil {
			return fmt.Errorf("go nil kubeClient")
		}

		hubClusterClient, err = clusterclient.NewForConfig(clusterCfg)
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
	if err := clusterv1.AddToScheme(kscheme.Scheme); err != nil {
		logf.Log.Info("add to scheme error", "error", err, "name", "managedcluster")
		return nil, err
	}

	if err := addonv1alpha1.AddToScheme(kscheme.Scheme); err != nil {
		logf.Log.Info("add to scheme error", "error", err, "name", "managedClusterAddon")
		return nil, err
	}

	c, err := client.New(config, client.Options{Scheme: kscheme.Scheme})
	if err != nil {
		logf.Log.Info("Failed to initialize a client connection to the cluster", "error", err.Error())
		return nil, err
	}
	return c, nil
}

func createManagedCluster(clusterName string, annotations map[string]string) (*clusterv1.ManagedCluster, error) {
	cluster, err := hubClusterClient.ClusterV1().ManagedClusters().Get(context.TODO(), clusterName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return hubClusterClient.ClusterV1().ManagedClusters().Create(
			context.TODO(),
			&clusterv1.ManagedCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:        clusterName,
					Annotations: annotations,
				},
				Spec: clusterv1.ManagedClusterSpec{
					HubAcceptsClient: true,
				},
			},
			metav1.CreateOptions{},
		)
	}

	return cluster, err
}

func assertManagedClusterNamespace(managedClusterName string) {
	By("Should create the managedCluster namespace", func() {
		Expect(wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
			_, err := hubClient.CoreV1().Namespaces().Get(ctx, managedClusterName, metav1.GetOptions{})
			if err != nil && errors.IsNotFound(err) {
				return true, nil
			}
			if err != nil {
				return false, err
			}
			return true, nil
		})).ToNot(HaveOccurred())
	})
}

func deleteManagedCluster(clusterName string) {
	err := hubClusterClient.ClusterV1().ManagedClusters().Delete(context.TODO(), clusterName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		Expect(err).ToNot(HaveOccurred())
	}

	err = hubClient.CoreV1().Namespaces().Delete(context.TODO(), clusterName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		Expect(err).ToNot(HaveOccurred())
	}
}

func newKlusterletAddonConfig(managedClusterName string) *agentv1.KlusterletAddonConfig {
	return &agentv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      managedClusterName,
			Namespace: managedClusterName,
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonAgentConfigSpec{
				Enabled: true,
			},
			CertPolicyControllerConfig: agentv1.KlusterletAddonAgentConfigSpec{
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
}
