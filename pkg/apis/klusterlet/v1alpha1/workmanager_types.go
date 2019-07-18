package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkManagerSpec defines the desired state of WorkManager
// +k8s:openapi-gen=true
type WorkManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	ClusterName      string            `json:"clusterName"`
	ClusterNamespace string            `json:"clusterNamespace"`
	ClusterLabels    map[string]string `json:"clusterLabels"`

	ConnectionManager string `json:"connectionManager"`
	Tiller            string `json:"tiller"`

	FullNameOverride string `json:"fullnameOverride"`

	TillerIntegration     WorkManagerTillerIntegration     `json:"tillerIntegration"`
	PrometheusIntegration WorkManagerPrometheusIntegration `json:"prometheusIntegration"`

	Service WorkManagerService `json:"service"`
	Ingress WorkManagerIngress `json:"ingress"`
}

// WorkManagerTillerIntegration defines tiller integration paramaters
// +k8s:openapi-gen=true
type WorkManagerTillerIntegration struct {
	Enabled           bool   `json:"enabled"`
	Endpoint          string `json:"endpoint"`
	CertIssuer        string `json:"certIssuer"`
	HelmReleasePrefix string `json:"helmReleasePrefix"`
	AutoGenSecret     bool   `json:"autoGenSecret"`
	User              string `json:"user"`
}

// WorkManagerPrometheusIntegration defines promethues integration paramaters
// +k8s:openapi-gen=true
type WorkManagerPrometheusIntegration struct {
	Enabled        bool   `json:"enabled"`
	Service        string `json:"service"`
	Secret         string `json:"secret"`
	UseBearerToken bool   `json:"useBearerToken"`
}

// WorkManagerService defines tiller integration paramaters
// +k8s:openapi-gen=true
type WorkManagerService struct {
	// +kubebuilder:validation:Enum=ClusterIP,NodePort,LoadBalancer
	ServiceType string `json:"type"`
}

// WorkManagerIngress defines ingress configuration for WorkManager
// +k8s:openapi-gen=true
type WorkManagerIngress struct {
	// +kubebuilder:validation:Enum=Ingress,Route,None
	IngressType string `json:"type"`
	Host        string `json:"host"`
	Port        string `json:"port"`
}

// WorkManagerStatus defines the observed state of WorkManager
// +k8s:openapi-gen=true
type WorkManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkManager is the Schema for the workmanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type WorkManager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkManagerSpec   `json:"spec,omitempty"`
	Status WorkManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkManagerList contains a list of WorkManager
type WorkManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkManager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkManager{}, &WorkManagerList{})
}
