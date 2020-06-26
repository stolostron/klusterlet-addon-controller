// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ghodss/yaml"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/open-cluster-management/endpoint-operator/pkg/bindata"
	"github.com/open-cluster-management/endpoint-operator/pkg/utils"
)

// constants for delete work and finalizer
const (
	KlusterletAddonFinalizer   = "agent.open-cluster-management.io/klusterletaddonconfig-cleanup"
	KlusterletAddonCRDsPostfix = "-klusterlet-addon-crds"
)

// createManifestWorkCRD - create manifest work for CRD
func createManifestWorkCRD(klusterletaddonconfig *agentv1.KlusterletAddonConfig, r *ReconcileKlusterletAddon) error {
	allFiles := bindata.AssetNames()
	installFiles := []string{}

	// get crds & aggregate clusterroles
	for _, file := range allFiles {
		if strings.HasPrefix(file, "crds/") && strings.Contains(file, "crd.yaml") {
			installFiles = append(installFiles, file)
		}
		if strings.HasPrefix(file, "resources/managed") && strings.Contains(file, "admin_aggregate_clusterrole.yaml") {
			installFiles = append(installFiles, file)
		}
	}

	// add all files into manifestwork
	var manifests []manifestworkv1.Manifest
	for _, file := range installFiles {
		data, err := bindata.Asset(file)
		if err != nil {
			log.Error(err, "Fail to get file "+file)
			return err
		}
		b, err := yaml.YAMLToJSON(data)
		if err != nil {
			log.Error(err, "Fail to unmarshal crd yaml", "content", data)
			return err
		}
		manifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Raw: b}}
		manifests = append(manifests, manifest)
	}

	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      klusterletaddonconfig.Name + KlusterletAddonCRDsPostfix,
			Namespace: klusterletaddonconfig.Namespace,
			//Labels:    labels,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}

	if err := utils.CreateOrUpdateManifestWork(manifestWork, r.client, klusterletaddonconfig, r.scheme); err != nil {
		log.Error(err, "Failed to create manifest work for CRD")
		return err
	}

	return nil
}
