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

// MeteringSpec defines the desired state of Metering
// +k8s:openapi-gen=true
type MeteringSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string                  `json:"fullnameOverride,omitempty"`
	API              MeteringSpecAPI         `json:"api,omitempty"`
	UI               MeteringSpecUI          `json:"ui,omitempty"`
	McmUI            MeteringSpecMcmUI       `json:"mcmui,omitempty"`
	DataManager      MeteringSpecDataManager `json:"dm,omitempty"`
	Reader           MeteringSpecReader      `json:"reader,omitempty"`
	Sender           MeteringSpecSender      `json:"sender,omitempty"`

	ImagePullSecrets []string          `json:"imagePullSecrets,omitempty"`
	Mongo            MeteringSpecMongo `json:"mongo,omitempty"`

	ServiceAccountNameEnabled bool `json:"serviceAccountNameEnabled"`
	ClusterRoleEnabled        bool `json:"clusterRoleEnabled"`
	NodeSelectorEnabled       bool `json:"nodeSelectorEnabled"`
}

// MeteringSpecSender defines the Sender configuration in the the metering spec
type MeteringSpecSender struct {
	Enabled                  bool        `json:"enabled"`
	ClusterName              string      `json:"clusterName,omitempty"`
	ClusterNamespace         string      `json:"clusterNamespace,omitempty"`
	HubKubeConfigSecret      string      `json:"hubKubeConfigSecret,omitempty"`
	Image                    image.Image `json:"image,omitempty"`
	NodeSelectorEnabled      bool        `json:"nodeSelectorEnabled"`
	PriorityClassNameEnabled bool        `json:"priorityClassNameEnabled"`
}

// MeteringSpecAPI defines the API configuration in the the metering spec
type MeteringSpecAPI struct {
	Enabled bool `json:"enabled"`
}

// MeteringSpecUI defines the UI configuration in the the metering spec
type MeteringSpecUI struct {
	Enabled bool `json:"enabled"`
}

// MeteringSpecMcmUI defines the MCMUI configuration in the the metering spec
type MeteringSpecMcmUI struct {
	Enabled bool `json:"enabled"`
}

// MeteringSpecDataManager defines the DataManager configuration in the the metering spec
type MeteringSpecDataManager struct {
	Enabled                  bool        `json:"enabled"`
	Image                    image.Image `json:"image,omitempty"`
	NodeSelectorEnabled      bool        `json:"nodeSelectorEnabled,"`
	PriorityClassNameEnabled bool        `json:"priorityClassNameEnabled"`
}

// MeteringSpecReader defines the Reader configuration in the the metering spec
type MeteringSpecReader struct {
	Enabled       bool        `json:"enabled"`
	Image         image.Image `json:"image,omitempty"`
	ClusterIssuer string      `json:"clusterIssuer,omitempty"`
}

// MeteringSpecMongo defines the mongo configuration in the the metering spec
type MeteringSpecMongo struct {
	ClusterCertsSecret string                    `json:"clustercertssecret,omitempty"`
	ClientCertsSecret  string                    `json:"clientcertssecret,omitempty"`
	Username           MeteringSpecMongoUsername `json:"username,omitempty"`
	Password           MeteringSpecMongoPassword `json:"password,omitempty"`
}

// MeteringSpecMongoUsername defines the mongo username configuration in the the metering spec
type MeteringSpecMongoUsername struct {
	Secret string `json:"secret,omitempty"`
	Key    string `json:"key,omitempty"`
}

// MeteringSpecMongoPassword defines the mongo password configuration in the the metering spec
type MeteringSpecMongoPassword struct {
	Secret string `json:"secret,omitempty"`
	Key    string `json:"key,omitempty"`
}

// MeteringStatus defines the observed state of Metering
// +k8s:openapi-gen=true
type MeteringStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Metering is the Schema for the meterings API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Metering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MeteringSpec   `json:"spec,omitempty"`
	Status MeteringStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MeteringList contains a list of Metering
type MeteringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Metering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Metering{}, &MeteringList{})
}
