#!/bin/bash

set -e

export GO111MODULE=off

# Go tools

if ! which patter > /dev/null; then      echo "Installing patter ..."; go get -u github.com/apg/patter; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go get -u github.com/wadey/gocovmerge; fi
# if ! which swagger > /dev/null; then     echo "Installing swagger..."; go get -u github.com/rws-github/go-swagger/cmd/swagger; fi
if ! which golangci-lint > /dev/null; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.23.6
fi


# Build tools

if ! which operator-sdk > /dev/null; then
    OPERATOR_SDK_VER=v0.15.1
    curr_dir=$(pwd)
    echo ">>> Installing Operator SDK"
    echo ">>> >>> Downloading source code"
    set +e
    # cannot use 'set -e' because this command always fails after project has been cloned down for some reason
    go get -d github.com/operator-framework/operator-sdk
    set -e
    cd $GOPATH/src/github.com/operator-framework/operator-sdk
    echo ">>> >>> Checking out $OPERATOR_SDK_VER"
    git checkout $OPERATOR_SDK_VER
    echo ">>> >>> Running make tidy"
    GO111MODULE=on make tidy
    echo ">>> >>> Running make install"
    GO111MODULE=on make install
    echo ">>> Done installing Operator SDK"
    operator-sdk version
    cd $curr_dir
fi


# Tools built into image

KUBECTL_VER=v1.16.3
echo "Downloading kubectl $KUBECTL_VER ..."
if [ ! -f "$BUILD_DIR/kubectl-linux-amd64" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-amd64     "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/amd64/kubectl"; fi
if [ ! -f "$BUILD_DIR/kubectl-linux-ppc64le" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-ppc64le "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/ppc64le/kubectl"; fi
if [ ! -f "$BUILD_DIR/kubectl-linux-s390x" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-s390x     "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/s390x/kubectl"; fi
echo "Downloaded kubectl to $BUILD_DIR"


# Misc tools

# # TIPS: If the npm install fails because of permissions, do not run the command with sudo, just run:
# # sudo chown -R $(whoami) ~/.npm
# # sudo chown -R $(whoami) /usr/local/lib/node_modules
# _tmpdir=/tmp/apilint
# if ! which apilint > /dev/null; then
#     echo "Installing apilint..."
#     rm -rf $_tmpdir
#     git clone https://github.com/rws-github/apilint $_tmpdir
#     cd $_tmpdir
#     npm install --production
#     npm install -g $_tmpdir
# fi
