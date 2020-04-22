#!/bin/bash
###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Licensed Materials - Property of IBM
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

set -e
# set -x

DOCKER_IMAGE=$1

CURR_FOLDER_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
KIND_KUBECONFIG="${CURR_FOLDER_PATH}/../kind_kubeconfig.yaml"
export KUBECONFIG=${KIND_KUBECONFIG}
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
kind export kubeconfig --name=endpoint-operator-test --kubeconfig ${KIND_KUBECONFIG}

echo "installing endpoint-operator"

kind load docker-image $DOCKER_IMAGE --name=endpoint-operator-test

#Create the namespace
kubectl apply -f ${PROJECT_DIR}/deploy/namespace.yaml

kubectl create secret docker-registry ${PULL_SECRET} \
      --docker-server=quay.io/open-cluster-management \
      --docker-username=$DOCKER_USER \
      --docker-password=$DOCKER_PASS \
      -n multicluster-endpoint

for dir in overlays/test/* ; do
  echo "Executing test "$dir
  kubectl apply -k $dir
  kubectl patch deployment endpoint-operator -n multicluster-endpoint -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"endpoint-operator\",\"image\":\"${DOCKER_IMAGE}\"}]}}}}"
  kubectl rollout status -n multicluster-endpoint deployment endpoint-operator --timeout=120s
  sleep 10
  ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/endpoint-operator-test/... -- --v=1
done

kind delete cluster --name endpoint-operator-test