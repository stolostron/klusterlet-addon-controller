// Copyright (c) 2020 Red Hat, Inc.

package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"github.com/prometheus/common/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type klusterletAddonConfigValidator struct{}

// Handle set the default values to every incoming KlusterletAddonConfig cr.
// Currently only handles create/update
func (k *klusterletAddonConfigValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == "CREATE" {
		err := k.validateCreate(req)
		if err != nil {
			log.Info("Create denied")
			return admission.Denied(err.Error())
		}
		log.Info("Create successful")
		return admission.Allowed("")
	}
	//If not create update
	if req.Operation == "UPDATE" {
		err := k.validateUpdate(req)
		if err != nil {
			log.Info("Update denied")
			return admission.Denied(err.Error())
		}
		log.Info("Update successful")
		return admission.Allowed("")
	}

	return admission.Allowed("")
}

func (k *klusterletAddonConfigValidator) validateCreate(req admission.Request) error {
	if _, err := k.validateKlusterletAddonConfigObj(req); err != nil {
		return err
	}

	return nil
}

func (k *klusterletAddonConfigValidator) validateUpdate(req admission.Request) error {
	// Parse existing and new MultiClusterHub resources
	existingKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := json.Unmarshal(req.OldObject.Raw, existingKlusterletAddonConfig); err != nil {
		return err
	}

	if _, err := k.validateKlusterletAddonConfigObj(req); err != nil {
		return err
	}

	return nil
}

func (k *klusterletAddonConfigValidator) validateKlusterletAddonConfigObj(req admission.Request) (*agentv1.KlusterletAddonConfig, error) {
	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := json.Unmarshal(req.Object.Raw, klusterletAddonConfig); err != nil {
		return klusterletAddonConfig, err
	}

	newVersion, err := semver.NewVersion(klusterletAddonConfig.Spec.Version)
	if err != nil {
		return klusterletAddonConfig, fmt.Errorf("Version %q is invalid semantic version", klusterletAddonConfig.Spec.Version)
	}

	versionList, err := klusterletAddonConfig.GetAvailableVersions()
	if err != nil {
		return klusterletAddonConfig, err
	}

	if !IsValidVersion(versionList, newVersion) {
		return klusterletAddonConfig, fmt.Errorf("Version %s is not available. Available Versions are: %s", klusterletAddonConfig.Spec.Version, versionList)
	}

	return klusterletAddonConfig, nil
}

// IsValidVersion check if version is valid
func IsValidVersion(versionList []*semver.Version, newVersion *semver.Version) bool {
	for _, v := range versionList {
		if v.Compare(newVersion) == 0 {
			return true
		}
	}

	return false
}
