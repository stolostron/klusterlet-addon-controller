package upgrade

import (
	"fmt"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	klusterletAddonOperator = "klusterlet-addon-operator"
	klusterletAddonCRDs     = "klusterlet-addon-crds"

	// AnnotationNodeSelector key name of nodeSelector annotation synced from mch
	AnnotationNodeSelector = "open-cluster-management/nodeSelector"
)

func manifestWorkName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}
func roleBindingName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s-v2", clusterName, name)
}

func agentManifestWorkName(clusterName, componentName string) string {
	return fmt.Sprintf("%s-klusterlet-addon-%s", clusterName, componentName)
}

func upgradePredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			_, existed := agentv1.DeprecatedAddonComponentNames[e.Meta.GetName()]
			return existed
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				klog.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			_, existed := agentv1.DeprecatedAddonComponentNames[e.Meta.GetName()]
			return existed
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.MetaOld == nil || e.MetaNew == nil ||
				e.ObjectOld == nil || e.ObjectNew == nil {
				klog.Error(nil, "Update event is invalid", "event", e)
				return false
			}
			_, existed := agentv1.DeprecatedAddonComponentNames[e.MetaNew.GetName()]
			return existed
		},
	})
}
