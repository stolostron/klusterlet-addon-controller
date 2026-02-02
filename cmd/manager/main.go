// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	ocinfrav1 "github.com/openshift/api/config/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller"
	"github.com/stolostron/klusterlet-addon-controller/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"

	// "github.com/operator-framework/operator-sdk/pkg/k8sutil"
	// kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Change below variables to serve metrics on different host or port.
var (
	metricsHost       = "0.0.0.0"
	metricsPort int32 = 8383
)
var (
	setupLog = logf.Log.WithName("setup")
)
var log = logf.Log.WithName("cmd")

func printVersion() {
	log.Info(fmt.Sprintf("Operator Version: %s", version.Version))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	var metricsAddr string

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.Parse()

	ctrl.SetLogger(zap.New())

	// if the controller is deployed by MCH, the env HUB_VERSION is required to set the MCH version.
	// otherwise the env HUB_VERSION is not required.
	version.Version = os.Getenv("HUB_VERSION")
	if version.Version == "" {
		log.Info("Cannot get operator version from env var, use 0.0.0")
		version.Version = "0.0.0"
	}
	printVersion()

	log.Info("Initializing klusterlet-addon-controller manager")

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to get config")
		os.Exit(1)
	}

	runtimeClient, err := newRuntimeClient(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	err = agentv1.LoadImages(runtimeClient)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Metrics: metricsserver.Options{
			BindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		},
		LeaderElection:   true,
		LeaderElectionID: "klusterlet-addon-controller-lock",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := managedclusterv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := manifestworkv1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := ocinfrav1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := addonv1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr, kubeClient, dynamicClient); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")
	log.Info("All controllers have been set up successfully, starting manager loop")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func newRuntimeClient(conf *rest.Config) (client.Client, error) {
	kubeClient, err := client.New(conf, client.Options{})
	if err != nil {
		log.Info("Failed to initialize a client connection to the cluster", "error", err.Error())
		return nil, fmt.Errorf("failed to initialize a client connection to the cluster")
	}
	return kubeClient, nil
}
