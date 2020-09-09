// Copyright (c) 2020 Red Hat, Inc.
package components

import (
	"fmt"
	"strings"

	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	appmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/endpoint-operator/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/endpoint-operator/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/endpoint-operator/pkg/components/workmgr/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	manifestworkMidName = "-klusterlet-addon-"
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
	// GetManagedClusterAddOnName returns the ManagedClusterAddOn name that matches this addon
	GetManagedClusterAddOnName() string
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
var addonMap map[string]KlusterletAddon

var managedClusterAddOnNameMap map[string]KlusterletAddon

func init() {
	addonMap = make(map[string]KlusterletAddon)
	managedClusterAddOnNameMap = make(map[string]KlusterletAddon)

	for _, addon := range AddonsArray {
		addonMap[addon.GetAddonName()] = addon
		managedClusterAddOnNameMap[addon.GetManagedClusterAddOnName()] = addon
	}
}

// ConstructManifestWorkName create a manifestwork name
func ConstructManifestWorkName(instance *agentv1.KlusterletAddonConfig, addon KlusterletAddon) string {
	return instance.Name + manifestworkMidName + addon.GetAddonName()
}

// GetAddonFromManifestWorkName returns KlusterletAddon given a manifestwork's name
// this is possible because we always use same naming convention for manifestwork in `ConstructManifestWorkName`
// will return error if failed to find a match
func GetAddonFromManifestWorkName(manifestworkName string) (KlusterletAddon, error) {
	err := fmt.Errorf("failed to find addon from ManifestWork %s", manifestworkName)
	// get addon name
	idx := strings.LastIndex(manifestworkName, manifestworkMidName)
	if idx < 0 || idx+len(manifestworkMidName) >= len(manifestworkName) {
		return nil, err
	}
	// get manifestwork name
	idx += len(manifestworkMidName)

	addonName := manifestworkName[idx:]
	if addon, ok := addonMap[addonName]; ok {
		return addon, nil
	}

	return nil, err
}

// GetAddonFromManifestWorkName returns KlusterletAddon given a managedclusteraddon's name
// this is possible because we always have addons with unique managedclusteraddon names
// will return error if failed to find a match
func GetAddonFromManagedClusterAddonName(name string) (KlusterletAddon, error) {
	err := fmt.Errorf("failed to find addon from ManagedClusterAddOn %s", name)
	if addon, ok := managedClusterAddOnNameMap[name]; ok {
		return addon, nil
	}
	return nil, err
}
