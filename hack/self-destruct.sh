#!/bin/bash
###############################################################################
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

if [ -z "${OPERATOR_NAMESPACE}" ]; then
	OPERATOR_NAMESPACE="multicluster-endpoint"
fi

# Delete all endpoints.multicloud.ibm.com
kubectl delete endpoints.multicloud.ibm.com -n ${OPERATOR_NAMESPACE}  --all --timeout=60s

# Delete Deployment
kubectl delete deployment endpoint-operator -n ${OPERATOR_NAMESPACE}

# Force delete all component CRDs if they still exist
component_crds=(
	applicationmanagers.multicloud.ibm.com
	certpoliciescontroller.multicloud.ibm.com
	ciscontrollers.multicloud.ibm.com
	connectionmanagers.multicloud.ibm.com
	iampoliciescontroller.multicloud.ibm.com
	policycontrollers.multicloud.ibm.com
	searchcollectors.multicloud.ibm.com
	workmanagers.multicloud.ibm.com
	endpoints.multicloud.ibm.com
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
