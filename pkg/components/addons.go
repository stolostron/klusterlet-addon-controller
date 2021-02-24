// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package components

import (
	"fmt"
	"strings"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	appmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/workmgr/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	manifestworkMidName = "-klusterlet-addon-"
)

var log = logf.Log.WithName("addons")

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

var AppMgr = appmgr.AddonAppMgr{}
var CertCtrl = certpolicyctrl.AddonCertPolicyCtrl{}
var IAMCtrl = iampolicyctrl.AddonIAMPolicyCtrl{}
var PolicyCtrl = policyctrl.AddonPolicyCtrl{}
var Search = search.AddonSearch{}
var WorkMgr = workmgr.AddonWorkMgr{}

// AddonsArray are all addons we support in this repo
// each one's GetAddonName() should be unique
// each one's GetManagedClusterAddOnName() should also be unique
var AddonsArray = []KlusterletAddon{
	AppMgr,
	CertCtrl,
	IAMCtrl,
	PolicyCtrl,
	Search,
	WorkMgr,
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

// NewAddonNamePredicate allows addon object with a name can be converted to an addon
// to reconcile. The addon object can be ManagedClusterAddons, ClusterManagementAddons,
// or Leases.
func NewAddonNamePredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			if _, err := GetAddonFromManagedClusterAddonName(e.Meta.GetName()); err != nil {
				return false
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			if _, err := GetAddonFromManagedClusterAddonName(e.Meta.GetName()); err != nil {
				return false
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld == nil || e.MetaNew == nil ||
				e.ObjectOld == nil || e.ObjectNew == nil {
				log.Error(nil, "Update event is invalid", "event", e)
				return false
			}
			if _, err := GetAddonFromManagedClusterAddonName(e.MetaOld.GetName()); err != nil {
				return false
			}
			if _, err := GetAddonFromManagedClusterAddonName(e.MetaNew.GetName()); err != nil {
				return false
			}
			return true
		},
	})
}
