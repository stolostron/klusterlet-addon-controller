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

if ! which patter > /dev/null; then      echo "Installing patter ..."; go install github.com/apg/patter@latest; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go install github.com/alexfalkowski/gocovmerge/v2@v2.14.0; fi
# if ! which swagger > /dev/null; then     echo "Installing swagger..."; go install github.com/rws-github/go-swagger/cmd/swagger; fi
