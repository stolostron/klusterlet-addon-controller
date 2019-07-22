IMAGE_NAME ?= ibm-klusterlet-operator
SCRATCH_REPO ?= hyc-cloud-private-scratch-docker-local.artifactory.swg-devops.com/ibmcom
SCRATCH_TAG ?= ${shell whoami}

.PHONY: install-crd
install-crd:
	for file in `ls deploy/crds/*crd.yaml`; do kubectl apply -f $$file; done

.PHONY: operator\:run
operator\:run:
	operator-sdk up local --namespace="" --operator-flags="--zap-encoder=console"

.PHONY: operator\:build
operator\:build:
	operator-sdk build ${SCRATCH_REPO}/${IMAGE_NAME}:${SCRATCH_TAG}

