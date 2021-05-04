module github.com/open-cluster-management/klusterlet-addon-controller

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.1
	github.com/open-cluster-management/api v0.0.0-20210409125704-06f2aec1a73f
	github.com/open-cluster-management/library-e2e-go v0.0.0-20200620112055-c80fc3c14997
	github.com/open-cluster-management/library-go v0.0.0-20200828173847-299c21e6c3fc
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/build-machinery-go v0.0.0-20210115170933-e575b44a7a94
	github.com/sclevine/agouti v3.0.0+incompatible
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.14.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.0
	k8s.io/apiextensions-apiserver v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.20.0
	sigs.k8s.io/controller-runtime v0.6.2
)

replace (
	github.com/coreos/etcd => go.etcd.io/etcd v3.3.22+incompatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.2.1
	github.com/go-logr/zapr => github.com/go-logr/zapr v0.2.0
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	github.com/mattn/go-sqlite3 => github.com/mattn/go-sqlite3 v1.10.0
	k8s.io/client-go => k8s.io/client-go v0.20.0
)
