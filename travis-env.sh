#!/bin/bash

# Licensed Materials - Property of IBM
# 5737-E67
# (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
# US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp.

# IBM Confidential
# OCO Source Materials
# 5737-E67
# (C) Copyright IBM Corporation 2018 All Rights Reserved
# The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

# Release Tag
if [ "$TRAVIS_BRANCH" = "master" ]; then
    DOCKER_TAG=latest
else
    DOCKER_TAG="${TRAVIS_BRANCH#release-}-latest"
fi
if [ "$TRAVIS_TAG" != "" ]; then
    DOCKER_TAG="${TRAVIS_TAG#v}"
fi
export DOCKER_TAG="$DOCKER_TAG"
export COMMIT=${TRAVIS_COMMIT::6}

# Release Tag
echo TRAVIS_EVENT_TYPE=$TRAVIS_EVENT_TYPE
echo TRAVIS_BRANCH=$TRAVIS_BRANCH
echo TRAVIS_TAG=$TRAVIS_TAG
echo DOCKER_TAG="$DOCKER_TAG"
echo COMMIT="$COMMIT"

export ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
	export ARCH="amd64"
fi

echo ARCH="$ARCH"
