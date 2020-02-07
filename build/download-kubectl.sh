#!/bin/bash

KUBECTL_VER=v1.16.3
BUILD_DIR=build

echo "Downloading kubectl $KUBECTL_VER ..."
if [ ! -f "$BUILD_DIR/kubectl-linux-amd64" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-amd64     "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/amd64/kubectl"; fi
if [ ! -f "$BUILD_DIR/kubectl-linux-ppc64le" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-ppc64le "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/ppc64le/kubectl"; fi
if [ ! -f "$BUILD_DIR/kubectl-linux-s390x" ]; then curl -f -s -L -o $BUILD_DIR/kubectl-linux-s390x     "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VER}/bin/linux/s390x/kubectl"; fi
echo "Downloaded kubectl to $BUILD_DIR"
