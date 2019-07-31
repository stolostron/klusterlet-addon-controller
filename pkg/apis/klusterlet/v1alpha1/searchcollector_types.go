// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

// Package v1alpha1 provides search collector spec
package v1alpha1

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SearchCollectorSpec defines the desired state of SearchCollector
// +k8s:openapi-gen=true
type SearchCollectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	// +kubebuilder:validation:MinLength=1
	ConnectionManager string `json:"connectionManager"`

	TillerIntegration SearchCollectorTillerIntegration `json:"tillerIntegration"`

	Image           image.Image `json:"image,omitempty"`
	ImagePullSecret string      `json:"imagePullSecret,omitempty"`
}

// SearchCollectorTillerIntegration defines the observed state of SearchCollector
// +k8s:openapi-gen=true
type SearchCollectorTillerIntegration struct {
	Enabled       bool   `json:"enabled"`
	Endpoint      string `json:"endpoint"`
	CertIssuer    string `json:"certIssuer"`
	AutoGenSecret bool   `json:"autoGenSecret"`
	User          string `json:"user"`
}

// SearchCollectorStatus defines the observed state of SearchCollector
// +k8s:openapi-gen=true
type SearchCollectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SearchCollector is the Schema for the searchcollectors API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SearchCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SearchCollectorSpec   `json:"spec,omitempty"`
	Status SearchCollectorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SearchCollectorList contains a list of SearchCollector
type SearchCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SearchCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SearchCollector{}, &SearchCollectorList{})
}
