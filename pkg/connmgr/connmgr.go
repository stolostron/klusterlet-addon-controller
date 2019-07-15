package connmgr

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"

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
	Image           image.Image               `json:"image,omitempty"`
	ImagePullSecret core.LocalObjectReference `json:"imagePullSecret,omitempty"`
	NodeSelector    core.NodeSelector         `json:"nodeSelector,omitempty"`
	Tolerations     []core.Toleration         `json:"tolerations,omitempty"`
}

// Hub defines an MCM hub for the ConnectionManager component
type Hub struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Secret core.LocalObjectReference `json:"secret,omitempty"`
}
