package conn_mgr

import (
	"k8s.io/kubernetes/pkg/apis/core"
)

// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ConnectionManager defines the desired state of the ConnectionManager component
// +k8s:openapi-gen=true
type ConnectionManager struct {
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`
	LogLevel    int    `json:"logLevel"`
	Replicas    int    `json:"replicas"`

	// +kubebuilder:validation:MinLength=1
	Hubs []Hub `json:"hubs"`

	Affinity        core.Affinity             `json:"affinity"`
	Image           Image                     `json:"image"`
	ImagePullSecret core.LocalObjectReference `json:"imagePullSecret"`
	NodeSelector    core.NodeSelector         `json:"nodeSelector"`
	Tolerations     []core.Toleration         `json:"tolerations"`
}

// Image defines the image to pull for a container
// +k8s:openapi-gen=true
type Image struct {
	// +kubebuilder:validation:MinLength=1
	Repository string `json:"repository"`
	// +kubebuilder:validation:MinLength=1
	Tag string `json:"tag"`
	// +kubebuilder:validation:Enum=Always,Never,IfNotPresent
	PullPolicy core.PullPolicy `json:"pullPolicy"`
}

// Hub defines an MCM hub for the ConnectionManager component
type Hub struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Secret core.LocalObjectReference `json:"secret"`
}
