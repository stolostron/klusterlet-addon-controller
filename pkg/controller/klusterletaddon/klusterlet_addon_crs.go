// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	"context"
	"fmt"
	"reflect"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	appmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/appmgr/v1"
	certpolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/certpolicycontroller/v1"
	iampolicyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/iampolicycontroller/v1"
	policyctrl "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/policyctrl/v1"
	search "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/searchcollector/v1"
	workmgr "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/workmgr/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/controller/clustermanagementaddon"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var addonsArray = []addons.KlusterletAddon{
	appmgr.AddonAppMgr{},
	certpolicyctrl.AddonCertPolicyCtrl{},
	iampolicyctrl.AddonIAMPolicyCtrl{},
	policyctrl.AddonPolicyCtrl{},
	search.AddonSearch{},
	workmgr.AddonWorkMgr{},
}

// newCRManifestWork returns ManifestWork of a component CR
func newCRManifestWork(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client) (*manifestworkv1.ManifestWork, error) {
	var cr runtime.Object

	var err error
	cr, err = addon.NewAddonCR(klusterletaddonconfig, addonoperator.KlusterletAddonNamespace)

	if err != nil {
		return nil, err
	}

	// construct manifestwork
	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
			Namespace: klusterletaddonconfig.Namespace,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: []manifestworkv1.Manifest{
					{
						RawExtension: runtime.RawExtension{Object: cr},
					},
				},
			},
		},
	}
	return manifestWork, nil
}

// syncManifestWorkCRs creates/updates/deletes all CR Manifestworks according to klusterletAddonConfig's configuration
// loops through all the components, and return the last error if there are errors, or return nil if succeeded
func syncManifestWorkCRs(klusterletaddonconfig *agentv1.KlusterletAddonConfig, r *ReconcileKlusterletAddon) error {
	var lastErr error
	lastErr = nil

	for _, addon := range addonsArray {
		addonName := addon.GetAddonName()
		if addon.IsEnabled(klusterletaddonconfig) {
			// create Manifestwork if enabled
			if manifestWork, err := newCRManifestWork(addon, klusterletaddonconfig, r.client); err != nil {
				lastErr = err
			} else if err = utils.CreateOrUpdateManifestWork(
				manifestWork,
				r.client,
				klusterletaddonconfig,
				r.scheme,
			); err != nil {
				log.Error(err, "Failed to create manifest work for addon "+addonName)
				lastErr = err
			}
		} else {
			// delete Manifestwork if disabled
			if err := utils.DeleteManifestWork(
				addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
				klusterletaddonconfig.Namespace,
				r.client,
				false,
			); err != nil && !errors.IsNotFound(err) {
				log.Error(err, fmt.Sprintf("Failed to delete %s ManifestWork", addonName))
				lastErr = err
			}
		}
	}

	return lastErr
}

// syncManagedClusterAddonCRs creates/updates/deletes all CR ManagedClusterAddon according to klusterletAddonConfig's configuration
// loops through all the components, and return the last error if there are errors, or return nil if succeeded
func syncManagedClusterAddonCRs(klusterletaddonconfig *agentv1.KlusterletAddonConfig, r *ReconcileKlusterletAddon) error {
	var lastErr error
	lastErr = nil
	for _, addon := range addonsArray {
		if addon.IsEnabled(klusterletaddonconfig) {
			// create ManagedClusterAddon if enabled, and will not block if failed.
			// created ManagedClusterAddon should has controller reference points to the klusterletaddonconfig
			// and it should has the correct AddonRef in status
			if err := updateManagedClusterAddon(addon, klusterletaddonconfig, r.client, r.scheme); err != nil {
				log.Error(err, "Failed to create ManagedClusterAddon "+addon.GetAddonName())
				lastErr = err
			}
		}
	}
	return lastErr
}

// updateManagedClusterAddon updates ManagedClusterAddon to make sure it has correct reference in status
// if ManagedClusterAddon for an addon is not exist, will create the ManagedClusterAddon
// and will set controller reference to be the given klusterletaddonconfig
func updateManagedClusterAddon(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	scheme *runtime.Scheme,
) error {
	managedClusterAddon := &addonv1alpha1.ManagedClusterAddOn{}
	// check if it exists
	if err := client.Get(
		context.TODO(),
		types.NamespacedName{
			Name:      addon.GetManagedClusterAddOnName(),
			Namespace: klusterletaddonconfig.Namespace,
		},
		managedClusterAddon,
	); err != nil && errors.IsNotFound(err) {
		// create new
		newManagedClusterAddon := &addonv1alpha1.ManagedClusterAddOn{
			TypeMeta: metav1.TypeMeta{
				APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
				Kind:       "ManagedClusterAddOn",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      addon.GetManagedClusterAddOnName(),
				Namespace: klusterletaddonconfig.Namespace,
			},
			Spec: addonv1alpha1.ManagedClusterAddOnSpec{
				InstallNamespace: addonoperator.KlusterletAddonNamespace,
			},
		}

		if err := controllerutil.SetControllerReference(klusterletaddonconfig, newManagedClusterAddon, scheme); err != nil {
			log.Error(err, "failed to set controller of ManagedClusterAddOn "+addon.GetManagedClusterAddOnName())
			return err
		}
		if err := client.Create(context.TODO(), newManagedClusterAddon); err != nil {
			log.Error(err, "")
			return err
		}
		managedClusterAddon = newManagedClusterAddon
	} else if err != nil {
		return err
	}
	ref := []addonv1alpha1.ObjectReference{
		{
			Group:     agentv1.SchemeGroupVersion.Group,
			Resource:  "klusterletaddonconfigs",
			Name:      klusterletaddonconfig.Name,
			Namespace: klusterletaddonconfig.Namespace,
		},
	}
	addonMeta := addonv1alpha1.AddOnMeta{}
	addonConf := addonv1alpha1.ConfigCoordinates{}
	if addonMap, ok := clustermanagementaddon.ClusterManagementAddOnMap[addon.GetManagedClusterAddOnName()]; ok {
		addonMeta.Description = addonMap.Description
		addonMeta.DisplayName = addonMap.DisplayName
		addonConf.CRDName = addonMap.CRDName
		addonConf.CRName = klusterletaddonconfig.Name
	}

	if !reflect.DeepEqual(managedClusterAddon.Status.RelatedObjects, ref) ||
		!reflect.DeepEqual(managedClusterAddon.Status.AddOnMeta, addonMeta) ||
		!reflect.DeepEqual(managedClusterAddon.Status.AddOnConfiguration, addonConf) {
		managedClusterAddon.Status.RelatedObjects = ref
		managedClusterAddon.Status.AddOnMeta = addonMeta
		managedClusterAddon.Status.AddOnConfiguration = addonConf

		if err := client.Status().Update(context.TODO(), managedClusterAddon); err != nil {
			log.Error(err, fmt.Sprintf("Failed to update ManagedClusterAddon %s status", managedClusterAddon.Name))
			return err
		}
	}

	return nil
}

// deleteManifestWorkCRs deletes all CR Manifestworks
// returns true if deletion of all components is completed or component not found
func deleteManifestWorkCRs(
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
	removeFinalizers bool) (bool, error) {
	allCompleted := true
	var lastErr error
	lastErr = nil
	for _, addon := range addonsArray {
		err := utils.DeleteManifestWork(
			addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
			klusterletaddonconfig.Namespace,
			client,
			removeFinalizers,
		)
		if err != nil && errors.IsNotFound(err) {
			continue
		}
		allCompleted = false
		if err != nil { // object still exist
			lastErr = err
		}
	}
	return allCompleted, lastErr
}
