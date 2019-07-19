package v1alpha1

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TillerSpec defines the desired state of Tiller
// +k8s:openapi-gen=true
type TillerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	FullNameOverride string `json:"fullnameOverride"`
	CACertIssuer     string `json:"caCertIssuer"`

	DefaultAdminUser string `json:"tiller_default_admin_user"`
	IAMHost          string `json:"tiller_iam_host,omitempty"`
	IAMPort          string `json:"tiller_iam_port,omitempty"`
	HistoryMax       int    `json:"tiller_history_max,omitempty"`
	CipherSuites     string `json:"tiller_ciphersuites,omitempty"`
	HostNetwork      string `json:"tiller_host_network,omitempty"`
	RouterHTTPSPort  int    `json:"router_https_port,omitempty"`

	ServiceType     string `json:"tiller_service_type,omitempty"`
	ServiceNodePort int    `json:"tiller_service_nodeport,omitempty"`

	Image image.Image `json:"image,omitempty"`
}

// TillerStatus defines the observed state of Tiller
// +k8s:openapi-gen=true
type TillerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Tiller is the Schema for the tillers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Tiller struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TillerSpec   `json:"spec,omitempty"`
	Status TillerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TillerList contains a list of Tiller
type TillerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tiller `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tiller{}, &TillerList{})
}
