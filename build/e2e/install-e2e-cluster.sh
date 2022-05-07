#!/bin/bash

###############################################################################
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

set -x
set -eo pipefail

echo_cluster_registered_ok() {
  kubectl get managedcluster cluster1 -oyaml
  echo "-------------------------------------------------------------------------"
  echo "Cluster cluster1 registered to the hub successfully."
  echo "-------------------------------------------------------------------------"
}

if [[ -f "./.kubeconfig" ]]; then
    echo "./.kubeconfig exists."
    echo $KUBECONFIG
    KUBECONFIG=$(KUBECONFIG:-"./.kubeconfig")
    echo $KUBECONFIG
fi
echo "kubeconfig: $KUBECONFIG"

echo "############  Cloning registration-operator"
rm -rf registration-operator

git clone --depth 1 --branch release-2.4 https://github.com/stolostron/registration-operator.git

cd registration-operator || {
  printf "cd failed, registration-operator does not exist"
  return 1
}

echo "############  Deploying ocm"
export IMAGE_NAME=quay.io/stolostron/registration-operator:release-2.4
export REGISTRATION_IMAGE=quay.io/stolostron/registration:release-2.4
export WORK_IMAGE=quay.io/stolostron/work:release-2.4
export PLACEMENT_IMAGE=quay.io/stolostron/placement:release-2.4

make deploy
if [ $? -ne 0 ]; then
 echo "############  Failed to deploy"
 exit 1
fi


# approve cluster join request and csr
n=0
until [ "$n" -ge 30 ]
do
    kubectl patch managedcluster cluster1 -p='{"spec":{"hubAcceptsClient":true}}' --type=merge && break
    n=$((n+1)) 
    sleep 10
done

clusterCondition=$(kubectl get managedcluster cluster1 -o jsonpath='{.status.conditions}')
if [[ "$clusterCondition" == *\"reason\":\"ManagedClusterAvailable\",\"status\":\"True\"* ]]; then
    echo "managedcluster cluster1 already be available."
    echo_cluster_registered_ok
    exit 0
fi

n=0
until [ "$n" -ge 30 ]
do
    clusterCSR=$(kubectl get csr -l open-cluster-management.io/cluster-name=cluster1 | grep -v NAME | awk '{print $1}')
    if [ -n "$clusterCSR" ]; then
      certificate=$(kubectl get csr "$clusterCSR" -o jsonpath='{.status.certificate}')
      if [ -z "$certificate" ]; then
        kubectl certificate approve "$clusterCSR"
      else
        break
      fi
    fi
    n=$((n+1)) 
    sleep 5
done

echo_cluster_registered_ok

cd ../ || exist
rm -rf registration-operator
