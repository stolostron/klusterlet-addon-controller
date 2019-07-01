package controller

import (
	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/controller/klusterletservice"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, klusterletservice.Add)
}
