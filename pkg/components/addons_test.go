package components

import "testing"

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
