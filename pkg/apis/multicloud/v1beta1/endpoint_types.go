// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EndpointSpec defines the desired state of Endpoint
type EndpointSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// +kubebuilder:validation:MinLength=1
	Version string `json:"version"`

	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// +kubebuilder:validation:MinLength=1
	ClusterNamespace string `json:"clusterNamespace"`

	ClusterLabels map[string]string `json:"clusterLabels"`

	BootStrapConfig map[string]string `json:"bootstrapConfig,omitempty"`

	PrometheusIntegration      EndpointPrometheusIntegrationSpec `json:"prometheusIntegration"`
	TopologyCollectorConfig    EndpointTopologyCollectorSpec     `json:"topologyCollector"`
	SearchCollectorConfig      EndpointSearchCollectorSpec       `json:"searchCollector"`
	PolicyController           EndpointPolicyControllerSpec      `json:"policyController"`
	ServiceRegistryConfig      EndpointServiceRegistrySpec       `json:"serviceRegistry"`
	ApplicationManagerConfig   EndpointApplicationManagerSpec    `json:"applicationManager"`
	ConnectionManagerConfig    EndpointConnectionManagerSpec     `json:"connectionManager"`
	CertPolicyControllerConfig EndpointCertPolicyControllerSpec  `json:"certPolicyController"`
	CISControllerConfig        EndpointCISControllerSpec         `json:"cisController"`
	IAMPolicyControllerConfig  EndpointIAMPolicyControllerSpec   `json:"iamPolicyController"`

	// +kubebuilder:validation:MinLength=1
	ImageRegistry    string `json:"imageRegistry"`
	ImageNamePostfix string `json:"imageNamePostfix,omitempty"`
	// +kubebuilder:validation:MinLength=1
	ImagePullSecret string `json:"imagePullSecret,omitempty"`

	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ComponentsImagesTag contains the tag of each component images
	ComponentsImagesTag map[string]string `json:"componentsImagesTag,omitempty"`

	// // ComponentTagMap contains the tag of each component
	// ComponentTagMap map[string]string `json:"componentTagMap"`
	// // ComponentImageMap contains the image name of each component
	// ComponentImageMap map[string]string `json:"componentImageMap"`

	// used for dev work only
	ComponentOperatorImage string `json:"componentOperatorImage,omitempty"`
}

// EndpointConnectionManagerSpec defines configuration for the ConnectionManager component
type EndpointConnectionManagerSpec struct {
}

// EndpointApplicationManagerSpec defines configuration for the ApplicationManager component
type EndpointApplicationManagerSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointSearchCollectorSpec defines configuration for the SearchCollector component
type EndpointSearchCollectorSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointCertPolicyControllertSpec defines configuration for the CertPolicyController component
type EndpointCertPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointIAMPolicyControllertSpec defines configuration for the IAMPolicyController component
type EndpointIAMPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointServiceRegistrySpec defines configuration for the ServiceRegistry component
type EndpointServiceRegistrySpec struct {
	Enabled                            bool   `json:"enabled"`
	DNSSuffix                          string `json:"dnsSuffix,omitempty"`
	Plugins                            string `json:"plugins,omitempty"`
	IstioIngressGateway                string `json:"istioIngressGateway,omitempty"`
	IstioserviceEntryRegistryNamespace string `json:"istioserviceEntryRegistryNamespace,omitempty"`
}

// EndpointWorkManagerSpec defines configuration for the WorkManager component
type EndpointWorkManagerSpec struct {
	ClusterLabels map[string]string `json:"clusterLabels"`
}

// EndpointPolicyControllerSpec defines configuration for the PolicyController component
type EndpointPolicyControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointTopologyCollectorSpec defines configuration for the WorkManager Promtheus Integration
type EndpointTopologyCollectorSpec struct {
	Enabled                 bool  `json:"enabled"`
	CollectorUpdateInterval int32 `json:"updateInterval"`
}

// EndpointPrometheusIntegrationSpec defines configuration for the Promtheus Integration
type EndpointPrometheusIntegrationSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointCISControllerSpec defines configuration for the CISController component
type EndpointCISControllerSpec struct {
	Enabled bool `json:"enabled"`
}

// EndpointStatus defines the observed state of Endpoint
type EndpointStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Endpoint is the Schema for the endpoints API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=endpoints,scope=Namespaced
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointSpec   `json:"spec,omitempty"`
	Status EndpointStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
