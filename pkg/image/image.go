package image

import (
	"k8s.io/kubernetes/pkg/apis/core"
)

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
