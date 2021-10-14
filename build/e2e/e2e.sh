#!/bin/bash

###############################################################################
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

./e2e.test -test.v -ginkgo.v > output

cat output
kubectl get manifestworks.work.open-cluster-management.io -n cluster1 -o yaml
kubectl get appliedmanifestworks.work.open-cluster-management.io -o yaml
kubectl get pods -n open-cluster-management-agent-addon
kubectl get applicationmanagers.agent.open-cluster-management.io -n open-cluster-management-agent-addon klusterlet-addon-appmgr -o yaml
