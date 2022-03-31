package upgrade

import (
	"context"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanupOldClusterManagementAddon is to remove the AddOnConfiguration.CRDName klusterletaddonconfigs
func CleanupOldClusterManagementAddon(c client.Client) {
	for _, addonName := range agentv1.ClusterManagementAddons {
		clusterManagementAddon := &addonv1alpha1.ClusterManagementAddOn{}
		if err := c.Get(context.TODO(), types.NamespacedName{Name: addonName}, clusterManagementAddon); err != nil {
			if errors.IsNotFound(err) {
				continue
			}

			klog.Errorf("failed to get clusterManagementAddon %v", addonName)
			continue
		}

		// need to update the spec to remove AddOnConfiguration.CRDName
		newAddon := clusterManagementAddon.DeepCopy()
		if newAddon.Spec.AddOnConfiguration.CRDName == "klusterletaddonconfigs.agent.open-cluster-management.io" {
			newAddon.Spec.AddOnConfiguration.CRDName = ""
			if err := c.Update(context.TODO(), newAddon, &client.UpdateOptions{}); err != nil {
				klog.Errorf("failed to update clusterManagementAddon %v", addonName)
			}
		}
	}
}
