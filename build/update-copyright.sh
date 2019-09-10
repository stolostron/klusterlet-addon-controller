#!/bin/bash

# Licensed Materials - Property of IBM
# 5737-E67
# (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
# US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp.

file=$1 
type=$2
COPYRIGHT="$type Licensed Materials - Property of IBM
$type 5737-E67
$type (C) Copyright IBM Corporation 2016, 2019 All Rights Reserved
$type US Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
"

echo "$COPYRIGHT" > /tmp/copyright 
cat $file >> /tmp/copyright
cp /tmp/copyright $file