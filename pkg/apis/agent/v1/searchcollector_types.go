// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SearchCollectorSpec defines the desired state of SearchCollector
type SearchCollectorSpec struct {
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
	HubKubeconfigSecret string `json:"hubKubeconfigSecret"`

	GlobalValues GlobalValues `json:"global,omitempty"`
}

// SearchCollectorStatus defines the observed state of SearchCollector
type SearchCollectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SearchCollector is the Schema for the searchcollectors API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=searchcollectors,scope=Namespaced
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
