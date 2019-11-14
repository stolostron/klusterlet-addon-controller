#!/bin/bash

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