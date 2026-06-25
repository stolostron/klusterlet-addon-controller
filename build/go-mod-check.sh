#!/bin/bash

###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

set -e

echo "Running go mod tidy check..."
go mod tidy
STATUS=$( git status --porcelain go.mod go.sum )

if [ ! -z "$STATUS" ]; then
    echo "FAILED: 'go mod tidy' modified go.mod and/or go.sum"
    exit 1
fi
echo "##### go-mod-check #### Success"
exit 0
