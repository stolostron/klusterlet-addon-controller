#!/bin/bash

set -e

echo "Running go mod tidy check..."
TRAVIS_BRANCH=$1
go mod tidy
STATUS=$( git status --porcelain go.mod go.sum )

if [[ ! -z "$STATUS"  &&  "x${TRAVIS_BRANCH}" != "x" ]]; then
    echo "FAILED: 'go mod tidy' modified go.mod and/or go.sum"
    exit 1
fi
echo "##### go-mod-check #### Success"
exit 0
