package globalproxy

import (
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	managedclusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
		&source.Kind{Type: &managedclusterv1.ManagedCluster{}},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
			return []reconcile.Request{
				{
					NamespacedName: types.NamespacedName{
						Name:      obj.GetName(),
						Namespace: obj.GetName(),
					},
				},
			}
		}),
	)
	if err != nil {
		return err
	}
	err = c.Watch(
		&source.Kind{Type: &agentv1.KlusterletAddonConfig{}},
		handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
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
	)
	if err != nil {
		return err
	}

	return nil
}
