package globalproxy

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
)

func Add(mgr manager.Manager, kubeClient kubernetes.Interface) error {
	return add(mgr, newGlobalProxyReconciler(mgr, kubeClient))
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("globalProxy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(
		source.Kind(mgr.GetCache(), &managedclusterv1.ManagedCluster{},
			handler.TypedEnqueueRequestsFromMapFunc[*managedclusterv1.ManagedCluster](
				func(ctx context.Context, cluster *managedclusterv1.ManagedCluster) []reconcile.Request {
					return []reconcile.Request{
						{
							NamespacedName: types.NamespacedName{
								Name:      cluster.GetName(),
								Namespace: cluster.GetName(),
							},
						},
					}
				}),
		))

	if err != nil {
		return err
	}
	err = c.Watch(
		source.Kind(mgr.GetCache(), &agentv1.KlusterletAddonConfig{},
			handler.TypedEnqueueRequestsFromMapFunc[*agentv1.KlusterletAddonConfig](
				func(ctx context.Context, obj *agentv1.KlusterletAddonConfig) []reconcile.Request {
					namespace := obj.GetNamespace()
					name := obj.GetName()
					if name == namespace {
						return []reconcile.Request{
							{
								NamespacedName: types.NamespacedName{
									Name:      namespace,
									Namespace: namespace,
								},
							},
						}
					}
					return nil
				}),
		),
	)
	if err != nil {
		return err
	}

	return nil
}
