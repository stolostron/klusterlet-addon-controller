// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package controller

import (
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller/csr"
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller/klusterletaddon"
	"github.com/stolostron/klusterlet-addon-controller/pkg/controller/managedclusteraddon"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(manager.Manager, kubernetes.Interface) error

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs,
		clustermanagementaddon.Add,
		managedclusteraddon.Add,
		klusterletaddon.Add,
		csr.Add)
}

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, kubeClient kubernetes.Interface) error {
	for _, f := range AddToManagerFuncs {
		if err := f(m, kubeClient); err != nil {
			return err
		}
	}
	return nil
}
