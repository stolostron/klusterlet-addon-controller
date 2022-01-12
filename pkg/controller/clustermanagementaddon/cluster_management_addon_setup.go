// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package clustermanagementaddon

import (
	"context"
	"fmt"
	"reflect"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	addons "github.com/stolostron/klusterlet-addon-controller/pkg/components"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "github.com/prometheus/common/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Constants for ClusterManagementAddons Names
var (
	ApplicationManager   = addons.AppMgr.GetManagedClusterAddOnName()
	CertPolicyController = addons.CertCtrl.GetManagedClusterAddOnName()
	IamPolicyController  = addons.IAMCtrl.GetManagedClusterAddOnName()
	PolicyController     = addons.PolicyCtrl.GetManagedClusterAddOnName()
	SearchCollector      = addons.Search.GetManagedClusterAddOnName()
	WorkManager          = addons.WorkMgr.GetManagedClusterAddOnName()
)

// ClusterManagementAddOnNames - list of clustermanagementaddon name for addon
var ClusterManagementAddOnNames = []string{
	ApplicationManager,
	CertPolicyController,
	IamPolicyController,
	PolicyController,
	SearchCollector,
	WorkManager,
}

// clusterManagementAddOnSpec holds DisplayName, Description and CRDName
type clusterManagementAddOnSpec struct {
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	CRDName     string `json:"crdName"`
}

// ClusterManagementAddOnMap - map to hold clusterManagementAddOn spec information
var ClusterManagementAddOnMap = map[string]clusterManagementAddOnSpec{
	ApplicationManager: clusterManagementAddOnSpec{
		DisplayName: "Application Manager",
		Description: "Processes events and other requests to managed resources.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
	CertPolicyController: clusterManagementAddOnSpec{
		DisplayName: "Cert Policy Controller",
		Description: "Monitors certificate expiration based on distributed policies.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
	IamPolicyController: clusterManagementAddOnSpec{
		DisplayName: "IAM Policy Controller",
		Description: "Monitors identity controls based on distributed policies.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
	PolicyController: clusterManagementAddOnSpec{
		DisplayName: "Policy Controller",
		Description: "Distributes configured policies and monitors Kubernetes-based policies.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
	SearchCollector: clusterManagementAddOnSpec{
		DisplayName: "Search Collector",
		Description: "Collects cluster data to be indexed by search components on the hub cluster.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
	WorkManager: clusterManagementAddOnSpec{
		DisplayName: "Work Manager",
		Description: "Handles endpoint work requests and managed cluster status.",
		CRDName:     "klusterletaddonconfigs.agent.open-cluster-management.io",
	},
}

// CreateClusterManagementAddon - creates ClusterManagementAddOns for all add-ons in klusterletaddonconfig
func CreateClusterManagementAddon(c client.Client) {
	for !getAllClusterManagementAddons(c) {
		for _, name := range ClusterManagementAddOnNames {
			clusterManagementAddon := &addonv1alpha1.ClusterManagementAddOn{}
			clusterManagementAddonSpec := ClusterManagementAddOnMap[name]
			if err := c.Get(context.TODO(), types.NamespacedName{Name: name}, clusterManagementAddon); err != nil {
				if errors.IsNotFound(err) {
					clusterManagementAddon := newClusterManagementAddon(name, clusterManagementAddonSpec)
					if err := c.Create(context.TODO(), clusterManagementAddon); err != nil {
						log.Error(err, fmt.Sprintf("Failed to create %s clustermanagementaddon ", name))
						return
					}
					log.Info(fmt.Sprintf("Create %s clustermanagementaddon", name))
					continue
				}
				switch err.(type) {
				case *cache.ErrCacheNotStarted:
					time.Sleep(time.Second)
					continue
				default:
					log.Error(err, fmt.Sprintf("Failed to get %s clustermanagementaddon", name))
					return
				}
			}
			log.Info(fmt.Sprintf("%s clustermanagementaddon is found", name))
			continue
		}
	}
}

func newClusterManagementAddon(addOnName string, clusterManagementAddonSpec clusterManagementAddOnSpec) *addonv1alpha1.ClusterManagementAddOn {
	return &addonv1alpha1.ClusterManagementAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ClusterManagementAddOn",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: addOnName,
		},
		Spec: addonv1alpha1.ClusterManagementAddOnSpec{
			AddOnMeta: addonv1alpha1.AddOnMeta{
				DisplayName: clusterManagementAddonSpec.DisplayName,
				Description: clusterManagementAddonSpec.Description,
			},
			AddOnConfiguration: addonv1alpha1.ConfigCoordinates{
				CRDName: clusterManagementAddonSpec.CRDName,
			},
		},
	}
}

func updateClusterManagementAddOn(client client.Client, addOnName string, oldClusterManagementAddOn *addonv1alpha1.ClusterManagementAddOn) error {
	clusterManagementAddonMeta := ClusterManagementAddOnMap[addOnName]

	newClusterManagementAddon := newClusterManagementAddon(addOnName, clusterManagementAddonMeta)
	if !reflect.DeepEqual(oldClusterManagementAddOn.Spec, newClusterManagementAddon.Spec) {
		oldClusterManagementAddOn.Spec = newClusterManagementAddon.Spec
		if err := client.Update(context.TODO(), oldClusterManagementAddOn); err != nil {
			log.Error(err, fmt.Sprintf("Failed to update %s clustermanagementaddon ", addOnName))
			return err
		}
	}
	return nil
}

func getAllClusterManagementAddons(client client.Client) bool {
	for _, name := range ClusterManagementAddOnNames {
		clusterManagementAddon := &addonv1alpha1.ClusterManagementAddOn{}
		if err := client.Get(context.TODO(), types.NamespacedName{Name: name}, clusterManagementAddon); err != nil {
			return false
		}
	}
	return true
}
