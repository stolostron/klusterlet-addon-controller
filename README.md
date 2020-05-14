# klusterlet operator

A Go operator built with the [operator-sdk](https://github.com/operator-framework/operator-sdk) that is used to manage the Create Update Delete of the component CR in the Klusterlet Component Operator.

## Prerequisites

- Must have [operator-sdk](https://github.com/operator-framework/operator-sdk) v0.15.1 installed

```shell
# can be installed with the following command
> make deps


## Prepare your cluster 

1. Create namespace

```shell
kubectl create namespace klusterlet
```

2. Create image pull secret for artifactory

- https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
- please name the image pull secret `multicluster-endpoint-operator-pull-secret` the instruction after will refer to it

3. Create klusterlet-bootstrap secret (use to register to hub)

- log into your hub cluster
- click the *user icon* (upper right corner)
- click *Configure client*
- click the "Copy to clipboard" button
- `export tmpKUBECONFIG=$(mktemp /tmp/kubeconfigXXXX)`
- `export KUBECONFIG=$tmpKUBECONFIG`
- paste from clipboard
- `unset KUBECONFIG`
- log into the cluster you want to install klusterlet on
- `kubectl create secret generic klusterlet-bootstrap -n klusterlet --from-file=kubeconfig=$tmpKUBECONFIG`

4. Install klusterlet CRD

```shell
make utils:crds:install
```

## Running Klusterlet Operator locally for development

1. Run Klusterlet Operator locally

```shell
make operator:run
```

## Running Klusterlet Operator in-cluster for deployment

1. Apply the `deploy/deploy.yaml` to create the ServiceAccount, ClusterRole, ClusterRoleBinding and Deployment for the operator

```shell
kubectl apply -f deploy/deploy.yaml
```

NOTE: this will use the amd64 version of the operator

## Installing Klusterlet using Klusterlet Operator

To create a klusterlet deployment with the klusterlet operator you need to create the klusterlet CR

Example of Klusterlet CR `/deploy/crds/agent.open-cluster-management.io_v1beta1_klusterlet_cr.yaml`

## Rebuilding zz_generated.deepcopy.go file
Any modifications to files pkg/apis/agent/v1beta1/*types.go will require you to run the
following:
```
operator-sdk generate k8s
```
to regenerate the zz_generated.deepcopy.go file.

## Build and publish a personal build to scratch artifactory

- `export GITHUB_USER=<GITHUB_USER>`
- `export GITHUB_TOKEN=<GITHUB_TOKEN>`
- `make init`
- `make operator:build`
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
