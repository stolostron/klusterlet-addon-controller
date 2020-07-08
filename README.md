# klusterlet addon controller

A Go operator built with the [operator-sdk](https://github.com/operator-framework/operator-sdk) that is used to manage the Create Update Delete of the component CR in the Klusterlet Component Operator.

## Prerequisites

- Must have [operator-sdk](https://github.com/operator-framework/operator-sdk) v0.18.1 installed

```shell
# can be installed with the following command
> make deps
```

## Prepare your cluster 

1. Import a managed cluster. Follow this guidelines to import cluster manually.

- [manual-import](https://github.com/open-cluster-management/rcm-controller/blob/master/docs/managedcluster_manual_import.md)

2. Install klusterlet CRD

```shell
make utils-crds-install
```

## Running Klusterlet addon controller locally for development

1. Run Klusterlet Addon Controller locally

```shell
make run
```

## Running Klusterlet addon controller in-cluster for deployment

1. Apply the `deploy/deploy.yaml` to create the ServiceAccount, ClusterRole, ClusterRoleBinding and Deployment for the operator

```shell
kubectl apply -f deploy/deploy.yaml
```

NOTE: this will use the amd64 version of the operator

## Installing klusterlet addons using Klusterlet addon controller

To create a klusterlet addon operator deployment with the klusterlet addon controller you need to create the klusterlet CR

Example of KlusterletAddonConfig CR <https://github.com/open-cluster-management/endpoint-operator/blob/master/deploy/crds/agent.open-cluster-management.io_v1_klusterletaddonconfig_cr.yaml>

## Rebuilding zz_generated.deepcopy.go file
Any modifications to files pkg/apis/agent/v1/*types.go will require you to run the
following:
```
operator-sdk generate k8s
```
to regenerate the zz_generated.deepcopy.go file.

## Build and publish a personal build to scratch artifactory

- `export GITHUB_USER=<GITHUB_USER>`
- `export GITHUB_TOKEN=<GITHUB_TOKEN>`
- `make init`
- `make build`
- `make docker:tag`
- `make docker:push`

## Run Functional Test

### Before Testing functional test with KinD

1. Make sure you have [ginkgo](https://onsi.github.io/ginkgo/) excutable ready in your env. If not, do the following:
   ```
    go get github.com/onsi/ginkgo/ginkgo
    go get github.com/onsi/gomega/...
   ```

2. Run functional test locally with KinD, you will need to install Kind https://kind.sigs.k8s.io/docs/user/quick-start/#installation

### Run Functional Test Locally with KinD

1. Export the image postfix for rcm-controller image:
   ```
    export COMPONENT_TAG_EXTENSION=-SNAPSHOT-2020-04-01-20-49-00
   ```
2. Run tests:
   - Run the following command to build the image, setup & start a kind cluster (Ideal for someone new to the repo and wanting to test changes):
    ```
    export DOCKER_USER=<Docker username>
    export DOCKER_PASS=<Docker password>
    make functional-test-full
   ```
   - Run the following command to setup & start a kind cluster:
   ```
    make component/test/functional
   ```
   - Run the following command to run the test on an existing cluster:
    ```
    export KUBECONFIG=...
    make functional-test
   ```
