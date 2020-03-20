// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"path/filepath"

	"github.com/ghodss/yaml"
	certmanagerv1alpha1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha1"
	"github.com/open-cluster-management/endpoint-operator/pkg/apis"
	"github.com/open-cluster-management/endpoint-operator/pkg/controller"
	"github.com/open-cluster-management/endpoint-operator/pkg/inspect"
	"github.com/open-cluster-management/endpoint-operator/version"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Add the zap logger flag set to the CLI. The flag set must
	// be added before calling pflag.Parse().
	pflag.CommandLine.AddFlagSet(zap.FlagSet())

	// Add flags registered by imported packages (e.g. glog and
	// controller-runtime)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	pflag.Parse()

	// Use a zap logr.Logger implementation. If none of the zap
	// flags are configured (or if the zap flag set is not being
	// used), this defaults to a production zap logger.
	//
	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Error(err, "Failed to get watch namespace")
		os.Exit(1)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	ctx := context.TODO()
	// Become the leader before proceeding
	err = leader.Become(ctx, "endpoint-operator-lock")
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
	// Get cluster info
	if err := inspect.InitClusterInfo(cfg); err != nil {
		log.Error(err, "Failed to get cluster info. Skipping.")
	}

	log.Info("Installing CRDs")
	if err := installCRDs(cfg); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := certmanagerv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Add the Metrics Service
	addMetrics(ctx, cfg, namespace)

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

// addMetrics will create the Services and Service Monitors to allow the operator export the metrics by using
// the Prometheus operator
func addMetrics(ctx context.Context, cfg *rest.Config, namespace string) {
	if err := serveCRMetrics(cfg); err != nil {
		if errors.Is(err, k8sutil.ErrRunLocal) {
			log.Info("Skipping CR metrics server creation; not running in a cluster.")
			return
		}
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}

	// Add to the below struct any other metrics ports you want to expose.
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
		{Port: operatorMetricsPort, Name: metrics.CRPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: operatorMetricsPort}},
	}

	// Create Service object to expose the metrics port(s).
	service, err := metrics.CreateMetricsService(ctx, cfg, servicePorts)
	if err != nil {
		log.Info("Could not create metrics Service", "error", err.Error())
	}

	// CreateServiceMonitors will automatically create the prometheus-operator ServiceMonitor resources
	// necessary to configure Prometheus to scrape metrics from this operator.
	services := []*v1.Service{service}
	_, err = metrics.CreateServiceMonitors(cfg, namespace, services)
	if err != nil {
		log.Info("Could not create ServiceMonitor object", "error", err.Error())
		// If this operator is deployed to a cluster without the prometheus-operator running, it will return
		// ErrServiceMonitorNotPresent, which can be used to safely skip ServiceMonitor creation.
		if err == metrics.ErrServiceMonitorNotPresent {
			log.Info("Install prometheus-operator in your cluster to create ServiceMonitor objects", "error", err.Error())
		}
	}
}

// serveCRMetrics gets the Operator/CustomResource GVKs and generates metrics based on those types.
// It serves those metrics on "http://metricsHost:operatorMetricsPort".
func serveCRMetrics(cfg *rest.Config) error {
	// Below function returns filtered operator/CustomResource specific GVKs.
	// For more control override the below GVK list with your own custom logic.
	filteredGVK, err := k8sutil.GetGVKsFromAddToScheme(apis.AddToScheme)
	if err != nil {
		return err
	}
	// Get the namespace the operator is currently deployed in.
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return err
	}
	// To generate metrics in other namespaces, add the values below.
	ns := []string{operatorNs}
	// Generate and serve custom resource specific metrics.
	err = kubemetrics.GenerateAndServeCRMetrics(cfg, ns, filteredGVK, metricsHost, operatorMetricsPort)
	if err != nil {
		return err
	}
	return nil
}

func installCRDs(cfg *rest.Config) error {
	crdClient := crdclientset.NewForConfigOrDie(cfg)

	crdsPath := "deploy/crds"
	files, err := ioutil.ReadDir(crdsPath)
	if err != nil {
		log.Error(err, "Fail to read CRDs directory", "path", crdsPath)
		return err
	}
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), "crd.yaml") {
			crdFilePath := filepath.Join(crdsPath, file.Name())
			log.V(1).Info("Found CRD Yaml", "file", crdFilePath)
			crdYaml, err := ioutil.ReadFile(filepath.Join(crdsPath, file.Name()))
			if err != nil {
				log.Error(err, "Fail to read file", "path", crdFilePath)
				return err
			}
			crd := &crdv1beta1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(crdYaml, crd); err != nil {
				log.Error(err, "Fail to unmarshal crd yaml", "content", crdYaml)
				return err
			}
			if err := createOrUpdateCRD(crd, crdClient); err != nil {
				log.Error(err, "Failed to create/update crd", "path", crdFilePath)
				return err
			}
		}
	}

	// TODO operator cannot run without cert manager crds existing
	// becuse they are not apart of the operator they should actually be
	// laid down as a pre-req
	certPath := "deploy/certmanager"
	certFiles, err := ioutil.ReadDir(certPath)
	if err != nil {
		log.Error(err, "Failed to cread cert manager CRDs directory", "path", certPath)
		return err
	}
	for _, file := range certFiles {
		if !file.IsDir() && strings.Contains(file.Name(), "crd.yaml") {
			certFilePath := filepath.Join(certPath, file.Name())
			log.V(1).Info("Found CRD Yaml", "file", certFilePath)
			crdYaml, err := ioutil.ReadFile(filepath.Join(certPath, file.Name()))
			if err != nil {
				log.Error(err, "Fail to read file", "path", certFilePath)
				return err
			}
			crd := &crdv1beta1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(crdYaml, crd); err != nil {
				log.Error(err, "Fail to unmarshal crd yaml", "content", crdYaml)
				return err
			}
			if err := createOrUpdateCRD(crd, crdClient); err != nil {
				log.Error(err, "Failed to create/update crd", "path", certFilePath)
				return err
			}
		}
	}

	return nil
}

func createOrUpdateCRD(crd *crdv1beta1.CustomResourceDefinition, crdClient *crdclientset.Clientset) error {
	log.Info("Create or update component CRD", "name", crd.Name)

	log.V(1).Info("Looking for CRD", "name", crd.Name)
	foundCRD, err := crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Error(err, "Unexpected error get CRD", "name", crd.Name)
			return err
		}

		log.V(1).Info("Creating CRD", "name", crd.Name)
		if _, err := crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
			log.Error(err, "Fail to create CRD", "name", crd.Name)
			return err
		}
		return nil
	}

	// NOTE: the UPDATE will always run since API server add additional stuff to the Spec but that's ok
	// 	However this does present a problem for when rolling back the version of klusterlet operator...
	//  If the newer version have a newer API than if we rollback to older version and it try to call Update
	//  the Update will fail
	log.V(1).Info("Updating CRD", "name", crd.Name)
	foundCRD.Spec = crd.Spec
	if _, err = crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().Update(foundCRD); err != nil {
		log.Error(err, "Fail to update CRD", "name", foundCRD.Name)
		return err
	}

	return nil
}
