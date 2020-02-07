// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplicationManagerSpec defines the desired state of ApplicationManager
type ApplicationManagerSpec struct {
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

	DeployableSpec   ApplicationManagerDeployableSpec   `json:"deployable"`
	SubscriptionSpec ApplicationManagerSubscriptionSpec `json:"subscription"`

	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

// ApplicationManagerDeployableSpec defines configuration for Deployable in ApplicationManager
type ApplicationManagerDeployableSpec struct {
	Image image.Image `json:"image"`
}

// ApplicationManagerSubscriptionSpec defines configuration for Subscription in ApplicationManager
type ApplicationManagerSubscriptionSpec struct {
	Image image.Image `json:"image"`
}

// ApplicationManagerStatus defines the observed state of ApplicationManager
type ApplicationManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationManager is the Schema for the applicationmanagers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=applicationmanagers,scope=Namespaced
type ApplicationManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationManagerSpec   `json:"spec,omitempty"`
	Status ApplicationManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationManagerList contains a list of ApplicationManager
type ApplicationManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ApplicationManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ApplicationManager{}, &ApplicationManagerList{})
}
