## Intro 
Klusterlet operator is a Go operator build with the operator-sdk that's use to manage the Create Update Delete of the component CR in the Klusterlet Component Operator 

## Prepare your cluster 
1. Create namespace
```
kubectl create namespace multicluster-endpoint
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
- `kubectl create secret generic klusterlet-bootstrap -n multicluster-endpoint --from-file=kubeconfig=$tmpKUBECONFIG`

4. Install klusterlet CRD
```
make utils:crds:install
```

## Running Klusterlet Operator locally for development
1. Run Klusterlet Operator locally
```
make operator:run
```

## Running Klusterlet Operator in-cluster for deployment
1. Apply the `deploy/deploy.yaml` to create the ServiceAccount, ClusterRole, ClusterRoleBinding and Deployment for the operator
```
kubectl apply -f deploy/deploy.yaml
```
NOTE: this will use the amd64 version of the operator

## Installing Klusterlet using Klusterlet Operator 
To create a klusterlet deployment with the klusterlet operator u need to create the klusterlet CR

Example of Klusterlet CR `/deploy/crds/multicloud_v1beta1_endpoint_cr.yaml`

## Build and publish a personal build to scratch artifactory
- `export GITHUB_USER=<GITHUB_USER>`
- `export GITHUB_TOKEN=<GITHUB_TOKEN>`
- `make init`
- `make operator:build`
- `make docker:tag`
- `make docker:push`
