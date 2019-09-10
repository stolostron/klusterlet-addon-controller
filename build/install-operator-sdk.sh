#!/bin/bash

# Licensed Materials - Property of IBM
# 5737-E67
# (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
# US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp.

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