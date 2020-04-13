// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package utils contains common utility functions that gets call by many differerent packages
package utils

// UniqueStringSlice takes a string[] and remove the duplicate value
func UniqueStringSlice(stringSlice []string) []string {
	keys := make(map[string]bool)
	uniqueStringSlice := []string{}

	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			uniqueStringSlice = append(uniqueStringSlice, entry)
		}
	}
	return uniqueStringSlice
}
