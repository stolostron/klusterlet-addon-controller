// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// Package controller contain the controller and the main reconcile function for the operator
package controller

import (
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clustermanagementaddon.Add)
}
