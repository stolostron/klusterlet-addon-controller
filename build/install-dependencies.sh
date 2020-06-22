#!/bin/bash -e
###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Licensed Materials - Property of IBM
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

export GO111MODULE=off

# Go tools

if ! which patter > /dev/null; then      echo "Installing patter ..."; go get -u github.com/apg/patter; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go get -u github.com/wadey/gocovmerge; fi
# if ! which swagger > /dev/null; then     echo "Installing swagger..."; go get -u github.com/rws-github/go-swagger/cmd/swagger; fi
if ! which golangci-lint > /dev/null; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
fi
if ! which ossc > /dev/null; then
	# do a get in a tmp dir to avoid local go.mod update
	cd $(mktemp -d) && GOSUMDB=off go get -u github.com/open-cluster-management/go-ossc/ossc
fi

if ! which go-bindata > /dev/null; then
	echo "Installing go-bindata..."
	cd $(mktemp -d) && GOSUMDB=off go get -u github.com/go-bindata/go-bindata/...
fi
go-bindata --version