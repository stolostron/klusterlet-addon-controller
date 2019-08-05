// Package controller contain the controller and the main reconcile function for the operator
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package controller

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/controller/endpoint"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, endpoint.Add)
}
