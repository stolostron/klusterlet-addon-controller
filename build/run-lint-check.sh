#!/bin/bash

###############################################################################
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

set -x
set -eo pipefail

GOLANGCI_LINT_CACHE=/tmp/golangci-cache
GOOS=$(go env GOOS)
GOPATH=$(go env GOPATH)
export GOFLAGS=""
# export PROJECT_DIR=$(shell 'pwd')
# export BUILD_DIR=$(PROJECT_DIR)/build

if ! which golangci-lint > /dev/null; then
    mkdir -p "${GOPATH}/bin"
    echo "${PATH}" | grep -q "${GOPATH}/bin"
    IN_PATH=$?
    if [ $IN_PATH != 0 ]; then
        echo "${GOPATH}/bin not in $$PATH"
        exit 1
    fi
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
fi

echo 'Running linting tool ...'
$(GOLANGCI_LINT_CACHE=${GOLANGCI_LINT_CACHE} CGO_ENABLED=1 golangci-lint run -c build/golangci.yml)
echo '##### lint-check #### Success'