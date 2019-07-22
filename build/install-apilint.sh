#!/bin/bash

# IBM Confidential
# OCO Source Materials
# 5737-E67
# (C) Copyright IBM Corporation 2018 All Rights Reserved
# The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

# TIPS:
# If the npm install fails because of permissions, do not run the command with sudo, just run:
# sudo chown -R $(whoami) ~/.npm
# sudo chown -R $(whoami) /usr/local/lib/node_modules

set -e

_tmpdir=/tmp/apilint

if ! which apilint > /dev/null; then
    echo "Installing apilint..."
    rm -rf $_tmpdir
    git clone https://github.com/rws-github/apilint $_tmpdir
    cd $_tmpdir
    npm install --production
    npm install -g $_tmpdir
fi