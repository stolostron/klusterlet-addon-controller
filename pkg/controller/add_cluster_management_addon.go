// Copyright (c) 2020 Red Hat, Inc.

// Package controller contain the controller and the main reconcile function for the operator
package controller

import (
	"github.com/open-cluster-management/endpoint-operator/pkg/controller/clustermanagementaddon"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, clustermanagementaddon.Add)
}
