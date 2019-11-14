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

// PolicyControllerSpec defines the desired state of PolicyController
// +k8s:openapi-gen=true
type PolicyControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`
	// +kubebuilder:validation:MinLength=1
	ConnectionManager string `json:"connectionManager"`
	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName,omitempty"`
	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace,omitempty"`
	// +kubebuilder:validation:MinLength=1
	DeployedOnHub bool `json:"deployedOnHub,omitempty"`
	//Image information for the PolicyController
	Image                       image.Image `json:"image,omitempty"`
	ImagePullSecret             string      `json:"imagePullSecret,omitempty"`
	PostDeleteJobServiceAccount string      `json:"postDeleteJobServiceAccount,omitempty"`
}

// PolicyControllerStatus defines the observed state of PolicyController
// +k8s:openapi-gen=true
type PolicyControllerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyController is the Schema for the policycontrollers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type PolicyController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicyControllerSpec   `json:"spec,omitempty"`
	Status PolicyControllerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyControllerList contains a list of PolicyController
type PolicyControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicyController `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PolicyController{}, &PolicyControllerList{})
}
