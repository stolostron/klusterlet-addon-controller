package common

const (
	// AnnotationKlusterletDeployMode is the annotation key of klusterlet deploy mode
	AnnotationKlusterletDeployMode = "import.open-cluster-management.io/klusterlet-deploy-mode"

	// AnnotationEnableHostedModeAddons is the key of annotation which indicates if the add-ons will be enabled
	// in hosted mode automatically for a managed cluster
	AnnotationEnableHostedModeAddons = "addon.open-cluster-management.io/enable-hosted-mode-addons"

	// AnnotationKlusterletHostingClusterName is the annotation key of hosting cluster name for klusterlet
	AnnotationKlusterletHostingClusterName = "import.open-cluster-management.io/hosting-cluster-name"

	// AnnotationAddOnHostingClusterName is the annotation key of hosting cluster name for add-ons
	AnnotationAddOnHostingClusterName = "addon.open-cluster-management.io/hosting-cluster-name"
)
