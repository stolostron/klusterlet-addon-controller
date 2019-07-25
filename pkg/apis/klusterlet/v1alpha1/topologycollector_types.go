//Package v1alpha1 Defines the API to support Multicluster Endpoints (klusterlets).
//IBM Confidential
//OCO Source Materials
//5737-E67
//(C) Copyright IBM Corporation 2019 All Rights Reserved
//The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TopologyCollectorSpec defines the desired state of TopologyCollector
// +k8s:openapi-gen=true
type TopologyCollectorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Enabled bool `json:"enabled"`
	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`
	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`
	// +kubebuilder:validation:MinLength=1
	ContainerRuntime string `json:"containerRuntime"`
	UpdateInterval   int32  `json:"updateInterval"`
	// Weave             clusterImage.Image        `json:"weave"`
	// Collector         clusterImage.Image        `json:"collector"`
	// Router            clusterImage.Image        `json:"router"`
	// ImagePullSecret   core.LocalObjectReference `json:"imagePullSecret"`
	FullNameOverride  string                          `json:"fullnameOverride"`
	ConnectionManager string                          `json:"connectionManager"`
	CACertIssuer      string                          `json:"caCertIssuer"`
	ServiceAccount    TopologyCollectorServiceAccount `json:"serviceAccount"`

	// Affinity     core.Affinity     `json:"affinity"`
	// NodeSelector core.NodeSelector `json:"nodeSelector"`
	// Tolerations  []core.Toleration `json:"tolerations"`
}

// TopologyCollectorServiceAccount defines service account configuration in the spec
// +k8s:openapi-gen=true
type TopologyCollectorServiceAccount struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

// TopologyCollectorStatus defines the observed state of TopologyCollector
// +k8s:openapi-gen=true
type TopologyCollectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopologyCollector is the Schema for the topologycollectors API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type TopologyCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopologyCollectorSpec   `json:"spec,omitempty"`
	Status TopologyCollectorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopologyCollectorList contains a list of TopologyCollector
type TopologyCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopologyCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TopologyCollector{}, &TopologyCollectorList{})
}
