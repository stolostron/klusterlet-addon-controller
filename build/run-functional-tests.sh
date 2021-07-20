#!/bin/bash

###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

set -e
#set -xs

CURR_FOLDER_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
KIND_KUBECONFIG="${CURR_FOLDER_PATH}/../kind_kubeconfig.yaml"

export KUBECONFIG=${KIND_KUBECONFIG}
export DOCKER_IMAGE=${1}

export FUNCT_TEST_TMPDIR="${CURR_FOLDER_PATH}/../test/functional/tmp"
export FUNCT_TEST_COVERAGE="${CURR_FOLDER_PATH}/../test/functional/coverage"

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
    curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.11.1/kind-$(uname)-amd64
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

echo "setting up test tmp folder"
[ -d "$FUNCT_TEST_TMPDIR" ] && rm -r "$FUNCT_TEST_TMPDIR"
mkdir -p "$FUNCT_TEST_TMPDIR"
mkdir -p "$FUNCT_TEST_TMPDIR/kind-config"

echo "setting up test coverage folder"
[ -d "$FUNCT_TEST_COVERAGE" ] && rm -r "$FUNCT_TEST_COVERAGE"
mkdir -p "${FUNCT_TEST_COVERAGE}"

echo "generating kind configfile"
cat << EOF > "${FUNCT_TEST_TMPDIR}/kind-config/kind-config.yaml"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    listenAddress: "0.0.0.0"
  - containerPort: 443
    hostPort: 443
    listenAddress: "0.0.0.0"
  - containerPort: 6443
    hostPort: 32800
    listenAddress: "0.0.0.0"
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration #for worker use JoinConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        system-reserved: memory=2Gi
  extraMounts:
  - hostPath: "${FUNCT_TEST_COVERAGE}"
    containerPath: /tmp/coverage
networking:
  apiServerPort: 6443
EOF


echo "creating cluster"

kind create cluster --name klusterlet-addon-controller-test  --config=${FUNCT_TEST_TMPDIR}/kind-config/kind-config.yaml --image kindest/node:v1.20.2 || exit 1

# setup kubeconfig
kind export kubeconfig --name=klusterlet-addon-controller-test --kubeconfig ${KIND_KUBECONFIG} 

for i in {1..7}; do
  echo "############$i  Checking kind pods status"

  RUNNING_POD=$(kubectl get pods --all-namespaces 2>&1 | grep '1/1' | grep -c 'Running')
  if [ "${RUNNING_POD}" -eq 9 ]; then
    break
  fi

  if [ $i -eq 6 ]; then
    echo "!!!!!!!!!!  the kind is not ready within 3 minutes"
   kubectl get pods --all-namespaces
    exit 1
  fi
  sleep 30
done

echo "##kind cluster is ready##"

kubectl get pods --all-namespaces

#Apply all dependent crds
echo "installing crds"
for file in `ls deploy/dev-crds/*.crd.yaml`; do kubectl apply -f $file; done
sleep 5

#Apply all dependent crs
echo "installing crs"
for file in `ls deploy/dev-crs/*.cr.yaml`; do kubectl apply -f $file; done

echo "installing klusterlet-addon-controller"
kind load docker-image $DOCKER_IMAGE --name=klusterlet-addon-controller-test 

#Create the namespace
kubectl apply -f ${PROJECT_DIR}/deploy/namespace.yaml

COMPONENT_DOCKER_REPO=`echo "$DOCKER_IMAGE" | cut -f1 -d:`

#Loop on scenario
for dir in overlays/test/* ; do
  echo "=========================================="
  echo "Executing test "$dir
  echo "=========================================="
  echo $DOCKER_IMAGE

  kubectl apply -k "$dir" --dry-run=true -o yaml | sed "s|REPLACE_NAME|${DOCKER_IMAGE}|g" | kubectl apply -f -

  # kubectl apply -k $dir
  # kubectl patch deployment klusterlet-addon-controller -n open-cluster-management -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"klusterlet-addon-controller\",\"image\":\"${DOCKER_IMAGE}\"}]}}}}"
  kubectl rollout status -n open-cluster-management deployment klusterlet-addon-controller --timeout=120s

  sleep 10

  ginkgo -v -tags functional -failFast --slowSpecThreshold=10 test/functional/... -- --v=1 --image-registry=$COMPONENT_DOCKER_REPO

  POD=$(kubectl get pod -n open-cluster-management  -o jsonpath="{.items[0].metadata.name}")
  kubectl exec -it $POD -n open-cluster-management -- ls /tmp/coverage

  kubectl delete deployment klusterlet-addon-controller -n open-cluster-management
  sleep 10
done

sleep 20
echo "delete kind cluster"
kind delete cluster --name klusterlet-addon-controller-test 

if [ `find $FUNCT_TEST_COVERAGE -prune -empty 2>/dev/null` ]; then
  echo "no coverage files found. skipping"
else
  echo "merging coverage files"

  gocovmerge "${FUNCT_TEST_COVERAGE}/"* >> "${FUNCT_TEST_COVERAGE}/cover-functional.out"
  COVERAGE=$(go tool cover -func="${FUNCT_TEST_COVERAGE}/cover-functional.out" | grep "total:" | awk '{ print $3 }' | sed 's/[][()><%]/ /g')
  echo "-------------------------------------------------------------------------"
  echo "TOTAL COVERAGE IS ${COVERAGE}%"
  echo "-------------------------------------------------------------------------"

  go tool cover -html "${FUNCT_TEST_COVERAGE}/cover-functional.out" -o ${PROJECT_DIR}/test/functional/coverage/cover-functional.html

  cat ${FUNCT_TEST_COVERAGE}/cover-functional.out | grep -v "zz_generated.deepcopy.go" > ${FUNCT_TEST_COVERAGE}/cover-functional-filtered.out
  COVERAGE_FILTERED=$(go tool cover -func="${FUNCT_TEST_COVERAGE}/cover-functional-filtered.out" | grep "total:" | awk '{ print $3 }' | sed 's/[][()><%]/ /g')
  echo "-------------------------------------------------------------------------"
  echo "TOTAL FILTERED (ie: exclude zz_generated.deepcopy.go) COVERAGE IS ${COVERAGE_FILTERED}%"
  echo "-------------------------------------------------------------------------"

  go tool cover -html "${FUNCT_TEST_COVERAGE}/cover-functional-filtered.out" -o ${PROJECT_DIR}/test/functional/coverage/cover-functional-filtered.out
fi
