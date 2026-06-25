//go:build testrunmain
// +build testrunmain

// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func TestRunMain(t *testing.T) {
	go main()
	// hacks for handling signals
	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			fmt.Printf("Signal Interupt: %s", sig.String())
			return
		case syscall.SIGTERM:
			//handle SIGTERM
			fmt.Printf("Signal SIGTERM: %s", sig.String())
			return
		}
	}()
}
