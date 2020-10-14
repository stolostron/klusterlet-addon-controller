// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package utils contains common utility functions that gets call by many differerent packages
package utils

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
)

var log = logf.Log.WithName("utils")

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

//AddFinalizer accepts cluster and adds provided finalizer to cluster
func AddFinalizer(o metav1.Object, finalizer string) {
	for _, f := range o.GetFinalizers() {
		if f == finalizer {
			return
		}
	}

	o.SetFinalizers(append(o.GetFinalizers(), finalizer))
}

//RemoveFinalizer accepts cluster and removes provided finalizer if present
func RemoveFinalizer(o metav1.Object, finalizer string) {
	var finalizers []string

	for _, f := range o.GetFinalizers() {
		if f != finalizer {
			finalizers = append(finalizers, f)
		}
	}

	if len(finalizers) == len(o.GetFinalizers()) {
		return
	}

	o.SetFinalizers(finalizers)
}

//HasFinalizer checks if a finalizer present
func HasFinalizer(o metav1.Object, finalizer string) bool {
	for _, f := range o.GetFinalizers() {
		if f == finalizer {
			return true
		}
	}
	return false
}

// compareManifestWorks returns true if 2 manifestworks' specs are the same
func compareManifestWorks(mw1 *manifestworkv1.ManifestWork, mw2 *manifestworkv1.ManifestWork) bool {
	if mw1 == nil && mw2 == nil {
		return true
	}
	if (mw1 == nil && mw2 != nil) || (mw2 == nil && mw1 != nil) {
		return false
	}
	if len(mw1.Spec.Workload.Manifests) != len(mw2.Spec.Workload.Manifests) {
		return false
	}
	used := make(map[int]bool)
	for _, m1 := range mw1.Spec.Workload.Manifests {
		hasMatch := false
		for j, m2 := range mw2.Spec.Workload.Manifests {
			if used[j] {
				continue
			}
			if compareManifests(&m1.RawExtension, &m2.RawExtension) {
				hasMatch = true
				used[j] = true
				break
			}
		}
		if !hasMatch {
			return false
		}
	}
	return true
}

// convertRawExtensiontoUnstructured converts a rawExtension to a unstructured object
func convertRawExtensiontoUnstructured(r *runtime.RawExtension) (*unstructured.Unstructured, error) {
	if r == nil {
		return nil, fmt.Errorf("fail to convert rawExtension")
	}
	var obj runtime.Object
	var scope conversion.Scope
	err := runtime.Convert_runtime_RawExtension_To_runtime_Object(r, &obj, scope)
	if err != nil {
		log.Error(err, "failed to convert rawExtension to runtime.Object", "rawExtension", r)
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		log.Error(err, "failed to convert runtime.Objectt to Unstructured", "runtime.Object", &obj)
		return nil, err
	}
	u := unstructured.Unstructured{Object: innerObj}
	return &u, nil
}

var rootAttributes = []string{
	"spec",
	"rules",
	"roleRef",
	"subjects",
	"secrets",
	"imagePullSecrets",
	"automountServiceAccountToken",
	"data",
}

// compareManifests compares if 2 manifests are the same, it only checks value we care
// (name/namespace/kind/group/spec/data)
func compareManifests(r1, r2 *runtime.RawExtension) bool {
	u1, err := convertRawExtensiontoUnstructured(r1)
	if err != nil {
		return false
	}
	u2, err := convertRawExtensiontoUnstructured(r2)
	if err != nil {
		return false
	}
	if u1 == nil || u2 == nil {
		return u2 == nil && u1 == nil
	}
	if u1.GetName() != u2.GetName() ||
		u1.GetNamespace() != u2.GetNamespace() ||
		u1.GetKind() != u2.GetKind() ||
		u1.GetAPIVersion() != u2.GetAPIVersion() {
		return false
	}
	hasDiff := false
	for _, r := range rootAttributes {
		if newValue, ok := u2.Object[r]; ok {
			if !reflect.DeepEqual(newValue, u1.Object[r]) {
				hasDiff = true
			}
		} else {
			if _, ok := u1.Object[r]; ok {
				hasDiff = true
			}
		}
	}
	return !hasDiff
}

// CreateOrUpdateManifestWork creates a new ManifestWork or update an existing ManifestWork
func CreateOrUpdateManifestWork(
	manifestwork *manifestworkv1.ManifestWork,
	client client.Client,
	owner metav1.Object,
	scheme *runtime.Scheme,
) error {
	var oldManifestwork manifestworkv1.ManifestWork

	err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: manifestwork.Name, Namespace: manifestwork.Namespace},
		&oldManifestwork,
	)
	if err == nil {
		// Check if update is require
		if !compareManifestWorks(&oldManifestwork, manifestwork) {
			oldManifestwork.Spec.Workload.Manifests = manifestwork.Spec.Workload.Manifests
			if err := client.Update(context.TODO(), &oldManifestwork); err != nil {
				log.Error(err, "Fail to update manifestwork")
				return err
			}
		}
	} else {
		if errors.IsNotFound(err) {
			if err := controllerutil.SetControllerReference(owner, manifestwork, scheme); err != nil {
				log.Error(err, "Unable to SetControllerReference")
				return err
			}
			if err := client.Create(context.TODO(), manifestwork); err != nil {
				log.Error(err, "Fail to create manifestwork")
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

// DeleteManifestWork deletes a manifestwork
// if removeFinalizers is set to true, will remove all finalizers to make sure it can be deleted
func DeleteManifestWork(name, namespace string, client client.Client, removeFinalizers bool) error {
	manifestWork := &manifestworkv1.ManifestWork{}
	var retErr error
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace},
		manifestWork,
	); err != nil {
		return err
	}

	if removeFinalizers && len(manifestWork.GetFinalizers()) > 0 {
		manifestWork.SetFinalizers([]string{})
		if err := client.Update(context.TODO(), manifestWork); err != nil {
			log.Error(err, fmt.Sprintf("Failed to remove finalizers of Manifestwork %s in %s namespace", name, namespace))
			retErr = err
		}
	}

	if manifestWork.DeletionTimestamp == nil {
		err := client.Delete(context.TODO(), manifestWork)
		if err != nil {
			return err
		}
	}

	return retErr
}

func GetManifestWork(name, namespace string, client client.Client) (*manifestworkv1.ManifestWork, error) {
	manifestWork := &manifestworkv1.ManifestWork{}

	if err := client.Get(
		context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace},
		manifestWork,
	); err != nil {
		return nil, err
	}

	return manifestWork, nil
}
