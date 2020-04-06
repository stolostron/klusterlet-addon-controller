#!/bin/bash

set -e

CURR_FOLDER_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
KIND_KUBECONFIG="${CURR_FOLDER_PATH}/../kind_kubeconfig.yaml"
export KUBECONFIG=${KIND_KUBECONFIG}
export DOCKER_IMAGE_AND_TAG=${1}
export PULL_SECRET=multicloud-image-pull-secret


if ! which kubectl > /dev/null; then
    echo "installing kubectl"
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
fi
if ! which kind > /dev/null; then
    echo "installing kind"
    curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.7.0/kind-$(uname)-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
fi
if ! which ginkgo > /dev/null; then
    echo "Installing ginkgo ..."
    GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
    GO111MODULE=off go get github.com/onsi/gomega/...
fi

echo "creating cluster"
kind create cluster --name endpoint-operator-test || exit 1

# setup kubeconfig
kind get kubeconfig --name endpoint-operator-test > ${KIND_KUBECONFIG}

echo "installing endpoint-operator"

#Create the namespace
kubectl create ns multicluster-endpoint

kubectl create secret docker-registry ${PULL_SECRET} \
      --docker-server=quay.io/open-cluster-management \
      --docker-username=${DOCKER_USER} \
      --docker-password=${DOCKER_PASS} \
      -n multicluster-endpoint

cat <<EOF > $PROJECT_DIR/overlays/template/kustomization.yaml
bases:
- ../../deploy

patchesStrategicMerge:
- |-
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: endpoint-operator
  spec:
    template:
      spec:
        imagePullSecrets:
        - name: $PULL_SECRET
EOF

kubectl apply -k deploy

# patch image
echo "patch image"
kubectl patch deployment endpoint-operator -n multicluster-endpoint -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"endpoint-operator\",\"image\":\"${DOCKER_IMAGE_AND_TAG}\"}]}}}}"
kubectl rollout status -n multicluster-endpoint deployment endpoint-operator --timeout=120s
sleep 10

ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/endpoint-operator-test/... -- --v=1

kind delete cluster --name endpoint-operator-test