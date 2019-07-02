package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/apis/core"
)

// IMPORTANT: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// KlusterletServiceSpec defines the desired state of KlusterletService
// +k8s:openapi-gen=true
type KlusterletServiceSpec struct {
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`
	// +kubebuilder:validation:MinLength=1
	Namespace string `json:"namespace"`

	ConnectionManager ConnectionManager `json:"connectionManager"`
	CertificateIssuer CertificateIssuer `json:"certificateIssuer"`
	Search            Search            `json:"search,omitempty"`
	WorkManager       WorkManager       `json:"workManager,omitempty"`
	PolicyController  PolicyController  `json:"policyController,omitempty"`
	ServiceRegistry   ServiceRegistry   `json:"serviceRegistry,omitempty"`
	TopologyCollector TopologyCollector `json:"topologyCollector,omitempty"`
}

// Image defines the image to pull for a container
// +k8s:openapi-gen=true
type Image struct {
	// +kubebuilder:validation:MinLength=1
	Repository string `json:"repository"`
	// +kubebuilder:validation:MinLength=1
	Tag string `json:"tag"`
	// +kubebuilder:validation:Enum=Always,Never,IfNotPresent
	PullPolicy core.PullPolicy `json:"pullPolicy"`
}

// Hub defines an MCM hub for the ConnectionManager component
type Hub struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// +kubebuilder:validation:MinLength=1
	Secret core.LocalObjectReference `json:"secret"`
}

// ConnectionManager defines the desired state of the ConnectionManager component
// +k8s:openapi-gen=true
type ConnectionManager struct {
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`
	LogLevel    int    `json:"logLevel"`
	Replicas    int    `json:"replicas"`

	// +kubebuilder:validation:MinLength=1
	Hubs []Hub `json:"hubs"`

	Affinity        core.Affinity             `json:"affinity"`
	Image           Image                     `json:"image"`
	ImagePullSecret core.LocalObjectReference `json:"imagePullSecret"`
	NodeSelector    core.NodeSelector         `json:"nodeSelector"`
	Tolerations     []core.Toleration         `json:"tolerations"`
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
