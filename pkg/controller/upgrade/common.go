package upgrade

import (
	"fmt"
)

const (
	klusterletAddonOperator   = "klusterlet-addon-operator"
	klusterletAddonCRDs       = "klusterlet-addon-crds"
	clusterImageRegistryLabel = "open-cluster-management.io/image-registry"

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
