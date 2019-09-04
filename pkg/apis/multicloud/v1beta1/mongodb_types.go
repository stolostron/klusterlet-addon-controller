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

// MongoDBSpec defines the desired state of MongoDB
// +k8s:openapi-gen=true
type MongoDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	FullNameOverride      string                      `json:"fullnameOverride,omitempty"`
	Auth                  MongoDBSpecAuth             `json:"auth,omitempty"`
	Curl                  MongoDBSpecCurl             `json:"curl,omitempty"`
	Image                 image.Image                 `json:"image,omitempty"`
	InstallImage          image.Image                 `json:"installImage,omitempty"`
	Metrics               MongoDBSpecMetrics          `json:"metrics,omitempty"`
	PersistentVolume      MongoDBSpecPersistentVolume `json:"persistentVolume,omitempty"`
	Replicas              int                         `json:"replicas,omitempty"`
	TLS                   MongoDBSpecTLS              `json:"tls,omitempty"`
	ImagePullSecrets      []string                    `json:"imagePullSecrets,omitempty"`
	LivenessProbe         MongoDBSpecProbe            `json:"livenessProbe,omitempty"`
	ReadinessProbe        MongoDBSpecProbe            `json:"readinessProbe,omitempty"`
	Resources             MongoDBSpecResources        `json:"resources,omitempty"`
	WiredTigerCacheSizeGb float32                     `json:"wiredTigerCacheSizeGb,omitempty"`

	PriorityClassNameEnabled  bool `json:"priorityClassNameEnabled"`
	ServiceAccountNameEnabled bool `json:"serviceAccountNameEnabled"`
	NodeSelectorEnabled       bool `json:"nodeSelectorEnabled"`
	ClusterRoleEnabled        bool `json:"clusterRoleEnabled"`
}

// MongoDBSpecResources defines the configuration of Resources in spec
// +k8s:openapi-gen=true
type MongoDBSpecResources struct {
	Limits   MongoDBSpecResourcesLimit   `json:"limits,omitempty"`
	Requests MongoDBSpecResourcesRequest `json:"requests,omitempty"`
}

// MongoDBSpecResourcesLimit defines the configuration of ResourcesLimit in spec
// +k8s:openapi-gen=true
type MongoDBSpecResourcesLimit struct {
	Memory string `json:"memory,omitempty"`
}

// MongoDBSpecResourcesRequest defines the configuration of ResourcesRequest in spec
// +k8s:openapi-gen=true
type MongoDBSpecResourcesRequest struct {
	Memory string `json:"memory,omitempty"`
}

// MongoDBSpecProbe defines the configuration of LivenessProbe and ReadinessProbe in spec
// +k8s:openapi-gen=true
type MongoDBSpecProbe struct {
	FailureThreshold    int `json:"failureThreshold,omitempty"`
	InitialDelaySeconds int `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int `json:"periodSeconds,omitempty"`
	SuccessThreshold    int `json:"successThreshold,omitempty"`
	TimeoutSeconds      int `json:"timeoutSeconds,omitempty"`
}

// MongoDBSpecTLS defines the configuration of TLS in spec
// +k8s:openapi-gen=true
type MongoDBSpecTLS struct {
	Enabled    bool   `json:"enabled"`
	CASecret   string `json:"casecret,omitempty"`
	Issuer     string `json:"issuer,omitempty"`
	IssuerKind string `json:"issuerKind,omitempty"`
}

// MongoDBSpecPersistentVolume defines the configuration of Metrics in spec
// +k8s:openapi-gen=true
type MongoDBSpecPersistentVolume struct {
	Enabled      bool     `json:"enabled"`
	AccessModes  []string `json:"accessModes,omitempty"`
	Size         string   `json:"size,omitempty"`
	StorageClass string   `json:"storageClass,omitempty"`
}

// MongoDBSpecMetrics defines the configuration of Metrics in spec
// +k8s:openapi-gen=true
type MongoDBSpecMetrics struct {
	Enabled bool        `json:"enabled"`
	Image   image.Image `json:"image,omitempty"`
}

// MongoDBSpecAuth defines the configuration of Auth in spec
// +k8s:openapi-gen=true
type MongoDBSpecAuth struct {
	Enabled bool `json:"enabled"`
}

// MongoDBSpecCurl defines the configuration of Curl in spec
// +k8s:openapi-gen=true
type MongoDBSpecCurl struct {
	Image image.Image `json:"image,omitempty"`
}

// MongoDBStatus defines the observed state of MongoDB
// +k8s:openapi-gen=true
type MongoDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongoDB is the Schema for the mongodbs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type MongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBSpec   `json:"spec,omitempty"`
	Status MongoDBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongoDBList contains a list of MongoDB
type MongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDB{}, &MongoDBList{})
}
