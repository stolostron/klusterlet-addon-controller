#!/bin/bash

echo ">>> Installing Operator SDK"
echo ">>> >>> Downloading source code"
# cannot use 'set -e' because this command fails after project has been cloned down for some reason
GO111MODULE=off go get -d github.com/operator-framework/operator-sdk

cd $GOPATH/src/github.com/operator-framework/operator-sdk

echo ">>> >>> Checking out v0.15.1"
git checkout v0.15.1

echo ">>> >>> Running make tidy"
make tidy

echo ">>> >>> Running make install"
make install

echo ">>> Done installing Operator SDK"

operator-sdk version
if [ $? != 0 ]; then
  echo ">>>> opereattor-sdk installation failed"
  exit 1
fi
