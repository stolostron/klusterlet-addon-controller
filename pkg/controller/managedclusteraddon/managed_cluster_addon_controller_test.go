// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedclusteraddon

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"gotest.tools/assert"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
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

func Test_updateProgressingStatus(t *testing.T) {
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
			wantReason: progressingReasonMissing,
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
			wantReason: progressingReasonDeleting,
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
			wantReason: progressingReasonDeleting,
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
			wantReason: progressingReasonCreated,
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
			wantReason: progressingReasonApplied,
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
			wantReason: progressingReasonApplied,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := updateProgressingStatus(tt.args.mca, tt.args.isEnabled, tt.args.notFound, tt.args.mw)
			if (err != nil) != tt.wantErr {
				t.Errorf("updateProgressingStatus() error expected %t returned %v", tt.wantErr, err)
			}
			if s != tt.wantStatus {
				t.Errorf("updateProgressingStatus() error expected %v returned %v", tt.wantStatus, s)
			}
			hasStatus := false
			for _, c := range tt.args.mca.Status.Conditions {
				if c.Type == "Progressing" && c.Status == tt.wantStatus && c.Reason == tt.wantReason {
					hasStatus = true
				}
			}
			if !hasStatus {
				t.Errorf("updateProgressingStatus() error expect %v to include Progressing=%v:%s",
					tt.args.mca.Status.Conditions, tt.wantStatus, tt.wantReason)
			}
		})
	}
}

// func (r *ReconcileManagedClusterAddOn) Reconcile(request reconcile.Request) (reconcile.Result, error)
func Test_Reconcile(t *testing.T) {
	// give lease check requeue time
	testscheme := scheme.Scheme
	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ManagedClusterAddOn{})
	testscheme.AddKnownTypes(agentv1.SchemeGroupVersion, &agentv1.KlusterletAddonConfig{})
	testscheme.AddKnownTypes(manifestworkv1.SchemeGroupVersion, &manifestworkv1.ManifestWork{})
	testscheme.AddKnownTypes(coordinationv1.SchemeGroupVersion, &coordinationv1.Lease{})

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
		mca            *addonv1alpha1.ManagedClusterAddOn
		errProgressing error
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
				mca:            &addonv1alpha1.ManagedClusterAddOn{},
				errProgressing: nil,
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
				errProgressing: nil,
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: "",
			wantType:   false,
		},
		{
			name: "has processing error",
			args: args{
				mca: &addonv1alpha1.ManagedClusterAddOn{
					Status: addonv1alpha1.ManagedClusterAddOnStatus{
						Conditions: []metav1.Condition{
							conditionAvailableFalse,
						},
					},
				},
				errProgressing: fmt.Errorf("failed install"),
			},
			wantStatus: metav1.ConditionTrue,
			wantReason: degradedReasonInstallError,
			wantType:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateDegradedStatus(tt.args.mca, tt.args.errProgressing)
			hasStatus := false
			for _, c := range tt.args.mca.Status.Conditions {
				if c.Type == "Degraded" && !tt.wantType {
					t.Errorf("updateProgressingStatus() error expect %v to not include Degraded", tt.args.mca.Status.Conditions)
					return
				}
				if c.Type == "Degraded" && c.Status == tt.wantStatus && c.Reason == tt.wantReason {
					hasStatus = true
				}
			}

			if tt.wantType && !hasStatus {
				t.Errorf("updateProgressingStatus() error expect %v to include Degraded=%v:%s",
					tt.args.mca.Status.Conditions, tt.wantStatus, tt.wantReason)
			}
		})
	}
}
