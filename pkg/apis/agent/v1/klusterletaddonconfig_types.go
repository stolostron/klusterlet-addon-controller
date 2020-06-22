// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KlusterletAddonConfigSpec defines the desired state of KlusterletAddonConfig
type KlusterletAddonConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	ClusterLabels map[string]string `json:"clusterLabels"`

	SearchCollectorConfig      KlusterletAddonConfigSearchCollectorSpec      `json:"searchCollector"`
	PolicyController           KlusterletAddonConfigPolicyControllerSpec     `json:"policyController"`
	ApplicationManagerConfig   KlusterletAddonConfigApplicationManagerSpec   `json:"applicationManager"`
	CertPolicyControllerConfig KlusterletAddonConfigCertPolicyControllerSpec `json:"certPolicyController"`
	CISControllerConfig        KlusterletAddonConfigCISControllerSpec        `json:"cisController"`
	IAMPolicyControllerConfig  KlusterletAddonConfigIAMPolicyControllerSpec  `json:"iamPolicyController"`

	ImageRegistry    string `json:"imageRegistry,omitempty"`
	ImageNamePostfix string `json:"imageNamePostfix,omitempty"`
	// +kubebuilder:validation:MinLength=1
	ImagePullSecret string `json:"imagePullSecret,omitempty"`

	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// // ComponentTagMap contains the tag of each component
	// ComponentTagMap map[string]string `json:"componentTagMap"`
	// // ComponentImageMap contains the image name of each component
	// ComponentImageMap map[string]string `json:"componentImageMap"`

	// used for dev work only
	ComponentOperatorImage string `json:"componentOperatorImage,omitempty"`
}

// KlusterletAddonConfigApplicationManagerSpec defines configuration for the ApplicationManager component
type KlusterletAddonConfigApplicationManagerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletAddonConfigSearchCollectorSpec defines configuration for the SearchCollector component
type KlusterletAddonConfigSearchCollectorSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletAddonConfigCertPolicyControllerSpec defines configuration for the CertPolicyController component
type KlusterletAddonConfigCertPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletAddonConfigIAMPolicyControllerSpec defines configuration for the IAMPolicyController component
type KlusterletAddonConfigIAMPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletAddonConfigWorkManagerSpec defines configuration for the WorkManager component
type KlusterletAddonConfigWorkManagerSpec struct {
	ClusterLabels map[string]string `json:"clusterLabels"`
}

// KlusterletAddonConfigPolicyControllerSpec defines configuration for the PolicyController component
type KlusterletAddonConfigPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletPrometheusIntegrationSpec defines configuration for the Promtheus Integration
type KlusterletPrometheusIntegrationSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletAddonConfigCISControllerSpec defines configuration for the CISController component
type KlusterletAddonConfigCISControllerSpec struct {
	Enabled bool `json:"enabled"`
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
