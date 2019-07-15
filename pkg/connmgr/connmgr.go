package connmgr

import (
	"k8s.io/kubernetes/pkg/apis/core"
)

// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// ConnectionManager defines the desired state of the ConnectionManager component
// +k8s:openapi-gen=true
type ConnectionManager struct {
	Enabled bool `json:"enabled,omitempty"`

	LogLevel int `json:"logLevel,omitempty"`
	Replicas int `json:"replicas,omitempty"`

	// +kubebuilder:validation:MinLength=1
	Hubs []Hub `json:"hubs,omitempty"`

	Affinity        core.Affinity             `json:"affinity,omitempty"`
	Image           Image                     `json:"image,omitempty"`
	ImagePullSecret core.LocalObjectReference `json:"imagePullSecret,omitempty"`
	NodeSelector    core.NodeSelector         `json:"nodeSelector,omitempty"`
	Tolerations     []core.Toleration         `json:"tolerations,omitempty"`
}

// Image defines the image to pull for a container
// +k8s:openapi-gen=true
type Image struct {
	// +kubebuilder:validation:MinLength=1
	Repository string `json:"repository,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:validation:Enum=Always,Never,IfNotPresent
	PullPolicy core.PullPolicy `json:"pullPolicy,omitempty"`
}

// Hub defines an MCM hub for the ConnectionManager component
type Hub struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Secret core.LocalObjectReference `json:"secret,omitempty"`
}
