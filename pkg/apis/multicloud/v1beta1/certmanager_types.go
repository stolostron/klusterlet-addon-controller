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

// CertManagerSpec defines the desired state of CertManager
// +k8s:openapi-gen=true
type CertManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	// +kubebuilder:validation:MinLength=1
	ClusterResourceNamespace string `json:"clusterResourceNamespace"`

	// +kubebuilder:validation:MinLength=1
	ServiceAccount CertManagerServiceAccount `json:"serviceAccount"`

	Image            image.Image                     `json:"image,omitempty"`
	PolicyController CertManagerPolicyControllerSpec `json:"policy-controller"`
}

// CertManagerPolicyControllerSpec defines policy-controller enabledment in ICP certmanager chart
// +k8s:openapi-gen=true
type CertManagerPolicyControllerSpec struct {
	// +kubebuilder:validation:MinLength=1
	Enabled bool `json:"enabled"`
}

// CertManagerServiceAccount defines service account configuration in the spec
// +k8s:openapi-gen=true
type CertManagerServiceAccount struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// CertManagerStatus defines the observed state of CertManager
// +k8s:openapi-gen=true
type CertManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManager is the Schema for the certmanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type CertManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertManagerSpec   `json:"spec,omitempty"`
	Status CertManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CertManagerList contains a list of CertManager
type CertManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertManager{}, &CertManagerList{})
}
