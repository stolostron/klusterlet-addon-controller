#!/bin/bash -e

# NOTE: This script should not be called directly. Please run `make test`.

set -o pipefail

_package=$1
_cover_pkgs=$2
echo -e "\nTesting package $_package"

# Make sure temporary files do not exist
rm -f test/coverage/cover.tmp

# Support for TAP output
_package_base=${PROJECT_DIR/$GOPATH\/src\/}  # TODO need a better solution since $(go list) doesn't work any more (won't work with go 1.11)
_tap_out_dir=$GOPATH/src/$_package_base/test/out
_tap_name="${_package/$_package_base/}"
_tap_name=${_tap_name//\//_}

mkdir -p $_tap_out_dir

# Run tests
# DO NOT USE -coverpkg=./...
go test -v -cover -coverpkg=$_cover_pkgs -covermode=atomic -coverprofile=test/coverage/cover.tmp $_package 2> >( grep -v "warning: no packages being tested depend on" >&2 ) | $GOPATH/bin/patter | tee $_tap_out_dir/$_tap_name.tap | grep -v "TAP version 13" | grep -v ": PASS:" | grep -v -i "# /us"

# Merge coverage files
if [ -f test/coverage/cover.tmp ]; then
    $GOPATH/bin/gocovmerge test/coverage/cover.tmp test/coverage/cover.out > test/coverage/cover.all
    mv test/coverage/cover.all test/coverage/cover.out
fi

# Clean up temporary files
rm -f test/coverage/cover.tmp
