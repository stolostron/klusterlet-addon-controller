//Package utils contains common utility functions that gets call by many differerent packages
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func TestUniqueStringSlice(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	testCases := []struct {
		Input          []string
		ExpectedOutput []string
	}{
		{[]string{"foo", "bar"}, []string{"foo", "bar"}},
		{[]string{"foo", "bar", "bar"}, []string{"foo", "bar"}},
		{[]string{"foo", "foo", "bar", "bar"}, []string{"foo", "bar"}},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.ExpectedOutput, UniqueStringSlice(testCase.Input))
	}
}
