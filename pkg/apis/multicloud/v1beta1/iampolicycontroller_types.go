// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package v1beta1

import (
	"github.com/open-cluster-management/endpoint-operator/pkg/image"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IAMPolicyControllerSpec defines the desired state of IAMPolicyController
type IAMPolicyControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	// +kubebuilder:validation:MinLength=1
	ConnectionManager string `json:"connectionManager"`

	Image           image.Image `json:"image,omitempty"`
	ImagePullSecret string      `json:"imagePullSecret,omitempty"`

	IsOpenShift bool `json:"ocp,omitempty"`
}

// IAMPolicyControllerStatus defines the observed state of IAMPolicyController
type IAMPolicyControllerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IAMPolicyController is the Schema for the iampolicycontroller API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=iampoliciescontroller,scope=Namespaced
type IAMPolicyController struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMPolicyControllerSpec   `json:"spec,omitempty"`
	Status IAMPolicyControllerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IAMPolicyControllerList contains a list of IAMPolicyController
type IAMPolicyControllerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMPolicyController `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IAMPolicyController{}, &IAMPolicyControllerList{})
}
