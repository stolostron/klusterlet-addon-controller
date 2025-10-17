#!/bin/bash

###############################################################################
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

set -x
set -eo pipefail

GOLANGCI_LINT_VERSION="v2.4.0"
GOLANGCI_LINT_CACHE=/tmp/golangci-cache
export GOFLAGS=""
PROJECT_DIR=$(pwd)
BUILD_DIR=${PROJECT_DIR}/build

# Use the default from vendor/github.com/openshift/build-machinery-go/make/lib/tmp.mk
# PERMANENT_TMP_GOPATH defaults to _output/tools
PERMANENT_TMP_GOPATH=${PERMANENT_TMP_GOPATH:-${PROJECT_DIR}/_output/tools}

# Install golangci-lint locally in project directory
GOLANGCI_LINT=${PERMANENT_TMP_GOPATH}/bin/golangci-lint
golangci_lint_gen_dir=$(dirname ${GOLANGCI_LINT})

if [[ ! -f "${GOLANGCI_LINT}" ]]; then
    echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION} into '${GOLANGCI_LINT}'"
    mkdir -p "${golangci_lint_gen_dir}"
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${golangci_lint_gen_dir}" ${GOLANGCI_LINT_VERSION}
else
    echo "Using existing golangci-lint from '${GOLANGCI_LINT}'"
fi

echo 'Running linting tool ...'
CGO_ENABLED=1 ${GOLANGCI_LINT} run -c build/golangci.yml
echo '##### lint-check #### Success'
