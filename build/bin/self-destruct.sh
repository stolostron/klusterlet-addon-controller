#!/bin/bash

if [ -z "${OPERATOR_NAMESPACE}" ]; then
	OPERATOR_NAMESPACE="multicluster-endpoint"
fi

# Delete all endpoints.multicloud.ibm.com 
kubectl delete endpoints.multicloud.ibm.com -n ${OPERATOR_NAMESPACE}  --all --timeout=180s

# Delete Deployment
kubectl delete deployment ibm-multicluster-endpoint-operator -n ${OPERATOR_NAMESPACE}

# Force delete all component CRDs if they still exist 
component_crds=(
	applicationmanagers.multicloud.ibm.com
	certmanagers.multicloud.ibm.com
	connectionmanagers.multicloud.ibm.com
	meterings.multicloud.ibm.com
	mongodbs.multicloud.ibm.com
	monitorings.multicloud.ibm.com
	policycontrollers.multicloud.ibm.com
	searchcollectors.multicloud.ibm.com
	serviceregistries.multicloud.ibm.com
	tillers.multicloud.ibm.com
	topologycollectors.multicloud.ibm.com
	workmanagers.multicloud.ibm.com
)

for crd in "${component_crds[@]}"; do
	for resource in `kubectl get ${crd} -o name -n multicluser-endpoint`; do
		# attemp to delete the component resource
		kubectl delete ${resource} -n multicluster-endpoint --timeout=60s
		# force remove the component resource by removing finalizer
		kubectl patch ${resource} --type="json" -p '[{"op": "remove", "path":"/metadata/finalizers"}]'
	done
done

kubectl delete namespace ${OPERATOR_NAMESPACE}
