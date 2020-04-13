#!/bin/bash
###############################################################################
# (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
# Note to U.S. Government Users Restricted Rights:
# U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
# Contract with IBM Corp.
# Licensed Materials - Property of IBM
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

# PARAMETERS
# $1 - API directory (the directory containing the swagger.json and previous versions)
# $2 - Previous version

__diff_jar=".build/openapi-diff-2.0.1.jar"
__diff_doc=".build/.api-diff.log"
__dltd_doc=".build/.api-diff-deleted.log"
__chng_doc=".build/.api-diff-changed.log"
__dltd_exempts="$1/swagger-$2-deletion-exemptions.txt"
__chng_exempts="$1/swagger-$2-change-exemptions.txt"
__section_separator="--------------------------------------------------------------------------"
__deleted_start="--                            What's Deleted                            --"
__changed_start="--                            What's Changed                            --"

mkdir -p .build
if [ ! -f "$__diff_jar" ]; then
    curl -f -s -L -o $__diff_jar https://github.com/rws-github/openapi-diff/releases/download/2.0.1-SNAPSHOT/openapi-diff-2.0.1-SNAPSHOT-jar-with-dependencies.jar
fi

# Converts a Swagger 2.0 doc to a OpenAPI 3.0.0 doc
# $1 - Swagger 2.0 file
# $2 - OpenApi 3.0.0 file
convertSwaggerToOpenAPI () {
    swagger2openapi "$1" > "$2" 2>/dev/null
}

if ! which swagger2openapi > /dev/null; then
    echo "Installing swagger2openapi npm library ..."
    npm install -g swagger2openapi
fi

convertSwaggerToOpenAPI $1/swagger.json $1/openapi.json

echo "Diffing $1/openapi.json against $1/openapi-$2.json"
java -jar $__diff_jar $1/openapi-$2.json $1/openapi.json > $__diff_doc

__exit_code=0

# tac is available on Linux and macOS must use tail -r
if which tac > /dev/null; then
    grep -A 1000 -e "$__deleted_start" $__diff_doc | tail -n +3 | grep -B 1000 -m 1 -e "$__section_separator" | tac | tail -n +3 | tac > $__dltd_doc
else
    grep -A 1000 -e "$__deleted_start" $__diff_doc | tail -n +3 | grep -B 1000 -m 1 -e "$__section_separator" | tail -r | tail -n +3 | tail -r > $__dltd_doc
fi
if [ -a "$__dltd_exempts" ]; then
    # Verify the exemptions match the findings
    if [[ $(< $__dltd_exempts) != $(< $__dltd_doc) ]]; then
        echo "Found unexpected deleted content in API $__api_name"
        echo "FOUND:"
        cat $__dltd_doc
        echo ""
        echo "EXPECTED:"
        cat $__dltd_exempts
        echo ""
        echo ""
        __exit_code=1
    fi
elif grep -e "$__deleted_start" $__diff_doc ; then
    echo "Found deleted content in API $__api_name"
    cat $__dltd_doc
    __exit_code=./
fi

# Scan for changes that break the API
if which tac > /dev/null; then
    grep -A 1000 -e "$__changed_start" $__diff_doc | tail -n +3 | grep -B 1000 -m 1 -e "$__section_separator" | tac | tail -n +3 | tac > $__chng_doc
else
    grep -A 1000 -e "$__changed_start" $__diff_doc | tail -n +3 | grep -B 1000 -m 1 -e "$__section_separator" | tail -r | tail -n +3 | tail -r > $__chng_doc
fi
if [ -a "$__chng_exempts" ]; then
    # Verify the exemptions match the findings
    if [[ $(< $__chng_exempts) != $(< $__chng_doc) ]]; then
        echo "Found unexpected changed content in API $__api_name:"
        cat $__chng_doc
        __exit_code=1
    fi
elif grep -e "$__changed_start" $__diff_doc > /dev/null ; then
    echo "Found changed content in API $__api_name"
    cat $__chng_doc
    __exit_code=1
fi

rm -f .build/.api-diff*.log

if [ "$__exit_code" == "0" ]; then
    echo "PASSED api-diff against $2"
else
    echo "FAILED api-diff against $2"
fi

exit $__exit_code