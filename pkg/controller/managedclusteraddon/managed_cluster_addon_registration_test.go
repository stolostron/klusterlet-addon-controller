// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedclusteraddon

import (
	"context"
	"testing"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	addons "github.com/stolostron/klusterlet-addon-controller/pkg/components"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func Test_deleteOutDatedRoleRoleBinding(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})

	klusterletaddonconfig1 := &agentv1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster",
			Namespace: "test-managedcluster",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonConfigApplicationManagerSpec{
				Enabled: true,
			},
		},
	}
	klusterletaddonconfig2 := &agentv1.KlusterletAddonConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "KlusterletAddonConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-2",
			Namespace: "test-managedcluster",
		},
		Spec: agentv1.KlusterletAddonConfigSpec{
			ApplicationManagerConfig: agentv1.KlusterletAddonConfigApplicationManagerSpec{
				Enabled: true,
			},
		},
	}
	roleOwned := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}

	rolebindingOwned := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}

	rolebindingOwnedOther := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}

	roleNotOwned := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}
	rolebindingNotOwned := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}
	roleOwnedOther := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-managedcluster-appmgr",
			Namespace: "test-managedcluster",
		},
	}

	roleNotRelated1 := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-role-1",
			Namespace: "test-managedcluster",
		},
	}

	roleNotRelated2 := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-role-2",
			Namespace: "test-managedcluster",
		},
	}
	rolebindingNotRelated := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "some-rolebinding",
			Namespace: "test-managedcluster",
		},
	}
	if err := controllerutil.SetControllerReference(klusterletaddonconfig1, roleOwned, testscheme); err != nil {
		t.Errorf("unexpected error when setting controller reference: %v", err)
	}
	if err := controllerutil.SetControllerReference(klusterletaddonconfig1, rolebindingOwned, testscheme); err != nil {
		t.Errorf("unexpected error when setting controller reference: %v", err)
	}
	if err := controllerutil.SetControllerReference(klusterletaddonconfig2, rolebindingOwnedOther, testscheme); err != nil {
		t.Errorf("unexpected error when setting controller reference: %v", err)
	}
	if err := controllerutil.SetControllerReference(klusterletaddonconfig2, roleOwnedOther, testscheme); err != nil {
		t.Errorf("unexpected error when setting controller reference: %v", err)
	}
	if err := controllerutil.SetControllerReference(klusterletaddonconfig1, roleNotRelated1, testscheme); err != nil {
		t.Errorf("unexpected error when setting controller reference: %v", err)
	}

	type args struct {
		client                client.Client
		addon                 addons.KlusterletAddon
		klusterletaddonconfig *agentv1.KlusterletAddonConfig
	}

	tests := []struct {
		name               string
		args               args
		numRoleLeft        int
		numRolebindingLeft int
		wantErr            bool
	}{
		{
			name: "role should be deleted",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, roleOwned, klusterletaddonconfig1),
			},
			numRoleLeft:        0,
			numRolebindingLeft: 0,
			wantErr:            false,
		},
		{
			name: "rolebinding should be deleted",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, rolebindingOwned, klusterletaddonconfig1),
			},
			numRoleLeft:        0,
			numRolebindingLeft: 0,
			wantErr:            false,
		},
		{
			name: "both role & rolebinding should be deleted",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, roleOwned, rolebindingOwned, klusterletaddonconfig1),
			},
			numRoleLeft:        0,
			numRolebindingLeft: 0,
			wantErr:            false,
		},
		{
			name: "no owner will be ignored",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, roleNotOwned, rolebindingNotOwned, klusterletaddonconfig1),
			},
			numRoleLeft:        1,
			numRolebindingLeft: 1,
			wantErr:            false,
		},
		{
			name: "not owned by current will be ignored",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, roleOwnedOther, rolebindingOwnedOther, klusterletaddonconfig1),
			},
			numRoleLeft:        1,
			numRolebindingLeft: 1,
			wantErr:            false,
		},
		{
			name: "not found will be ignored",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client:                fake.NewFakeClientWithScheme(testscheme, klusterletaddonconfig1),
			},
			numRoleLeft:        0,
			numRolebindingLeft: 0,
			wantErr:            false,
		},
		{
			name: "not related role/rolebindings will not be removed",
			args: args{
				klusterletaddonconfig: klusterletaddonconfig1,
				addon:                 addons.AppMgr,
				client: fake.NewFakeClientWithScheme(
					testscheme,
					klusterletaddonconfig1,
					roleNotRelated1,
					roleNotRelated2,
					rolebindingNotRelated,
				),
			},
			numRoleLeft:        2,
			numRolebindingLeft: 1,
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := deleteOutDatedRoleRoleBinding(tt.args.addon, tt.args.klusterletaddonconfig, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("deleteOutDatedRoleRoleBinding() get error %v, wantErr %t", err, tt.wantErr)
			}
			// check num of roles
			roleList := &rbacv1.RoleList{}
			if err := tt.args.client.List(context.TODO(), roleList); err != nil {
				t.Errorf("unexpected error when list roles: %v", err)
			} else if len(roleList.Items) != tt.numRoleLeft {
				t.Errorf("deleteOutDatedRoleRoleBinding() get wrong # of roles left %d, want %d",
					len(roleList.Items), tt.numRoleLeft)
			}

			// check num of rolebindings
			rolebindingList := &rbacv1.RoleBindingList{}
			if err := tt.args.client.List(context.TODO(), rolebindingList); err != nil {
				t.Errorf("unexpected error when list roles: %v", err)
			} else if len(rolebindingList.Items) != tt.numRolebindingLeft {
				t.Errorf("deleteOutDatedRoleRoleBinding() get wrong # of roles left %d, want %d",
					len(rolebindingList.Items), tt.numRolebindingLeft)
			}
		})
	}

}
