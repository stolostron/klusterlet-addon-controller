// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KlusterletSpec defines the desired state of Klusterlet
type KlusterletSpec struct {
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

	BootStrapConfig map[string]string `json:"bootstrapConfig,omitempty"`

	SearchCollectorConfig      KlusterletSearchCollectorSpec      `json:"searchCollector"`
	PolicyController           KlusterletPolicyControllerSpec     `json:"policyController"`
	ApplicationManagerConfig   KlusterletApplicationManagerSpec   `json:"applicationManager"`
	ConnectionManagerConfig    KlusterletConnectionManagerSpec    `json:"connectionManager"`
	CertPolicyControllerConfig KlusterletCertPolicyControllerSpec `json:"certPolicyController"`
	CISControllerConfig        KlusterletCISControllerSpec        `json:"cisController"`
	IAMPolicyControllerConfig  KlusterletIAMPolicyControllerSpec  `json:"iamPolicyController"`

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

// KlusterletConnectionManagerSpec defines configuration for the ConnectionManager component
type KlusterletConnectionManagerSpec struct {
}

// KlusterletApplicationManagerSpec defines configuration for the ApplicationManager component
type KlusterletApplicationManagerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletSearchCollectorSpec defines configuration for the SearchCollector component
type KlusterletSearchCollectorSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletCertPolicyControllerSpec defines configuration for the CertPolicyController component
type KlusterletCertPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletIAMPolicyControllerSpec defines configuration for the IAMPolicyController component
type KlusterletIAMPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletWorkManagerSpec defines configuration for the WorkManager component
type KlusterletWorkManagerSpec struct {
	ClusterLabels map[string]string `json:"clusterLabels"`
}

// KlusterletPolicyControllerSpec defines configuration for the PolicyController component
type KlusterletPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletPrometheusIntegrationSpec defines configuration for the Promtheus Integration
type KlusterletPrometheusIntegrationSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletCISControllerSpec defines configuration for the CISController component
type KlusterletCISControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// KlusterletStatus defines the observed state of Klusterlet
type KlusterletStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Klusterlet is the Schema for the klusterlets API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=klusterlets,scope=Namespaced
type Klusterlet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlusterletSpec   `json:"spec,omitempty"`
	Status KlusterletStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KlusterletList contains a list of Klusterlet
type KlusterletList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Klusterlet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Klusterlet{}, &KlusterletList{})
}
