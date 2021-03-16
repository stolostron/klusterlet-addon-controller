// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package components

import (
	"os"
	"testing"
)

func TestGetAddonFromManagedClusterAddonName(t *testing.T) {
	//func GetAddonFromManagedClusterAddonName(name string) (KlusterletAddon, error)
	tests := []struct {
		name          string
		arg           string
		wantAddonName string
		wantErr       bool
	}{
		{
			name:          "success appmgr",
			arg:           "application-manager",
			wantAddonName: "appmgr",
			wantErr:       false,
		},
		{
			name:          "success policyctrl",
			arg:           "policy-controller",
			wantAddonName: "policyctrl",
			wantErr:       false,
		},
		{
			name:          "success iampolicyctrl",
			arg:           "iam-policy-controller",
			wantAddonName: "iampolicyctrl",
			wantErr:       false,
		},
		{
			name:          "success certpolicyctrl",
			arg:           "cert-policy-controller",
			wantAddonName: "certpolicyctrl",
			wantErr:       false,
		},
		{
			name:          "success search",
			arg:           "search-collector",
			wantAddonName: "search",
			wantErr:       false,
		},
		{
			name:          "success workmgr",
			arg:           "work-manager",
			wantAddonName: "workmgr",
			wantErr:       false,
		},
		{
			name:          "failed policyctrl",
			arg:           "some-test-policy-controller",
			wantAddonName: "policyctrl",
			wantErr:       true,
		},
		{
			name:          "failed empty name",
			arg:           "",
			wantAddonName: "",
			wantErr:       true,
		},
		{
			name:          "failed random name",
			arg:           "a-unique-random-name",
			wantAddonName: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addon, err := GetAddonFromManagedClusterAddonName(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddonFromManagedClusterAddonName() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil && tt.wantErr == false && (addon == nil || tt.wantAddonName != addon.GetAddonName()) {
				t.Errorf("GetAddonFromManagedClusterAddonName() addonName = %v, want %v", addon.GetAddonName(), tt.wantAddonName)
				return
			}
		})
	}
}
func TestGetAddonFromManifestWorkName(t *testing.T) {
	// GetAddonFromManifestWorkName(manifestworkName string) (KlusterletAddon, error)
	tests := []struct {
		name          string
		arg           string
		wantAddonName string
		wantErr       bool
	}{
		{
			name:          "success appmgr",
			arg:           "some-random-ns-klusterlet-addon-appmgr",
			wantAddonName: "appmgr",
			wantErr:       false,
		},
		{
			name:          "success policyctrl",
			arg:           "anamespace-klusterlet-addon-policyctrl",
			wantAddonName: "policyctrl",
			wantErr:       false,
		},
		{
			name:          "success iampolicyctrl",
			arg:           "a.name-space:weirdformat-klusterlet-addon-iampolicyctrl",
			wantAddonName: "iampolicyctrl",
			wantErr:       false,
		},
		{
			name:          "success certpolicyctrl",
			arg:           "-klusterlet-addon-certpolicyctrl",
			wantAddonName: "certpolicyctrl",
			wantErr:       false,
		},
		{
			name:          "success search",
			arg:           "test-klusterlet-addon-search",
			wantAddonName: "search",
			wantErr:       false,
		},
		{
			name:          "success workmgr",
			arg:           "test-klusterlet-addon-workmgr",
			wantAddonName: "workmgr",
			wantErr:       false,
		},
		{
			name:          "failed policyctrl",
			arg:           "some-test-klusterlet-addon-policy-controller",
			wantAddonName: "policyctrl",
			wantErr:       true,
		},
		{
			name:          "failed empty name",
			arg:           "",
			wantAddonName: "",
			wantErr:       true,
		},
		{
			name:          "failed random name",
			arg:           "a-unique-random-name",
			wantAddonName: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addon, err := GetAddonFromManifestWorkName(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAddonFromManifestWorkName() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil && tt.wantErr == false && (addon == nil || tt.wantAddonName != addon.GetAddonName()) {
				name := ""
				if addon != nil {
					addon.GetAddonName()
				}
				t.Errorf("GetAddonFromManifestWorkName() addonName = %v, want %v", name, tt.wantAddonName)
				return
			}
		})
	}
}

func TestGetManagedClusterAddOnName(t *testing.T) {
	tests := []struct {
		name    string
		addon   KlusterletAddon
		envName string
		envVal  string
		want    string
	}{
		{
			name:    "appmgr with env",
			addon:   AppMgr,
			envName: "APPMGR_NAME",
			envVal:  "diff-Appmgr",
			want:    "diff-Appmgr",
		},
		{
			name:    "appmgr without env",
			addon:   AppMgr,
			envName: "",
			envVal:  "",
			want:    "application-manager",
		},
		{
			name:    "certpolicymgr without env",
			addon:   CertCtrl,
			envName: "",
			envVal:  "",
			want:    "cert-policy-controller",
		},
		{
			name:    "certpolicymgr with env",
			addon:   CertCtrl,
			envName: "CERTPOLICYCTRL_NAME",
			envVal:  "diff-cert",
			want:    "diff-cert",
		},
		{
			name:    "iampolicyctrl with env",
			addon:   IAMCtrl,
			envName: "IAMPOLICYCTRL_NAME",
			envVal:  "diff-iam",
			want:    "diff-iam",
		},
		{
			name:    "policyctrl with env",
			addon:   PolicyCtrl,
			envName: "POLICYCTRL_NAME",
			envVal:  "diff-policy",
			want:    "diff-policy",
		},
		{
			name:    "search with env",
			addon:   Search,
			envName: "SEARCH_NAME",
			envVal:  "diff-search",
			want:    "diff-search",
		},
		{
			name:    "workmgr with env",
			addon:   WorkMgr,
			envName: "WORKMGR_NAME",
			envVal:  "diff-work",
			want:    "diff-work",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envName != "" && tt.envVal != "" {
				os.Setenv(tt.envName, tt.envVal)
				defer os.Unsetenv(tt.envName)
			}
			got := tt.addon.GetManagedClusterAddOnName()
			if got != tt.want {
				t.Errorf("GetManagedClusterAddOnName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
