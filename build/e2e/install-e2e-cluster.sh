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
CLUSTER_IP=${CLUSTER_IP:-$(kubectl get svc kubernetes -n default -o jsonpath="{.spec.clusterIP}")}
echo "clusterip: $CLUSTER_IP"
CLUSTER_CONTEXT=${CLUSTER_CONTEXT:-$(kubectl config current-context)}
echo "context: $CLUSTER_CONTEXT"
# prepare bootstrap-hub-kubeconfig secret
cp "${KUBECONFIG}" e2e-kubeconfig
kubectl config set clusters."${CLUSTER_CONTEXT}".server https://"${CLUSTER_IP}" --kubeconfig e2e-kubeconfig
kubectl create namespace open-cluster-management-agent --dry-run=client -o yaml | kubectl apply -f -
kubectl delete secret bootstrap-hub-kubeconfig -n open-cluster-management-agent --ignore-not-found
kubectl create secret generic bootstrap-hub-kubeconfig --from-file=kubeconfig=e2e-kubeconfig -n open-cluster-management-agent

# install cluster manager and klusterlet operator
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.3/install.sh | bash -s v0.18.3
kubectl apply -f https://operatorhub.io/install/cluster-manager.yaml
kubectl apply -f https://operatorhub.io/install/klusterlet.yaml

## check at least one pod of cluster-manager/klusterlet is Running
n=0
clusterManagerPodCount=0
klusterletManagerPodCount=0
set +e # grep will fail if No resources found by kubectl get
until [[ "${clusterManagerPodCount}" -ge 1 && "${klusterletManagerPodCount}" -ge 1 ]] || [ "$n" -ge 3000 ]
do
    clusterManagerPodCount=$(kubectl get pod -A -l 'app in (cluster-manager)' --ignore-not-found=true | grep -c Running)
    klusterletManagerPodCount=$(kubectl get pod -A -l 'app in (klusterlet)' --ignore-not-found=true | grep -c Running)
    n=$((n+1)) 
    sleep 1
done
set -e

# scale replica to 1 to save resources
kubectl scale --replicas=1 -n kube-system deployment/coredns
kubectl scale --replicas=1 -n operators deployment/cluster-manager
kubectl scale --replicas=1 -n operators deployment/klusterlet

cat <<EOF | kubectl apply -f -
apiVersion: operator.open-cluster-management.io/v1
kind: ClusterManager
metadata:
  name: cluster-manager
spec:
  placementImagePullSpec: 'quay.io/open-cluster-management/placement:v0.1.0'
  registrationImagePullSpec: 'quay.io/open-cluster-management/registration:v0.4.0'
  workImagePullSpec: 'quay.io/open-cluster-management/work:v0.4.0'
EOF

cat <<EOF | kubectl apply -f -
apiVersion: operator.open-cluster-management.io/v1
kind: Klusterlet
metadata:
  name: klusterlet
spec:
  clusterName: cluster1
  externalServerURLs:
    - url: 'https://localhost'
  namespace: open-cluster-management-agent
  registrationImagePullSpec: 'quay.io/open-cluster-management/registration:v0.4.0'
  workImagePullSpec: 'quay.io/open-cluster-management/work:v0.4.0'
EOF

# approve cluster join request and csr
n=0
until [ "$n" -ge 120 ]
do
    kubectl patch managedcluster cluster1 -p='{"spec":{"hubAcceptsClient":true}}' --type=merge && break
    n=$((n+1)) 
    sleep 1
done

clusterCondition=$(kubectl get managedcluster cluster1 -o jsonpath='{.status.conditions}')
if [[ "$clusterCondition" == *\"reason\":\"ManagedClusterAvailable\",\"status\":\"True\"* ]]; then
    echo "managedcluster cluster1 already be available."
    echo_cluster_registered_ok
    exit 0
fi

n=0
until [ "$n" -ge 120 ]
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
    sleep 1
done

echo_cluster_registered_ok