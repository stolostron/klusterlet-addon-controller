// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/apis"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/controller"
	"github.com/open-cluster-management/endpoint-operator/pkg/controller/clustermanagementaddon"
	"github.com/open-cluster-management/endpoint-operator/version"
	ocinfrav1 "github.com/openshift/api/config/v1"

	//"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	//kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
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

	printVersion()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to get config")
		os.Exit(1)
	}

	kubeclient, err := newK8s(cfg)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	err = agentv1.LoadConfigmaps(kubeclient)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:               os.Getenv("WATCH_NAMESPACE"),
		MetricsBindAddress:      fmt.Sprintf("%s:%d", metricsHost, metricsPort),
		LeaderElection:          true,
		LeaderElectionID:        "klusterlet-addon-controller-lock",
		LeaderElectionNamespace: "open-cluster-management",
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

	// create all ClusterManagementAddons for monolith addons
	clustermanagementaddon.CreateClusterManagementAddon(kubeclient)

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}
}

func newK8s(conf *rest.Config) (client.Client, error) {
	kubeClient, err := client.New(conf, client.Options{})
	if err != nil {
		log.Info("Failed to initialize a client connection to the cluster", "error", err.Error())
		return nil, fmt.Errorf("Failed to initialize a client connection to the cluster")
	}
	return kubeClient, nil
}
