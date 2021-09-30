#!/bin/bash

###############################################################################
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################


./e2e.test -test.v -ginkgo.v > output 2>&1

cat output
kubectl get mcl -o yaml
kubectl get pods -n open-cluster-management
kubectl get pods -n open-cluster-management-agent
kubectl get klusterletaddonconfigs.agent.open-cluster-management.io -n cluster1 cluster1 -o yaml
kubectl get manifestworks -n cluster1 -o yaml
kubectl get appliedmanifestworks.work.open-cluster-management.io -o yaml
kubectl get ns open-cluster-management-agent-addon -o yaml
kubectl describe ns open-cluster-management-agent-addon
kubectl get pods -n open-cluster-management-agent | grep klusterlet-work-agent | awk '{print $1}' | xargs kubectl logs -n open-cluster-management-agent
kubectl get pods -n open-cluster-management | grep klusterlet-addon | awk '{print $1}' | xargs kubectl logs -n open-cluster-management
