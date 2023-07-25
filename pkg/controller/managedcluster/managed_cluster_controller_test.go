// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedcluster

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mcv1 "open-cluster-management.io/api/cluster/v1"

	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/common"
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
		mc       *mcv1.ManagedCluster
		validate func(t *testing.T, kubeclient client.Client)
	}{
		{
			name: "create addon config for hosted addon enabled cluster",
			mc: newManagedCluster(testClusterName, map[string]string{
				common.AnnotationKlusterletDeployMode:         "Hosted",
				common.AnnotationKlusterletHostingClusterName: "local-cluster",
				common.AnnotationEnableHostedModeAddons:       "true",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if !kac.Spec.PolicyController.Enabled {
					t.Errorf("expected policy add-ons are enabled, but not enabled")
				}

				if kac.Spec.ApplicationManagerConfig.Enabled || kac.Spec.CertPolicyControllerConfig.Enabled ||
					kac.Spec.IAMPolicyControllerConfig.Enabled || kac.Spec.SearchCollectorConfig.Enabled {
					t.Errorf("expected other add-ons are disabled, but some of them is enabled")
				}
			},
		},
		{
			name: "create hypershift cluster klusterlet addon config",
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation: "test.test.HypershiftDeployment.cluster.open-cluster-management.io",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "create claim cluster klusterlet addon config",
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation: "test.test.ClusterClaim.hive.openshift.io/v1",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "create klusterlet addon config for normal managed cluster with the annotation",
			mc: newManagedCluster(testClusterName, map[string]string{
				common.AnnotationCreateWithDefaultKlusterletAddonConfig: "true",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "do not create klusterlet addon config for hypershift",
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation:                          "test.test.HypershiftDeployment.cluster.open-cluster-management.io",
				disableAddonAutomaticInstallationAnnotationKey: "true",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if !errors.IsNotFound(err) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "do not create klusterlet addon config for claim",
			mc: newManagedCluster(testClusterName, map[string]string{
				provisionerAnnotation:                          "test.test.ClusterClaim.hive.openshift.io/v1",
				disableAddonAutomaticInstallationAnnotationKey: "true",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if !errors.IsNotFound(err) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "do not create klusterlet addon config for normal managed cluster without the annotation",
			mc: newManagedCluster(testClusterName, map[string]string{
				common.AnnotationCreateWithDefaultKlusterletAddonConfig: "false",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if !errors.IsNotFound(err) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
		{
			name: "do not create klusterlet addon config for normal managed cluster with the annotation but disable addon automatic installation",
			mc: newManagedCluster(testClusterName, map[string]string{
				common.AnnotationCreateWithDefaultKlusterletAddonConfig: "true",
				disableAddonAutomaticInstallationAnnotationKey:          "true",
			}),
			validate: func(t *testing.T, kubeclient client.Client) {
				var kac kacv1.KlusterletAddonConfig
				err := kubeclient.Get(context.TODO(),
					types.NamespacedName{Namespace: testClusterName, Name: testClusterName}, &kac)
				if !errors.IsNotFound(err) {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeclient := fake.NewClientBuilder().WithScheme(testscheme).WithObjects(tt.mc).Build()
			reconciler := &ReconcileManagedCluster{
				client: kubeclient,
				scheme: testscheme,
			}

			_, err := reconciler.Reconcile(context.TODO(), request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			tt.validate(t, kubeclient)
		})
	}
}
