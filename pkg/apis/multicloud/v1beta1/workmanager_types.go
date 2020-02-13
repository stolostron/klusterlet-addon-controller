// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-cluster-management/endpoint-operator/pkg/image"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WorkManagerSpec defines the desired state of WorkManager
type WorkManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	ClusterLabels map[string]string `json:"clusterLabels"`

	// +kubebuilder:validation:MinLength=1
	ConnectionManager string `json:"connectionManager"`

	// +kubebuilder:validation:MinLength=1
	FullNameOverride string `json:"fullnameOverride"`

	PrometheusIntegration WorkManagerPrometheusIntegration `json:"prometheusIntegration"`

	Service WorkManagerService `json:"service"`
	Ingress WorkManagerIngress `json:"ingress"`

	WorkManagerConfig WorkManagerConfig `json:"workManager,omitempty"`

	ImagePullSecret string `json:"imagePullSecret,omitempty"`
}

// WorkManagerConfig defines work-manager container configuration parameters
type WorkManagerConfig struct {
	Enabled bool        `json:"enabled"`
	Image   image.Image `json:"image"`
}

// WorkManagerPrometheusIntegration defines prometheus integration parameters
type WorkManagerPrometheusIntegration struct {
	Enabled        bool   `json:"enabled"`
	Service        string `json:"service"`
	Secret         string `json:"secret"`
	ScrapeTarget   string `json:"scrapeTarget"`
	UseBearerToken bool   `json:"useBearerToken"`
}

// WorkManagerService defines service integration parameters
type WorkManagerService struct {
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	ServiceType string `json:"type"`
}

// WorkManagerIngress defines ingress configuration for WorkManager
type WorkManagerIngress struct {
	// +kubebuilder:validation:Enum=Ingress;Route;None
	IngressType string `json:"type"`
	Host        string `json:"host"`
	Port        string `json:"port"`
}

// WorkManagerStatus defines the observed state of WorkManager
type WorkManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkManager is the Schema for the workmanagers API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=workmanagers,scope=Namespaced
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
