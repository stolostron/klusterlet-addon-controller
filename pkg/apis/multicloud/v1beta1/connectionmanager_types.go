// Package v1beta1 of apis contain the API type definition for the components
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConnectionManagerSpec defines the desired state of ConnectionManager
// +k8s:openapi-gen=true
type ConnectionManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	BootStrapConfig map[string]string `json:"bootstrapConfig"`

	Image           image.Image                     `json:"image,omitempty"`
	ImagePullSecret string                          `json:"imagePullSecret,omitempty"`
	GlobalView      ConnectionManagerGlobalViewSpec `json:"globalView"`
}

// ConnectionManagerGlobalViewSpec defines the spec for connection manager global view
// +k8s:openapi-gen=true
type ConnectionManagerGlobalViewSpec struct {
	Enabled         bool   `json:"enabled"`
	CollectorLabels string `json:"collectorLabels"`
}

// ConnectionManagerStatus defines the observed state of ConnectionManager
// +k8s:openapi-gen=true
type ConnectionManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConnectionManager is the Schema for the connectionmanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ConnectionManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectionManagerSpec   `json:"spec,omitempty"`
	Status ConnectionManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConnectionManagerList contains a list of ConnectionManager
type ConnectionManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectionManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConnectionManager{}, &ConnectionManagerList{})
}
