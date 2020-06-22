#!/bin/bash
###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Licensed Materials - Property of IBM
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

set -e

echo "Running go mod tidy check..."
TRAVIS_BRANCH=$1
go mod tidy
STATUS=$( git status --porcelain go.mod go.sum )

if [[ ! -z "$STATUS"  &&  "x${TRAVIS_BRANCH}" != "x" ]]; then
    echo "FAILED: 'go mod tidy' modified go.mod and/or go.sum"
    exit 1
fi
echo "##### go-mod-check #### Success"
exit 0
