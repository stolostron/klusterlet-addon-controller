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

// MonitoringSpec defines the desired state of Monitoring
type MonitoringSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	FullNameOverride             string                                 `json:"fullnameOverride"`
	Enabled                      bool                                   `json:"enabled"`
	Mode                         string                                 `json:"mode"`
	ImagePullSecret              string                                 `json:"imagePullSecrets,omitempty"`
	MonitoringFullnameOverride   string                                 `json:"monitoringFullnameOverride"`
	PrometheusFullnameOverride   string                                 `json:"prometheusFullnameOverride"`
	GrafanaFullnameOverride      string                                 `json:"grafanaFullnameOverride"`
	TLS                          MonitoringTLS                          `json:"tls"`
	Prometheus                   MonitoringPrometheus                   `json:"prometheus"`
	Alertmanager                 MonitoringAlertmanager                 `json:"alertmanager"`
	Grafana                      MonitoringGrafana                      `json:"grafana"`
	KubeStateMetrics             MonitoringKubeStateMetrics             `json:"kubeStateMetrics"`
	NodeExporter                 MonitoringNodeExporter                 `json:"nodeExporter"`
	CollectdExporter             MonitoringCollectdExporter             `json:"collectdExporter"`
	ElasticsearchExporter        MonitoringElasticsearchExporter        `json:"elasticsearchExporter"`
	ConfigmapReload              MonitoringConfigmapReload              `json:"configmapReload"`
	Router                       MonitoringRouter                       `json:"router"`
	PrometheusOperatorController MonitoringPrometheusOperatorController `json:"prometheusOperatorController"`
	PrometheusOperator           MonitoringPrometheusOperator           `json:"prometheusOperator"`
	PrometheusConfigReloader     MonitoringPrometheusConfigReloader     `json:"prometheusConfigReloader"`
	Curl                         MonitoringCurl                         `json:"curl"`
}

// MonitoringTLS defines TLS configuration in the spec
type MonitoringTLS struct {
	Enabled bool            `json:"enabled"`
	Issuer  string          `json:"issuer"`
	CA      MonitoringTLSCA `json:"ca"`
}

// MonitoringTLSCA defines CA configuration in the TLS
type MonitoringTLSCA struct {
	SecretName string `json:"secretName"`
}

// MonitoringPrometheus defines Prometheus configuration in the Spec
type MonitoringPrometheus struct {
	Image      image.Image                    `json:"image"`
	EtcdTarget MonitoringPrometheusEtcdTarget `json:"etcdTarget"`
}

// MonitoringPrometheusEtcdTarget defines EtcdTarget configuration in the Prometheus
type MonitoringPrometheusEtcdTarget struct {
	Enabled bool `json:"enabled"`
}

// MonitoringAlertmanager defines Alertmanager configuration in the Spec
type MonitoringAlertmanager struct {
	Enabled bool `json:"enabled"`
}

// MonitoringGrafana defines Grafana configuration in the Spec
type MonitoringGrafana struct {
	Enabled bool `json:"enabled"`
}

// MonitoringKubeStateMetrics defines KubeStateMetrics configuration in the Spec
type MonitoringKubeStateMetrics struct {
	Enabled bool `json:"enabled"`
}

// MonitoringNodeExporter defines NodeExporter configuration in the Spec
type MonitoringNodeExporter struct {
	Enabled bool `json:"enabled"`
}

// MonitoringCollectdExporter defines CollectdExporter configuration in the Spec
type MonitoringCollectdExporter struct {
	Enabled bool `json:"enabled"`
}

// MonitoringElasticsearchExporter defines ElasticsearchExporter configuration in the Spec
type MonitoringElasticsearchExporter struct {
	Enabled bool `json:"enabled"`
}

// MonitoringConfigmapReload defines ConfigmapReload configuration in the Spec
type MonitoringConfigmapReload struct {
	Image image.Image `json:"image"`
}

// MonitoringRouter defines Router configuration in the Spec
type MonitoringRouter struct {
	Image image.Image `json:"image"`
}

// MonitoringAlertruleController defines AlertruleController configuration in the Spec
type MonitoringAlertruleController struct {
	Image image.Image `json:"image"`
}

// MonitoringPrometheusOperatorController defines PrometheusOperatorController configuration in the Spec
type MonitoringPrometheusOperatorController struct {
	Image image.Image `json:"image"`
}

// MonitoringPrometheusOperator defines Grafana configuration in the Spec
type MonitoringPrometheusOperator struct {
	Image image.Image `json:"image"`
}

// MonitoringPrometheusConfigReloader defines PrometheusConfigReloader configuration in the Spec
type MonitoringPrometheusConfigReloader struct {
	Image image.Image `json:"image"`
}

// MonitoringCurl defines Curl configuration in the Spec
type MonitoringCurl struct {
	Image image.Image `json:"image"`
}

// MonitoringStatus defines the observed state of Monitoring
type MonitoringStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Monitoring is the Schema for the monitorings API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=monitorings,scope=Namespaced
type Monitoring struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitoringSpec   `json:"spec,omitempty"`
	Status MonitoringStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MonitoringList contains a list of Monitoring
type MonitoringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitoring `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitoring{}, &MonitoringList{})
}
