package clustermanagementaddon

import (
	"reflect"
	"testing"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcileClusterManagementAddOn_Reconcile(t *testing.T) {
	testscheme := scheme.Scheme

	testscheme.AddKnownTypes(addonv1alpha1.SchemeGroupVersion, &addonv1alpha1.ClusterManagementAddOn{})

	testApplicationManagerCMA := &addonv1alpha1.ClusterManagementAddOn{
		TypeMeta: metav1.TypeMeta{
			APIVersion: addonv1alpha1.SchemeGroupVersion.String(),
			Kind:       "ClusterManagementAddOn",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "application-manager",
		},
		Spec: addonv1alpha1.ClusterManagementAddOnSpec{
			AddOnMeta: addonv1alpha1.AddOnMeta{
				DisplayName: "Application",
				Description: "test application manager",
			},
		},
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: "application-manager",
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
			name: "clustermanagementaddon do not exist, create it",
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
			name: "clustermanagementaddon modified, revert it",
			fields: fields{
				client: fake.NewFakeClientWithScheme(testscheme, testApplicationManagerCMA),
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
			r := &ReconcileClusterManagementAddOn{
				client: tt.fields.client,
				scheme: tt.fields.scheme,
			}

			got, err := r.Reconcile(tt.args.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReconcileClusterManagementAddOn.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileClusterManagementAddOn.Reconcile() = %v, want %v", got, tt.want)
			}
		})
	}

}
