package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// KlusterletServiceSpec defines the desired state of KlusterletService
// +k8s:openapi-gen=true
type KlusterletServiceSpec struct {
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	ConnectionManager ConnectionManager `json:"connection-manager"`
	CertificateIssuer CertificateIssuer `json:"certificate-issuer"`
	Search            Search            `json:"search"`
	WorkManager       WorkManager       `json:"work-manager"`
	PolicyController  PolicyController  `json:"policy-controller"`
	ServiceRegistry   ServiceRegistry   `json:"service-registry"`
	TopologyCollector TopologyCollector `json:"topology-collector"`
}

// ConnectionManager defines the desired state of the ConnectionManager component
// +k8s:openapi-gen=true
type ConnectionManager struct {
}

// CertificateIssuer defines the desired state of the CertificateIssuer component
// +k8s:openapi-gen=true
type CertificateIssuer struct {
}

// Search defines the desired state of the Search component
// +k8s:openapi-gen=true
type Search struct {
}

// WorkManager defines the desired state of the WorkManager component
// +k8s:openapi-gen=true
type WorkManager struct {
}

// PolicyController defines the desired state of the PolicyController component
// +k8s:openapi-gen=true
type PolicyController struct {
}

// ServiceRegistry defines the desired state of the ServiceRegistry component
// +k8s:openapi-gen=true
type ServiceRegistry struct {
}

// TopologyCollector defines the desired state of the TopologyCollector component
// +k8s:openapi-gen=true
type TopologyCollector struct {
}

// KlusterletServiceStatus defines the observed state of KlusterletService
// +k8s:openapi-gen=true
type KlusterletServiceStatus struct {
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KlusterletService is the Schema for the klusterletservices API
// +k8s:openapi-gen=true
type KlusterletService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KlusterletServiceSpec   `json:"spec,omitempty"`
	Status KlusterletServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KlusterletServiceList contains a list of KlusterletService
type KlusterletServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KlusterletService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KlusterletService{}, &KlusterletServiceList{})
}
