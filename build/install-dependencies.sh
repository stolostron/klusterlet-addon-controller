#!/bin/bash -e

###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

export GO111MODULE=off

# Go tools

if ! which patter > /dev/null; then      echo "Installing patter ..."; go install github.com/apg/patter; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go install github.com/wadey/gocovmerge; fi
# if ! which swagger > /dev/null; then     echo "Installing swagger..."; go install github.com/rws-github/go-swagger/cmd/swagger; fi
if ! which golangci-lint > /dev/null; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.37.1
fi

if ! which go-bindata > /dev/null; then
	echo "Installing go-bindata..."
	cd $(mktemp -d) && GO111MODULE=off go install github.com/go-bindata/go-bindata/...
fi
go-bindata --version