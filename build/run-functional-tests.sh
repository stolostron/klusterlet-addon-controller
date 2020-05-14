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
#set -x

DOCKER_IMAGE=$1-coverage
if [ -z $FUNCT_TEST_TMPDIR ]; then
 export FUNCT_TEST_TMPDIR=/tmp/`uuidgen`
fi

mkdir -p ${FUNCT_TEST_TMPDIR}

echo "FUNCT_TEST_TMPDIR="$FUNCT_TEST_TMPDIR

KIND_KUBECONFIG="${PROJECT_DIR}/kind_kubeconfig.yaml"
echo "KIND_KUBECONFIG="$KIND_KUBECONFIG

if [ -z $DOCKER_USER ]; then
   echo "DOCKER_USER is not defined!"
   exit 1
fi
if [ -z $DOCKER_PASS ]; then
   echo "DOCKER_PASS is not defined!"
   exit 1
fi


export KUBECONFIG=${KIND_KUBECONFIG}
export PULL_SECRET=multicloud-image-pull-secret

wait_file() {
  local file="$1"; shift
  local wait_seconds="${1:-10}"; shift # 10 seconds as default timeout

  until test $((wait_seconds--)) -eq 0 -o -f "$file" ; do sleep 1; done

  ((++wait_seconds))
}

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
kind --version

if ! which ginkgo > /dev/null; then
    echo "Installing ginkgo ..."
    GO111MODULE=off go get github.com/onsi/ginkgo/ginkgo
    GO111MODULE=off go get github.com/onsi/gomega/...
fi

if ! which gocovmerge > /dev/null; then
    echo "Installing gocovmerge...";
    GO111MODULE=off go get -u github.com/wadey/gocovmerge;
fi


echo "creating cluster"

sed "s#REPLACE_DIR#${FUNCT_TEST_TMPDIR}/test/coverage-functional/endpoint-operator#" ${PROJECT_DIR}/build/kind-config/kind-config.yaml > ${FUNCT_TEST_TMPDIR}/kind-config-generated.yaml

cat ${FUNCT_TEST_TMPDIR}/kind-config-generated.yaml

#Create local directory to hold coverage results
mkdir -p ${FUNCT_TEST_TMPDIR}/test/coverage-functional/endpoint-operator

kind create cluster --name klusterlet-operator-test --config=${FUNCT_TEST_TMPDIR}/kind-config-generated.yaml  || exit 1

# setup kubeconfig
kind export kubeconfig --name=klusterlet-operator-test --kubeconfig ${KIND_KUBECONFIG} 

echo "installing klusterlet-operator"

kind load docker-image $DOCKER_IMAGE --name=klusterlet-operator-test 

#Create the namespace
kubectl apply -f ${PROJECT_DIR}/deploy/namespace.yaml

kubectl create secret docker-registry ${PULL_SECRET} \
      --docker-server=quay.io/open-cluster-management \
      --docker-username=$DOCKER_USER \
      --docker-password=$DOCKER_PASS \
      -n klusterlet

#Loop on scenario
for dir in overlays/test/* ; do
  echo "=========================================="
  echo "Executing test "$dir
  echo "=========================================="
  kubectl apply -k $dir
  kubectl patch deployment klusterlet-operator -n klusterlet -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"klusterlet-operator\",\"image\":\"${DOCKER_IMAGE}\"}]}}}}"
  kubectl rollout status -n klusterlet deployment klusterlet-operator --timeout=120s
  sleep 10
  ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/klusterlet-operator-test/... -- --v=1
  kubectl delete deployment klusterlet-operator -n klusterlet
done

kind delete cluster --name klusterlet-operator-test 

rm -rf ${PROJECT_DIR}/test/coverage-functional
mkdir -p ${PROJECT_DIR}/test/coverage-functional

mv ${FUNCT_TEST_TMPDIR}/test/coverage-functional/endpoint-operator/* ${PROJECT_DIR}/test/coverage-functional/

gocovmerge ${PROJECT_DIR}/test/coverage-functional/* >> ${PROJECT_DIR}/test/coverage-functional/cover-functional.out
COVERAGE=$(go tool cover -func=${PROJECT_DIR}/test/coverage-functional/cover-functional.out | grep "total:" | awk '{ print $3 }' | sed 's/[][()><%]/ /g')
echo "-------------------------------------------------------------------------"
echo "TOTAL COVERAGE IS ${COVERAGE}%"
echo "-------------------------------------------------------------------------"

go tool cover -html ${PROJECT_DIR}/test/coverage-functional/cover-functional.out -o ${PROJECT_DIR}/test/coverage-functional/cover-functional.html
