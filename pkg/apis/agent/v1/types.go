package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// GlobalValues defines the global values
// +k8s:openapi-gen=true
type GlobalValues struct {
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	ImagePullSecret string            `json:"imagePullSecret,omitempty"`
	ImageOverrides  map[string]string `json:"imageOverrides,omitempty"`
}
