// Copyright (c) 2020 Red Hat, Inc.
package components

import (
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KlusterletAddon interface {
	// GetAddonName retuns the addon name
	GetAddonName() string
	// RequiresHubKubeconfig returns true if this addon need to genrate a kubeconfig on hubside
	CheckHubKubeconfigRequired() bool
	// IsEnabled checks whether the addon is enabled in the klusterletaddonconfig
	IsEnabled(instance *agentv1.KlusterletAddonConfig) bool
	// NewAddonCR returns a CR of the addon by using the given klusterletaddonconfig & managedcluster's namespace
	NewAddonCR(instance *agentv1.KlusterletAddonConfig, namespace string) (runtime.Object, error)
}
