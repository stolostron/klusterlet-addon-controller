// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// GlobalValues defines the global values
// +k8s:openapi-gen=true
type GlobalValues struct {
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	ImagePullSecret string            `json:"imagePullSecret,omitempty"`
	ImageOverrides  map[string]string `json:"imageOverrides,omitempty"`
	NodeSelector    map[string]string `json:"nodeSelector,omitempty"`
	ProxyConfig     map[string]string `json:"proxyConfig,omitempty"`
}

const (
	HTTPProxy  = "HTTP_PROXY"
	HTTPSProxy = "HTTPS_PROXY"
	NoProxy    = "NO_PROXY"
)

// AddonAgentConfig is the configurations for addon agents.
type AddonAgentConfig struct {
	KlusterletAddonConfig    *KlusterletAddonConfig
	ClusterName              string
	NodeSelector             map[string]string
	Registry                 string
	ImagePullSecret          string
	ImagePullSecretNamespace string
	ImagePullPolicy          corev1.PullPolicy
}

const (
	// UpgradeLabel is to label the upgraded manifestWork.
	UpgradeLabel = "open-cluster-management.io/upgrade"

	KlusterletAddonNamespace = "open-cluster-management-agent-addon"
)

const (
	WorkManagerAddonName     = "work-manager"
	ApplicationAddonName     = "application-manager"
	CertPolicyAddonName      = "cert-policy-controller"
	IamPolicyAddonName       = "iam-policy-controller"
	PolicyAddonName          = "policy-controller"
	ConfigPolicyAddonName    = "config-policy-controller"
	PolicyFrameworkAddonName = "governance-policy-framework"
	SearchAddonName          = "search-collector"
)

// KlusterletAddons is for klusterletAddon refactor, set true if the addon is ready to install by itself.
var KlusterletAddons = map[string]bool{
	WorkManagerAddonName:     true,
	ApplicationAddonName:     false,
	CertPolicyAddonName:      true,
	IamPolicyAddonName:       true,
	PolicyAddonName:          true,
	ConfigPolicyAddonName:    true,
	PolicyFrameworkAddonName: true,
	SearchAddonName:          false,
}

// KlusterletAddonImageNames is the image key names for each addon agents in image-manifest configmap
var KlusterletAddonImageNames = map[string][]string{
	WorkManagerAddonName: []string{"multicloud_manager"},
	ApplicationAddonName: []string{"multicluster_operators_subscription"},
	CertPolicyAddonName:  []string{"cert_policy_controller"},
	IamPolicyAddonName:   []string{"iam_policy_controller"},
	PolicyAddonName: []string{"config_policy_controller", "governance_policy_spec_sync",
		"governance_policy_status_sync", "governance_policy_template_sync"},
	ConfigPolicyAddonName: []string{"config_policy_controller"},
	PolicyFrameworkAddonName: []string{"governance_policy_spec_sync", "governance_policy_status_sync",
		"governance_policy_template_sync"},
	SearchAddonName: []string{"search_collector"},
}
