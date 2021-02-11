#!/bin/bash -e

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
