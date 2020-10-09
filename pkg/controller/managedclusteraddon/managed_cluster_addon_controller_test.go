// Copyright (c) 2020 Red Hat, Inc.
package managedclusteraddon

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
	"gotest.tools/assert"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func Test_setStatusCondition(t *testing.T) {
	oldTime := metav1.NewTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	newTime := metav1.NewTime(time.Now())
	conditionAvailableTrue := metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "Available",
		Message:            "Available True",
	}
	conditionAvailableFalse := metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: newTime,
		Reason:             "NotAvailable",
		Message:            "Available False",
	}
	conditionProgressingTrue1 := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "Installing",
		Message:            "Progressing Msg",
	}
	conditionProgressingTrue2 := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: newTime,
		Reason:             "Uninstalling",
		Message:            "Progressing Msg",
	}
	conditionProgressingFalse1 := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: oldTime,
		Reason:             "Done",
		Message:            "Done1",
	}
	conditionProgressingFalse2 := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: newTime,
		Reason:             "Done",
		Message:            "Done2",
	}
	conditionDegradedTrue := metav1.Condition{
		Type:               "Degraded",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "Degraded",
		Message:            "Degraded",
	}
	type args struct {
		conditions *[]metav1.Condition
		condition  *metav1.Condition
	}
	type testcase struct {
		name        string
		args        args
		wantInclude metav1.Condition
	}

	tests := []testcase{
		{
			name: "same type & status & reason, should do nothing",
			args: args{
				conditions: &[]metav1.Condition{
					conditionDegradedTrue,
					conditionProgressingFalse1,
					conditionAvailableFalse,
				},
				condition: &conditionProgressingFalse2,
			},
			wantInclude: conditionProgressingFalse1,
		},
		{
			name: "empty conditions, should add",
			args: args{
				conditions: &[]metav1.Condition{
					conditionProgressingFalse1,
					conditionAvailableTrue,
				},
				condition: &conditionDegradedTrue,
			},
			wantInclude: conditionDegradedTrue,
		},
		{
			name: "no same type conditions, should add",
			args: args{
				conditions: &[]metav1.Condition{},
				condition:  &conditionDegradedTrue,
			},
			wantInclude: conditionDegradedTrue,
		},
		{
			name: "different status, should update",
			args: args{
				conditions: &[]metav1.Condition{
					conditionDegradedTrue,
					conditionProgressingFalse1,
					conditionAvailableTrue,
				},
				condition: &conditionAvailableFalse,
			},
			wantInclude: conditionAvailableFalse,
		},
		{
			name: "different reason, should update",
			args: args{
				conditions: &[]metav1.Condition{
					conditionDegradedTrue,
					conditionProgressingTrue1,
					conditionAvailableTrue,
				},
				condition: &conditionProgressingTrue2,
			},
			wantInclude: conditionProgressingTrue2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setStatusCondition(tt.args.conditions, tt.args.condition)
			hasEqual := false
			// check has include ones
			for _, c := range *tt.args.conditions {
				if reflect.DeepEqual(c, tt.wantInclude) {
					hasEqual = true
					break
				}
			}
			if !hasEqual {
				t.Errorf("setStatusCondition() error expect %v to include %v", tt.args.conditions, tt.wantInclude)
			}
		})
	}
}
func Test_checkManifestWorkStatus(t *testing.T) {
	// func checkManifestWorkStatus(m *manifestworkv1.ManifestWork) (numFailed, numSucceeded, numTotal int)
	manifestConditionSucceed1 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 0,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest complete",
				Reason:             "AppliedManifestComplete",
				Status:             metav1.ConditionTrue,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}
	manifestConditionFailed1 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 0,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest failed",
				Reason:             "AppliedManifestFailed",
				Status:             metav1.ConditionFalse,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}
	manifestConditionSucceed2 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 1,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest complete",
				Reason:             "AppliedManifestComplete",
				Status:             metav1.ConditionTrue,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}
	manifestConditionFailed2 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 1,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest failed",
				Reason:             "AppliedManifestFailed",
				Status:             metav1.ConditionFalse,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}

	secret1 := &corev1.Secret{
		Data: map[string][]byte{
			"kubeconfig": []byte("kubeConfigData1"),
		},
	}
	secret2 := &corev1.Secret{
		Data: map[string][]byte{
			"kubeconfig": []byte("kubeConfigData2"),
		},
	}
	secret3 := &corev1.Secret{
		Data: map[string][]byte{
			"kubeconfig": []byte("kubeConfigData3"),
		},
	}
	type testcase struct {
		name          string
		args          *manifestworkv1.ManifestWork
		wantFailed    int
		wantSucceeded int
		wantTotal     int
	}
	tests := []testcase{
		{
			name:          "empty manifestwork",
			args:          nil,
			wantFailed:    0,
			wantSucceeded: 0,
			wantTotal:     0,
		},
		{
			name: "empty status",
			args: &manifestworkv1.ManifestWork{
				Spec: manifestworkv1.ManifestWorkSpec{
					Workload: manifestworkv1.ManifestsTemplate{
						Manifests: []manifestworkv1.Manifest{
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret3}},
						},
					},
				},
			},
			wantFailed:    0,
			wantSucceeded: 0,
			wantTotal:     3,
		},
		{
			name: "in progress",
			args: &manifestworkv1.ManifestWork{
				Spec: manifestworkv1.ManifestWorkSpec{
					Workload: manifestworkv1.ManifestsTemplate{
						Manifests: []manifestworkv1.Manifest{
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret3}},
						},
					},
				},
				Status: manifestworkv1.ManifestWorkStatus{
					ResourceStatus: manifestworkv1.ManifestResourceStatus{
						Manifests: []manifestworkv1.ManifestCondition{
							manifestConditionSucceed1,
							manifestConditionSucceed2,
						},
					},
				},
			},
			wantFailed:    0,
			wantSucceeded: 2,
			wantTotal:     3,
		},
		{
			name: "all succeeded",
			args: &manifestworkv1.ManifestWork{
				Spec: manifestworkv1.ManifestWorkSpec{
					Workload: manifestworkv1.ManifestsTemplate{
						Manifests: []manifestworkv1.Manifest{
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
						},
					},
				},
				Status: manifestworkv1.ManifestWorkStatus{
					ResourceStatus: manifestworkv1.ManifestResourceStatus{
						Manifests: []manifestworkv1.ManifestCondition{
							manifestConditionSucceed1,
							manifestConditionSucceed2,
						},
					},
				},
			},
			wantFailed:    0,
			wantSucceeded: 2,
			wantTotal:     2,
		},
		{
			name: "has failure",
			args: &manifestworkv1.ManifestWork{
				Spec: manifestworkv1.ManifestWorkSpec{
					Workload: manifestworkv1.ManifestsTemplate{
						Manifests: []manifestworkv1.Manifest{
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret3}},
						},
					},
				},
				Status: manifestworkv1.ManifestWorkStatus{
					ResourceStatus: manifestworkv1.ManifestResourceStatus{
						Manifests: []manifestworkv1.ManifestCondition{
							manifestConditionFailed1,
							manifestConditionFailed2,
						},
					},
				},
			},
			wantFailed:    2,
			wantSucceeded: 0,
			wantTotal:     3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numFailed, numSucceeded, numTotal := checkManifestWorkStatus(tt.args)
			assert.Equal(t, tt.wantFailed, numFailed)
			assert.Equal(t, tt.wantSucceeded, numSucceeded)
			assert.Equal(t, tt.wantTotal, numTotal)
		})
	}
}

func Test_updateAvailableStatus(t *testing.T) {
	oldTime := metav1.NewMicroTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	newTime := metav1.NewMicroTime(time.Now())
	mca := &addonv1alpha1.ManagedClusterAddOn{}
	type args struct {
		mca      *addonv1alpha1.ManagedClusterAddOn
		notFound bool
		lease    *coordinationv1.Lease
	}
	type testcase struct {
		name       string
		args       args
		wantStatus metav1.ConditionStatus
		wantErr    bool
	}
	seconds := int32(60)
	tests := []testcase{
		{
			name: "lease not found",
			args: args{
				mca:      mca,
				notFound: true,
				lease:    &coordinationv1.Lease{},
			},
			wantStatus: metav1.ConditionFalse,
			wantErr:    false,
		},
		{
			name: "lease empty",
			args: args{
				mca:      mca,
				notFound: false,
				lease: &coordinationv1.Lease{
					Spec: coordinationv1.LeaseSpec{
						LeaseDurationSeconds: nil,
						RenewTime:            nil,
					},
				},
			},
			wantStatus: metav1.ConditionFalse,
			wantErr:    true,
		},
		{
			name: "lease found",
			args: args{
				mca:      mca,
				notFound: false,
				lease: &coordinationv1.Lease{
					Spec: coordinationv1.LeaseSpec{
						LeaseDurationSeconds: &seconds,
						RenewTime:            &newTime,
					},
				},
			},
			wantStatus: metav1.ConditionTrue,
			wantErr:    false,
		},
		{
			name: "lease expires",
			args: args{
				mca:      mca,
				notFound: false,
				lease: &coordinationv1.Lease{
					Spec: coordinationv1.LeaseSpec{
						RenewTime: &oldTime,
					},
				},
			},
			wantStatus: metav1.ConditionUnknown,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := updateAvailableStatus(tt.args.mca, tt.args.notFound, tt.args.lease)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateAvailableStatus() error expected %t returned %v", tt.wantErr, err)
			}
			if s != tt.wantStatus {
				t.Errorf("updateAvailableStatus() error expected %v returned %v", tt.wantStatus, s)
			}
			hasStatus := false
			for _, c := range tt.args.mca.Status.Conditions {
				if c.Type == "Available" && c.Status == tt.wantStatus {
					hasStatus = true
				}
			}
			if !hasStatus {
				t.Errorf("updateAvailableStatus() error expect include Available=%v", tt.wantStatus)
			}
		})
	}
}

func Test_updateProcessingStatus(t *testing.T) {
	manifestConditionSucceed1 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 0,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest complete",
				Reason:             "AppliedManifestComplete",
				Status:             metav1.ConditionTrue,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}
	manifestConditionFailed1 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 0,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest failed",
				Reason:             "AppliedManifestFailed",
				Status:             metav1.ConditionFalse,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}
	manifestConditionSucceed2 := manifestworkv1.ManifestCondition{
		ResourceMeta: manifestworkv1.ManifestResourceMeta{
			Ordinal: 1,
		},
		Conditions: []metav1.Condition{
			metav1.Condition{
				LastTransitionTime: metav1.NewTime(time.Now()),
				Message:            "Apply manifest complete",
				Reason:             "AppliedManifestComplete",
				Status:             metav1.ConditionTrue,
				Type:               string(manifestworkv1.ManifestApplied),
			},
		},
	}

	secret1 := &corev1.Secret{
		Data: map[string][]byte{
			"kubeconfig": []byte("kubeConfigData1"),
		},
	}
	secret2 := &corev1.Secret{
		Data: map[string][]byte{
			"kubeconfig": []byte("kubeConfigData2"),
		},
	}

	type args struct {
		mca       *addonv1alpha1.ManagedClusterAddOn
		isEnabled bool
		notFound  bool
		mw        *manifestworkv1.ManifestWork
	}
	type testcase struct {
		name       string
		args       args
		wantStatus metav1.ConditionStatus
		wantReason string
		wantErr    bool
	}
	tests := []testcase{
		{
			name: "enabled & not found",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: true,
				notFound:  true,
				mw:        &manifestworkv1.ManifestWork{},
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: processingReasonMissing,
			wantErr:    false,
		},
		{
			name: "disabled & not found",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: false,
				notFound:  true,
				mw:        &manifestworkv1.ManifestWork{},
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: processingReasonDeleting,
			wantErr:    false,
		},
		{
			name: "disabled & found",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: false,
				notFound:  false,
				mw: &manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
				},
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: processingReasonDeleting,
			wantErr:    false,
		},
		{
			name: "enabled & partially finished",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: true,
				notFound:  false,
				mw: &manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
					Status: manifestworkv1.ManifestWorkStatus{
						ResourceStatus: manifestworkv1.ManifestResourceStatus{
							Manifests: []manifestworkv1.ManifestCondition{
								manifestConditionSucceed1,
							},
						},
					},
				},
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: processingReasonCreated,
			wantErr:    false,
		},
		{
			name: "enabled & all finished",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: true,
				notFound:  false,
				mw: &manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
					Status: manifestworkv1.ManifestWorkStatus{
						ResourceStatus: manifestworkv1.ManifestResourceStatus{
							Manifests: []manifestworkv1.ManifestCondition{
								manifestConditionSucceed1,
								manifestConditionSucceed2,
							},
						},
					},
				},
			},
			wantStatus: metav1.ConditionFalse,
			wantReason: processingReasonApplied,
			wantErr:    false,
		},
		{
			name: "enabled & has error",
			args: args{
				mca:       &addonv1alpha1.ManagedClusterAddOn{},
				isEnabled: true,
				notFound:  false,
				mw: &manifestworkv1.ManifestWork{
					Spec: manifestworkv1.ManifestWorkSpec{
						Workload: manifestworkv1.ManifestsTemplate{
							Manifests: []manifestworkv1.Manifest{
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret1}},
								manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: secret2}},
							},
						},
					},
					Status: manifestworkv1.ManifestWorkStatus{
						ResourceStatus: manifestworkv1.ManifestResourceStatus{
							Manifests: []manifestworkv1.ManifestCondition{
								manifestConditionFailed1,
								manifestConditionSucceed2,
							},
						},
					},
				},
			},
			wantStatus: metav1.ConditionFalse,
			wantReason: processingReasonApplied,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := updateProcessingStatus(tt.args.mca, tt.args.isEnabled, tt.args.notFound, tt.args.mw)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateProcessingStatus() error expected %t returned %v", tt.wantErr, err)
			}
			if s != tt.wantStatus {
				t.Errorf("updateProcessingStatus() error expected %v returned %v", tt.wantStatus, s)
			}
			hasStatus := false
			for _, c := range tt.args.mca.Status.Conditions {
				if c.Type == "Progressing" && c.Status == tt.wantStatus && c.Reason == tt.wantReason {
					hasStatus = true
				}
			}
			if !hasStatus {
				t.Errorf("updateProcessingStatus() error expect %v to include Progressing=%v:%s",
					tt.args.mca.Status.Conditions, tt.wantStatus, tt.wantReason)
			}
		})
	}
}

//func (r *ReconcileManagedClusterAddOn) Reconcile(request reconcile.Request) (reconcile.Result, error)
func Test_Reconcile(t *testing.T) {
	// give lease check requeue time
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})
	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(coordinationv1.SchemeGroupVersion, &coordinationv1.Lease{})

	testKlusterletAddonConfig := &agentv1.KlusterletAddonConfig{
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
			Version: "2.0.0",
		},
	}
	testManagedClusterAddOn := &addonv1alpha1.ManagedClusterAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ManagedClusterAddOn",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
		Status: addonv1alpha1.ManagedClusterAddOnStatus{
			RelatedObjects: []addonv1alpha1.ObjectReference{
				addonv1alpha1.ObjectReference{
					Group:    "agent.open-cluster-management.io",
					Resource: "klusterletaddonconfigs",
					Name:     "test-managedcluster",
				},
			},
		},
	}
	testManagedClusterAddOnNoRef := &addonv1alpha1.ManagedClusterAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ManagedClusterAddOn",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
	}

	duration := int32(60)
	duration120 := int32(60)
	time1 := metav1.NewMicroTime(time.Now().Add(-time.Second * 290))
	time2 := metav1.NewMicroTime(time.Now().Add(-time.Second * 240))
	time3 := metav1.NewMicroTime(time.Now())
	// should requeue after 30 seconds
	testLease1 := &coordinationv1.Lease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Lease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
		Spec: coordinationv1.LeaseSpec{
			LeaseDurationSeconds: &duration,
			RenewTime:            &time1,
		},
	}
	// requeue after 60 seconds
	testLease2 := &coordinationv1.Lease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Lease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
		Spec: coordinationv1.LeaseSpec{
			LeaseDurationSeconds: &duration,
			RenewTime:            &time2,
		},
	}
	// requeue after
	testLease3 := &coordinationv1.Lease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Lease",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
		Spec: coordinationv1.LeaseSpec{
			LeaseDurationSeconds: &duration120,
			RenewTime:            &time3,
		},
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "application-manager",
			Namespace: "test-managedcluster",
		},
	}

	type fields struct {
		client client.Client
		scheme *runtime.Scheme
	}

	type args struct {
		request reconcile.Request
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reconcile.Result
		wantErr bool
	}{
		{
			name: "managedclusteraddon does not exist",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue: false,
			},
			wantErr: false,
		},
		{
			name: "managedclusteraddon missing klusterletaddonconfig ref",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme, testManagedClusterAddOnNoRef),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue: false,
			},
			wantErr: false,
		},
		{
			name: "managedclusteraddon exists & lease will expire in 10 seconds",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme,
					testManagedClusterAddOn, testKlusterletAddonConfig, testLease1),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue:      true,
				RequeueAfter: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "managedclusteraddon & lease will expire in 60 seconds",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme,
					testManagedClusterAddOn, testKlusterletAddonConfig, testLease2),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue:      true,
				RequeueAfter: 60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "managedclusteraddon & lease up to date",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme,
					testManagedClusterAddOn, testKlusterletAddonConfig, testLease3),
				scheme: testscheme,
			},
			args: args{
				request: req,
			},
			want: reconcile.Result{
				Requeue:      true,
				RequeueAfter: 60 * 5 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.scheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})
			r := &ReconcileManagedClusterAddOn{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}

			got, err := r.Reconcile(tt.args.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileManagedClusterAddOn.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// allow requeueafter to be 2 seconds not accurate
			if got.Requeue != tt.want.Requeue ||
				got.RequeueAfter-tt.want.RequeueAfter > 2*time.Second ||
				got.RequeueAfter-tt.want.RequeueAfter < -2*time.Second {
				t.Errorf("ReconcileManagedClusterAddOn.Reconcile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_updateDegradedStatus(t *testing.T) {
	type args struct {
		mca           *addonv1alpha1.ManagedClusterAddOn
		errProcessing error
		errAvailable  error
	}
	type testcase struct {
		name       string
		args       args
		wantStatus metav1.ConditionStatus
		wantReason string
		wantType   bool
	}
	oldTime := metav1.NewTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	conditionDegradedTrue := metav1.Condition{
		Type:               "Degraded",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "Degraded",
		Message:            "Degraded",
	}
	conditionAvailableFalse := metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: oldTime,
		Reason:             "NotAvailable",
		Message:            "Available False",
	}
	tests := []testcase{
		{
			name: "no error",
			args: args{
				mca:           &addonv1alpha1.ManagedClusterAddOn{},
				errProcessing: nil,
				errAvailable:  nil,
			},
			wantType:   false,
			wantStatus: metav1.ConditionTrue,
			wantReason: "",
		},
		{
			name: "no error should clean up previous errors",
			args: args{
				mca: &addonv1alpha1.ManagedClusterAddOn{
					Status: addonv1alpha1.ManagedClusterAddOnStatus{
						Conditions: []metav1.Condition{
							conditionDegradedTrue,
							conditionAvailableFalse,
						},
					},
				},
				errProcessing: nil,
				errAvailable:  nil,
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: "",
			wantType:   false,
		},
		{
			name: "has processing error",
			args: args{
				mca:           &addonv1alpha1.ManagedClusterAddOn{},
				errProcessing: fmt.Errorf("failed install"),
				errAvailable:  fmt.Errorf("timeout"),
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: degradedReasonInstallError,
			wantType:   true,
		},
		{
			name: "has available error",
			args: args{
				mca: &addonv1alpha1.ManagedClusterAddOn{
					Status: addonv1alpha1.ManagedClusterAddOnStatus{
						Conditions: []metav1.Condition{
							conditionDegradedTrue,
							conditionAvailableFalse,
						},
					},
				},
				errProcessing: nil,
				errAvailable:  fmt.Errorf("timeout"),
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: degradedReasonTimeout,
			wantType:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateDegradedStatus(tt.args.mca, tt.args.errProcessing, tt.args.errAvailable)
			hasStatus := false
			for _, c := range tt.args.mca.Status.Conditions {
				if c.Type == "Degraded" && !tt.wantType {
					t.Errorf("updateProcessingStatus() error expect %v to not include Degraded", tt.args.mca.Status.Conditions)
					return
				}
				if c.Type == "Degraded" && c.Status == tt.wantStatus && c.Reason == tt.wantReason {
					hasStatus = true
				}
			}

			if tt.wantType && !hasStatus {
				t.Errorf("updateProcessingStatus() error expect %v to include Degraded=%v:%s",
					tt.args.mca.Status.Conditions, tt.wantStatus, tt.wantReason)
			}
		})
	}
}

func Test_checkInstallTimeout(t *testing.T) {
	oldTime := metav1.NewTime(time.Now().Add(-310 * time.Second))
	newTime := metav1.NewTime(time.Now())
	conditionProgressingFalse := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: oldTime,
		Reason:             "Progressing",
		Message:            "Progressing",
	}
	conditionProgressingFalseNew := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: newTime,
		Reason:             "Progressing",
		Message:            "Progressing",
	}
	conditionAvailableFalse := metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionFalse,
		LastTransitionTime: oldTime,
		Reason:             "NotAvailable",
		Message:            "Available False",
	}
	conditionProgressingTrue := metav1.Condition{
		Type:               "Progressing",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "Progressing",
		Message:            "Progressing",
	}
	conditionAvailableTrue := metav1.Condition{
		Type:               "Available",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: oldTime,
		Reason:             "NotAvailable",
		Message:            "Available True",
	}
	type testcase struct {
		name    string
		arg     *addonv1alpha1.ManagedClusterAddOn
		wantErr bool
	}
	tests := []testcase{
		testcase{
			name: "is available",
			arg: &addonv1alpha1.ManagedClusterAddOn{
				Status: addonv1alpha1.ManagedClusterAddOnStatus{
					Conditions: []metav1.Condition{
						conditionProgressingFalse,
						conditionAvailableTrue,
					},
				},
			},
			wantErr: false,
		},
		testcase{
			name: "is progressing",
			arg: &addonv1alpha1.ManagedClusterAddOn{
				Status: addonv1alpha1.ManagedClusterAddOnStatus{
					Conditions: []metav1.Condition{
						conditionProgressingTrue,
					},
				},
			},
			wantErr: false,
		},
		testcase{
			name: "is still on time",
			arg: &addonv1alpha1.ManagedClusterAddOn{
				Status: addonv1alpha1.ManagedClusterAddOnStatus{
					Conditions: []metav1.Condition{
						conditionProgressingFalseNew,
						conditionAvailableFalse,
					},
				},
			},
			wantErr: false,
		},
		testcase{
			name: "is timeout",
			arg: &addonv1alpha1.ManagedClusterAddOn{
				Status: addonv1alpha1.ManagedClusterAddOnStatus{
					Conditions: []metav1.Condition{
						conditionProgressingFalse,
						conditionAvailableFalse,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := checkInstallTimeout(tt.arg)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("checkInstallTimeout() returns error %v, which is not expected", gotErr)
			}
		})
	}

}
