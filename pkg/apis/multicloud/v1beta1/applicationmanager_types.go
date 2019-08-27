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

// ApplicationManagerSpec defines the desired state of ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerSpec struct {
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

	TillerIntegration ApplicationManagerTillerIntegration `json:"tillerIntegration"`

	DeployableSpec                 ApplicationManagerDeployableSpec                 `json:"deployable"`
	SubscriptionSpec               ApplicationManagerSubscriptionSpec               `json:"subscription"`
	HelmCRDSpec                    ApplicationManagerHelmCRDSpec                    `json:"helmcrd"`
	HelmCRDAdmissionControllerSpec ApplicationManagerHelmCRDAdmissionControllerSpec `json:"helmCRDAdmissionController"`

	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

// ApplicationManagerDeployableSpec defines configuration for Deployable in ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerDeployableSpec struct {
	Image image.Image `json:"image"`
}

// ApplicationManagerTillerIntegration defines tiller integration parameters
// +k8s:openapi-gen=true
type ApplicationManagerTillerIntegration struct {
	Enabled       bool   `json:"enabled"`
	Endpoint      string `json:"endpoint"`
	CertIssuer    string `json:"certIssuer"`
	AutoGenSecret bool   `json:"autoGenSecret"`
	User          string `json:"user"`
}

// ApplicationManagerSubscriptionSpec defines configuration for Subscription in ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerSubscriptionSpec struct {
	Image image.Image `json:"image"`
}

// ApplicationManagerHelmCRDSpec defines configuration for HelmCRD in ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerHelmCRDSpec struct {
	Image    image.Image `json:"image"`
	IP       string      `json:"ip"`
	Hostname string      `json:"hostname"`
}

// ApplicationManagerHelmCRDAdmissionControllerSpec defines configuration for HelmCRDAdmissionController in ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerHelmCRDAdmissionControllerSpec struct {
	Image    image.Image `json:"image"`
	CABundle string      `json:"caBundle"`
}

// ApplicationManagerStatus defines the observed state of ApplicationManager
// +k8s:openapi-gen=true
type ApplicationManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApplicationManager is the Schema for the applicationmanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
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
