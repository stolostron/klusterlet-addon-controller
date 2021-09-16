// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TODO：1. move ClusterName, ClusterNamespace, NodeSelector, ImageRegistry, ImagePullSecret and ImagePullPolicy to
//  internal variables since they do not need to be customized configured.
// TODO：2. refactor addon config spec using an unified definition.

// KlusterletAddonConfigSpec defines the desired state of KlusterletAddonConfig
type KlusterletAddonConfigSpec struct {
	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +optional
	Version string `json:"version,omitempty"`

	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +kubebuilder:validation:MinLength=1
	// +optional
	ClusterName string `json:"clusterName,omitempty"`

	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +kubebuilder:validation:MinLength=1
	// +optional
	ClusterNamespace string `json:"clusterNamespace,omitempty"`

	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +optional
	ClusterLabels map[string]string `json:"clusterLabels,omitempty"`

	// NodeSelector defines which Nodes the Pods are scheduled on. The default is an empty list.
	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// GlobalProxy defines the cluster-wide proxy configuration of managed cluster.
	// +optional
	GlobalProxy ProxyConfig `json:"globalProxy,omitempty"`

	SearchCollectorConfig      KlusterletAddonAgentConfigSpec `json:"searchCollector"`
	PolicyController           KlusterletAddonAgentConfigSpec `json:"policyController"`
	ApplicationManagerConfig   KlusterletAddonAgentConfigSpec `json:"applicationManager"`
	CertPolicyControllerConfig KlusterletAddonAgentConfigSpec `json:"certPolicyController"`
	IAMPolicyControllerConfig  KlusterletAddonAgentConfigSpec `json:"iamPolicyController"`

	// ImageRegistry defined the custom registry address of the images.
	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +optional
	ImageRegistry string `json:"imageRegistry,omitempty"`

	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +kubebuilder:validation:MinLength=1
	ImagePullSecret string `json:"imagePullSecret,omitempty"`

	// DEPRECATED in release 2.4 and will be removed in release 2.5 since not used anymore.
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

// ProxyConfig defines the global proxy env for OCP cluster
type ProxyConfig struct {
	// HTTPProxy is the URL of the proxy for HTTP requests.  Empty means unset and will not result in an env var.
	// +optional
	HTTPProxy string `json:"httpProxy,omitempty"`

	// HTTPSProxy is the URL of the proxy for HTTPS requests.  Empty means unset and will not result in an env var.
	// +optional
	HTTPSProxy string `json:"httpsProxy,omitempty"`

	// NoProxy is a comma-separated list of hostnames and/or CIDRs for which the proxy should not be used.
	// Empty means unset and will not result in an env var.
	// +optional
	NoProxy string `json:"noProxy,omitempty"`
}

type GlobalProxyStatus string

const (
	GlobalProxyStatusTrue  GlobalProxyStatus = "true"
	GlobalProxyStatusFalse GlobalProxyStatus = "false"
)

// KlusterletAddonAgentConfigSpec defines configuration for each addon agent.
type KlusterletAddonAgentConfigSpec struct {
	// Enabled is the flag to enable/disable the addon. default is false.
	// +optional
	Enabled bool `json:"enabled"`

	// EnableGlobalProxy is the flag to enable/disable the GlobalProxy configuration for the pods of addon.
	// default is false.
	// +kubebuilder:validation:Enum=true;false
	// +optional
	EnableGlobalProxy GlobalProxyStatus `json:"enableGlobalProxy,omitempty"`
}

// KlusterletAddonConfigStatus defines the observed state of KlusterletAddonConfig
type KlusterletAddonConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KlusterletAddonConfig is the Schema for the klusterletaddonconfigs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=klusterletaddonconfigs,scope=Namespaced
type KlusterletAddonConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlusterletAddonConfigSpec   `json:"spec,omitempty"`
	Status KlusterletAddonConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KlusterletAddonConfigList contains a list of klusterletAddonConfig
type KlusterletAddonConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KlusterletAddonConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KlusterletAddonConfig{}, &KlusterletAddonConfigList{})
}
