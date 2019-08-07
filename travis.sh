#!/bin/bash -e

# IBM Confidential
# OCO Source Materials
# 5737-E67
# (C) Copyright IBM Corporation 2018 All Rights Reserved
# The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

function announce() {
  travis_time_start
  echo \$ $@
  $@
  travis_time_finish
}

function travis_time_start() {
  travis_timer_id=$(printf %08x $(( RANDOM * RANDOM )))
  travis_start_time=$(travis_nanoseconds)
  echo -en "travis_time:start:$travis_timer_id\r${ANSI_CLEAR}"
}

function travis_time_finish() {
  local result=$?
  travis_end_time=$(travis_nanoseconds)
  local duration=$(($travis_end_time-$travis_start_time))
  echo -en "\ntravis_time:end:$travis_timer_id:start=$travis_start_time,finish=$travis_end_time,duration=$duration\r${ANSI_CLEAR}"
  return $result
}

function travis_nanoseconds() {
  local cmd="date"
  local format="+%s%N"
  local os=$(uname)

  if hash gdate > /dev/null 2>&1; then
    cmd="gdate" # use gdate if available
  elif [[ "$os" = Darwin ]]; then
    format="+%s000000000" # fallback to second precision on darwin (does not support %N)
  fi

  $cmd -u $format
}

function fold_start() {
  echo -e "travis_fold:start:$1\033[33;1m$2\033[0m"
}

function fold_end() {
  echo -e "\ntravis_fold:end:$1\r"
}

echo TARGET=$TARGET
echo OS=$OS
echo TRAVIS_OS_NAME=$TRAVIS_OS_NAME
echo ARCH=$ARCH
echo COMMIT=$COMMIT

fold_start deps "Dependencies"
# work around for pulling go binary from place other than github.com
git config --global url.git@github.ibm.com:.insteadOf https://github.ibm.com/
announce make deps
fold_end deps

fold_start build "Build"
announce make build
fold_end build

fold_start check "Check"
announce make check
fold_end check

fold_start api "API"
announce make swagger:diff
fold_end api

fold_start test "Test"
announce make test
fold_end test

fold_start tools "Operator SDK Install"
announce make operator:tools
fold_end tools

fold_start image "Image"
announce make image
fold_end image

if [[ "$TRAVIS_EVENT_TYPE" != "pull_request" ]]; then
  fold_start publish "Publish"
  # publish to Artifactory
  export DOCKER_REGISTRY=hyc-cloud-private-integration-docker-local.artifactory.swg-devops.com
  export DOCKER_USER=$ARTIFACTORY_USER
  export DOCKER_PASS=$ARTIFACTORY_KEY
  export DOCKER_NAMESPACE=ibmcom
  announce make docker:login
  if [[ "$TRAVIS_OS_NAME" == "linux" ]]; then
      export DOCKER_TAG=$DOCKER_TAG
      announce make docker:push-arch
  fi
  fold_end publish
else
  echo "Not pushing image on pull request"
fi
