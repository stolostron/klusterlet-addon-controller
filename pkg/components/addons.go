// Copyright (c) 2020 Red Hat, Inc.
package components

import (
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/runtime"

	appmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/endpoint-operator/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/workmgr/v1"
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

// AddonsArray are all addons we support in this repo
// each one's GetAddonName() should be unique
// each one's GetManagedClusterAddOnName() should also be unique
var AddonsArray = []KlusterletAddon{
	appmgr.AddonAppMgr{},
	certpolicyctrl.AddonCertPolicyCtrl{},
	iampolicyctrl.AddonIAMPolicyCtrl{},
	policyctrl.AddonPolicyCtrl{},
	search.AddonSearch{},
	workmgr.AddonWorkMgr{},
}
