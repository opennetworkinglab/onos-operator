package controller

import (
	"github.com/opennetworkinglab/onos-operator/pkg/controller/onoscluster"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, onoscluster.Add)
}
