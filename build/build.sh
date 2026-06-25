#!/bin/bash -e

###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Copyright (c) Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project
###############################################################################

# PARAMETERS
# $1 - Final image name and tag to be produced
echo Building klusterlet addon controller
echo GOOS: $GOOS
echo GOARCH: $GOARCH

docker build . \
$DOCKER_BUILD_OPTS \
-t $DOCKER_IMAGE:$DOCKER_BUILD_TAG \
-f build/Dockerfile

if [ ! -z "$1" ]; then
    echo "Retagging image as $1"
    docker tag $DOCKER_IMAGE:$DOCKER_BUILD_TAG $1
fi
