#!/bin/bash

echo ">>> Installing Operator SDK"
echo ">>> >>> Downloading source code"
go get -d github.com/operator-framework/operator-sdk

cd $GOPATH/src/github.com/operator-framework/operator-sdk

echo ">>> >>> Checking out version 0.9.x"
git checkout v0.9.x

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