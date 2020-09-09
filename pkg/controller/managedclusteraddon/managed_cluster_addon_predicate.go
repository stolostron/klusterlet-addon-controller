// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package managedclusteraddon

import (
	"reflect"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	addons "github.com/open-cluster-management/endpoint-operator/pkg/components"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// return if object is controlled by a klusterlettaddonconfig
func ownerIsKlusterletAddonConfig(obj metav1.Object) bool {
	if obj == nil {
		return false
	}
	// If filtered to a Controller, only take the Controller OwnerReference
	if ownerRef := metav1.GetControllerOf(obj); ownerRef != nil {
		refGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
		if err != nil {
			log.Error(err, "Could not parse OwnerReference APIVersion",
				"api version", ownerRef.APIVersion)
			return false
		}
		if ownerRef.Kind == "KlusterletAddonConfig" && refGV.Group == agentv1.SchemeGroupVersion.Group {
			return true
		}
	}

	return false
}

// check if the given updateEvent is not nil, will return boolean and an error message
func isValidUpdateEventVariable(e event.UpdateEvent) (bool, string) {
	if e.MetaOld == nil {
		return false, "Update event has no old metadata"
	}
	if e.ObjectOld == nil {
		return false, "Update event has no old runtime object to update"
	}
	if e.ObjectNew == nil {
		return false, "Update event has no new runtime object for update"
	}
	if e.MetaNew == nil {
		return false, "Update event has no new metadata"
	}
	return true, ""
}

// newManifestWorkPredicate allows request object with a valid name (can be converted to an addon) to reconcile
// the manifestwork has to be owned by a klusterletaddonconfig
// for update event, will only allow reconcile when status is changed
func newManifestWorkPredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManifestWorkName(e.Meta.GetName()); err != nil {
				return false
			}
			return ownerIsKlusterletAddonConfig(e.Meta)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManifestWorkName(e.Meta.GetName()); err != nil {
				return false
			}
			return ownerIsKlusterletAddonConfig(e.Meta)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if isValid, msg := isValidUpdateEventVariable(e); !isValid {
				log.Error(nil, msg, "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManifestWorkName(e.MetaOld.GetName()); err != nil {
				return false
			}
			if _, err := addons.GetAddonFromManifestWorkName(e.MetaNew.GetName()); err != nil {
				return false
			}
			newManifestWork, okNew := e.ObjectNew.(*manifestworkv1.ManifestWork)
			oldManifestWork, okOld := e.ObjectOld.(*manifestworkv1.ManifestWork)
			if newManifestWork.DeletionTimestamp != nil && oldManifestWork.DeletionTimestamp == nil {
				return ownerIsKlusterletAddonConfig(e.MetaNew)
			}
			if okNew && okOld {
				return !reflect.DeepEqual(newManifestWork.Status, oldManifestWork.Status)
			}
			return false
		},
	})
}

// newManagedClusterAddonNamePredicate allows request object with a valid name,
// which means the name can be converted to an addon, to reconcile
func newManagedClusterAddonNamePredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Create event has no runtime object to create", "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManagedClusterAddonName(e.Meta.GetName()); err != nil {
				return false
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if e.Object == nil {
				log.Error(nil, "Delete event has no runtime object to delete", "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManagedClusterAddonName(e.Meta.GetName()); err != nil {
				return false
			}
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if isValid, msg := isValidUpdateEventVariable(e); !isValid {
				log.Error(nil, msg, "event", e)
				return false
			}
			if _, err := addons.GetAddonFromManagedClusterAddonName(e.MetaOld.GetName()); err != nil {
				return false
			}
			if _, err := addons.GetAddonFromManagedClusterAddonName(e.MetaNew.GetName()); err != nil {
				return false
			}
			return true
		},
	})
}
