package helpers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var InstallConfigYaml = []byte(`
apiVersion: v1
baseDomain: aws-cluster
metadata:
  name: 'cluster'
baseDomain: test.redhat.com
networking:
  networkType: OpenShiftSDN
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineNetwork:
  - cidr: 192.168.124.0/24
  serviceNetwork:
  - 172.30.0.0/16
proxy:
  httpsProxy: https://username:password@proxy.example.com:123/
  httpProxy: https://username:password@proxy.example.com:123/
  noProxy: 123.example.com,10.88.0.0/16
platform:
  gcp:
    projectID: yzw-yzw
    region: us-east1
`)
var InstallConfigBareMetalYaml = []byte(`
apiVersion: v1
baseDomain: aws-cluster
metadata:
  name: 'cluster'
baseDomain: test.redhat.com
networking:
  networkType: OpenShiftSDN
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineNetwork:
  - cidr: 192.168.124.0/24
  serviceNetwork:
  - 172.30.0.0/16
proxy:
  httpsProxy: https://username:password@proxy.example.com:123/
  httpProxy: https://username:password@proxy.example.com:123/
  noProxy: 123.example.com,10.88.0.0/16
platform:
  baremetal:
    libvirtURI: qemu+ssh://root@192.168.124.1/system
`)
var InstallConfigNoProxyYaml = []byte(`
apiVersion: v1
baseDomain: aws-cluster
metadata:
  name: 'cluster'
platform:
  gcp:
    projectID: yzw-yzw
    region: us-east1
`)

func NewInstallConfigSecret(name, namespace string, installConfig []byte) *corev1.Secret {
	data := map[string][]byte{}
	if len(installConfig) != 0 {
		data["install-config.yaml"] = installConfig
	}
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
		Type: corev1.SecretTypeOpaque,
	}
}
