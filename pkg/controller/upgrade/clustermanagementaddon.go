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

// CleanupOldClusterManagementAddon deletes the old clusterManagementAddons.
// some clusterManagementAddons are deployed by helm, will meet upgrade failures if the existed clusterManagementAddon
// is not deployed by helm. so need to delete the old clusterManagementAddons which is not deployed by helm.
// applicationAddon is an exception because it is not deployed by helm in 2.5.
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

		// new workManager and application clusterManagementAddon are created before addon-controller pod.
		// so need to update the spec to remove AddOnConfiguration.CRDName
		if clusterManagementAddon.Name == agentv1.WorkManagerAddonName ||
			clusterManagementAddon.Name == agentv1.ApplicationAddonName {
			newAddon := clusterManagementAddon.DeepCopy()
			if newAddon.Spec.AddOnConfiguration.CRDName == "klusterletaddonconfigs.agent.open-cluster-management.io" {
				newAddon.Spec.AddOnConfiguration.CRDName = ""
				if err := c.Update(context.TODO(), newAddon, &client.UpdateOptions{}); err != nil {
					klog.Errorf("failed to update clusterManagementAddon %v", addonName)
				}
			}
			continue
		}

		annotations := clusterManagementAddon.GetAnnotations()
		createdByHelm := false
		for _, annotation := range annotations {
			if annotation == "meta.helm.sh/release-name" || annotation == "meta.helm.sh/release-namespace" {
				createdByHelm = true
				break
			}
		}
		if !createdByHelm {
			if err := c.Delete(context.TODO(), clusterManagementAddon, &client.DeleteOptions{}); err != nil {
				klog.Errorf("failed to delete old clusterManagementAddon %v", addonName)
			}
		}
	}
}
