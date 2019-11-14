// Package v1beta1 of apis contain the API type definition for the components
// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TillerSpec defines the desired state of Tiller
// +k8s:openapi-gen=true
type TillerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	// +kubebuilder:validation:MinLength=1
	CACertIssuer string `json:"caCertIssuer"`

	// +kubebuilder:validation:MinLength=1
	DefaultAdminUser string `json:"tiller_default_admin_user"`

	Image           image.Image `json:"image,omitempty,omitempty"`
	ImagePullSecret string      `json:"imagePullSecret,omitempty"`
	KubeClusterType string      `json:"kubernetes_cluster_type"`
}

// TillerStatus defines the observed state of Tiller
// +k8s:openapi-gen=true
type TillerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tiller is the Schema for the tillers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Tiller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TillerSpec   `json:"spec,omitempty"`
	Status TillerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TillerList contains a list of Tiller
type TillerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tiller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tiller{}, &TillerList{})
}
