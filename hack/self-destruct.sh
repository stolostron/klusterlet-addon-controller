#!/bin/bash
###############################################################################
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

if [ -z "${OPERATOR_NAMESPACE}" ]; then
	OPERATOR_NAMESPACE="klusterlet"
fi

# Delete all klusterlets.agent.open-cluster-management.io
kubectl delete klusterlets.agent.open-cluster-management.io -n ${OPERATOR_NAMESPACE}  --all --timeout=60s

# Delete Deployment
kubectl delete deployment klusterlet-operator -n ${OPERATOR_NAMESPACE}

# Force delete all component CRDs if they still exist
component_crds=(
	applicationmanagers.agent.open-cluster-management.io
	certpoliciescontroller.agent.open-cluster-management.io
	ciscontrollers.agent.open-cluster-management.io
	connectionmanagers.agent.open-cluster-management.io
	iampoliciescontroller.agent.open-cluster-management.io
	policycontrollers.agent.open-cluster-management.io
	searchcollectors.agent.open-cluster-management.io
	workmanagers.agent.open-cluster-management.io
	klusterlets.agent.open-cluster-management.io
)

for crd in "${component_crds[@]}"; do
	echo "force delete all CustomResourceDefination ${crd} resources..."
	for resource in `kubectl get ${crd} -o name -n ${OPERATOR_NAMESPACE}`; do
		echo "attempt to delete ${crd} resource ${resource}..."
		kubectl delete ${resource} -n ${OPERATOR_NAMESPACE} --timeout=15s
		echo "force remove ${crd} resource ${resource}..."
		kubectl patch ${resource} -n ${OPERATOR_NAMESPACE} --type="json" -p '[{"op": "remove", "path":"/metadata/finalizers"}]'
	done
	echo "force delete all CustomResourceDefination ${crd} resources..."
	kubectl delete crd ${crd}
done

kubectl delete namespace ${OPERATOR_NAMESPACE}
