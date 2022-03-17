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
	ConfigPolicyAddonName    = "config-policy-controller"
	IamPolicyAddonName       = "iam-policy-controller"
	PolicyAddonName          = "policy-controller"
	PolicyFrameworkAddonName = "governance-policy-framework"
	SearchAddonName          = "search-collector"
)

// KlusterletAddons is a list of managedClusterAddons which can be updated by addon-controller.
// true means it is deployed by addon-controller, can be updated and deleted.
// false means it is not deployed by addon-controller, only can be updated, but cannot be deleted.
var KlusterletAddons = map[string]bool{
	ApplicationAddonName:     true,
	ConfigPolicyAddonName:    true,
	CertPolicyAddonName:      true,
	IamPolicyAddonName:       true,
	PolicyAddonName:          true,
	PolicyFrameworkAddonName: true,
	SearchAddonName:          true,
}

// KlusterletAddonImageNames is the image key names for each addon agents in image-manifest configmap
var KlusterletAddonImageNames = map[string][]string{
	ApplicationAddonName:  []string{"multicluster_operators_subscription"},
	ConfigPolicyAddonName: []string{"config_policy_controller"},
	CertPolicyAddonName:   []string{"cert_policy_controller"},
	IamPolicyAddonName:    []string{"iam_policy_controller"},
	PolicyAddonName: []string{"config_policy_controller", "governance_policy_spec_sync",
		"governance_policy_status_sync", "governance_policy_template_sync"},
	PolicyFrameworkAddonName: []string{"governance_policy_spec_sync", "governance_policy_status_sync",
		"governance_policy_template_sync"},
	SearchAddonName: []string{"search_collector"},
}

// ClusterManagementAddons is a list of ClusterManagementAddons need to delete during the upgrade from 2.4 to 2.5
var ClusterManagementAddons = []string{
	ApplicationAddonName,
	CertPolicyAddonName,
	IamPolicyAddonName,
	PolicyAddonName,
	SearchAddonName,
}

var DeprecatedManagedClusterAddons = []string{
	PolicyAddonName,
}

var KlusterletAddonComponentNames = map[string]string{
	WorkManagerAddonName: "workmgr",
	ApplicationAddonName: "appmgr",
	CertPolicyAddonName:  "certpolicyctrl",
	IamPolicyAddonName:   "iampolicyctrl",
	PolicyAddonName:      "policyctrl",
	SearchAddonName:      "search",
}

var DeprecatedAgentManifestworks = []string{
	"klusterlet-addon-appmgr",
	"klusterlet-addon-certpolicyctrl",
	"klusterlet-addon-iampolicyctrl",
	"klusterlet-addon-policyctrl",
	"klusterlet-addon-workmgr",
	"klusterlet-addon-search",
}
