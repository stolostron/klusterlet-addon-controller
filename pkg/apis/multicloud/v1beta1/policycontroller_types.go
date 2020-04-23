// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PolicyControllerSpec defines the desired state of PolicyController
type PolicyControllerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`
	// +kubebuilder:validation:MinLength=1
	ConnectionManager string `json:"connectionManager"`
	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName,omitempty"`
	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace,omitempty"`
	DeployedOnHub    bool   `json:"deployedOnHub,omitempty"`

	GlobalValues GlobalValues `json:"global,omitempty"`

	PostDeleteJobServiceAccount string `json:"postDeleteJobServiceAccount,omitempty"`
}

// PolicyControllerStatus defines the observed state of PolicyController
type PolicyControllerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PolicyController is the Schema for the policycontrollers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=policycontrollers,scope=Namespaced
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
