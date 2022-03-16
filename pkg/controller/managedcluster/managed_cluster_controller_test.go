// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedcluster

import (
	"context"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcv1 "open-cluster-management.io/api/cluster/v1"

	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

func newManagedCluster(clusterName string, anno map[string]string) *mcv1.ManagedCluster {
	return &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        clusterName,
			Annotations: anno,
		},
		Spec: mcv1.ManagedClusterSpec{HubAcceptsClient: true},
	}
}

func TestReconcileManagedCluster(t *testing.T) {
	testClusterName := "cluster1"
	testscheme := scheme.Scheme
	_ = mcv1.AddToScheme(testscheme)
	_ = apis.AddToScheme(testscheme)

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: testClusterName,
		},
	}

	tests := []struct {
		name     string
		initObjs []runtime.Object
		mc       *mcv1.ManagedCluster
		want     reconcile.Result
	}{
		{
			name:     "create hypershift cluster klusterlet addon config",
			initObjs: nil,
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation: "test.test.HypershiftDeployment.cluster.open-cluster-management.io",
			}),
			want: reconcile.Result{Requeue: false},
		},
		{
			name:     "create claim cluster klusterlet addon config",
			initObjs: nil,
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation: "test.test.ClusterClaim.hive.openshift.io/v1",
			}),
			want: reconcile.Result{Requeue: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			reconciler := &ReconcileManagedCluster{
				client: fake.NewFakeClientWithScheme(testscheme, append(tt.initObjs, tt.mc)...),
				scheme: testscheme,
			}

			actual, err := reconciler.Reconcile(request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("expected %v but got %v", tt.want, actual)
			}

			var kac kacv1.KlusterletAddonConfig
			err = reconciler.client.Get(context.TODO(), types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
