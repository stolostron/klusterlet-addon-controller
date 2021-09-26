// Copyright Contributors to the Open Cluster Management project

package klusterletaddon

import (
	"fmt"
	"reflect"
	"time"

	"context"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type GlobalProxyReconciler struct {
	client client.Client
	scheme *runtime.Scheme
}

func newGlobalProxyReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &GlobalProxyReconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}
}

func globalProxyReconcilerAdd(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("globalProxy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(
		&source.Kind{Type: &corev1.Secret{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				namespace := obj.Meta.GetNamespace()
				installConfigSecretName := fmt.Sprintf("%s-install-config", namespace)
				name := obj.Meta.GetName()
				// only handle the install-config secret in cluster namespace
				if name == installConfigSecretName {
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
			},
		)},
	)
	if err != nil {
		return err
	}
	err = c.Watch(
		&source.Kind{Type: &agentv1.KlusterletAddonConfig{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				namespace := obj.Meta.GetNamespace()
				name := obj.Meta.GetName()
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
			},
		)},
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *GlobalProxyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Name", req.Name)
	reqLogger.Info("Reconciling GlobalProxy")

	klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Name: req.Name, Namespace: req.Namespace},
		klusterletAddonConfig); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
		}
		return reconcile.Result{}, err
	}

	newStatus := klusterletAddonConfig.Status.DeepCopy()

	installConfigSecret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{Name: fmt.Sprintf("%s-install-config", req.Name), Namespace: req.Namespace},
		installConfigSecret); err != nil {
		if errors.IsNotFound(err) {
			newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
			meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
				Type:    agentv1.OCPGlobalProxyDetected,
				Status:  metav1.ConditionFalse,
				Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
				Message: "The cluster is not provisioned by ACM.",
			})
			return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
		}
		return reconcile.Result{}, err
	}

	globalProxy, err := getGlobalProxyConfig(installConfigSecret)
	if err != nil {
		newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionFalse,
			Reason:  agentv1.ReasonOCPGlobalProxyDetectedFail,
			Message: err.Error(),
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	if globalProxy.NoProxy == "" && globalProxy.HTTPProxy == "" && globalProxy.HTTPSProxy == "" {
		newStatus.OCPGlobalProxy = agentv1.ProxyConfig{}
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionFalse,
			Reason:  agentv1.ReasonOCPGlobalProxyNotDetected,
			Message: "There is no cluster-wide proxy config in install config.",
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	if !reflect.DeepEqual(newStatus.OCPGlobalProxy, globalProxy) {
		newStatus.OCPGlobalProxy = globalProxy
		meta.SetStatusCondition(&newStatus.Conditions, metav1.Condition{
			Type:    agentv1.OCPGlobalProxyDetected,
			Status:  metav1.ConditionTrue,
			Reason:  agentv1.ReasonOCPGlobalProxyDetected,
			Message: "Detected the cluster-wide proxy config in install config.",
		})
		return reconcile.Result{}, r.updateStatus(req.Namespace, newStatus)
	}

	// Set the ProxyPolicy of ApplicationManager in KlusterletAddonConfig to OCPGlobalProxy when
	// ApplicationManager is enabled and ProxyPolicy is not set by user.
	if klusterletAddonConfig.Spec.ApplicationManagerConfig.Enabled &&
		klusterletAddonConfig.Spec.ApplicationManagerConfig.ProxyPolicy == "" {
		klusterletAddonConfig.Spec.ApplicationManagerConfig.ProxyPolicy = agentv1.ProxyPolicyOCPGlobalProxy
		return ctrl.Result{}, r.client.Update(context.TODO(), klusterletAddonConfig)
	}
	return ctrl.Result{}, nil
}

func (r *GlobalProxyReconciler) updateStatus(clusterName string, status *agentv1.KlusterletAddonConfigStatus) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		klusterletAddonConfig := &agentv1.KlusterletAddonConfig{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: clusterName, Namespace: clusterName},
			klusterletAddonConfig)
		if err != nil {
			return err
		}

		newStatus := klusterletAddonConfig.Status.DeepCopy()
		newStatus.OCPGlobalProxy = status.OCPGlobalProxy
		for _, condition := range status.Conditions {
			meta.SetStatusCondition(&newStatus.Conditions, condition)
		}
		if !equality.Semantic.DeepEqual(klusterletAddonConfig.Status, newStatus) {
			klusterletAddonConfig.Status = *newStatus
			return r.client.Status().Update(context.TODO(), klusterletAddonConfig, &client.UpdateOptions{})
		}
		return nil
	})
	return err
}

// getGlobalProxyConfig gets proxyConfig from install-config secret
func getGlobalProxyConfig(installConfigSecret *corev1.Secret) (agentv1.ProxyConfig, error) {
	proxyConfig := agentv1.ProxyConfig{}
	if len(installConfigSecret.Data) == 0 {
		return proxyConfig, fmt.Errorf("miss Data in install config secret %v", installConfigSecret.Name)
	}

	installConfigData, ok := installConfigSecret.Data["install-config.yaml"]
	if !ok {
		return proxyConfig, fmt.Errorf("miss install-config.yaml in install config secret %v", installConfigSecret.Name)
	}

	// installConfig, err := base64.StdEncoding.DecodeString(string(installConfigData))
	// if err != nil {
	// 	klog.Errorf("invalid install-config.yaml: %v in install config secret %v. err %v\n", string(installConfigData),installConfigSecret.Name, err)
	// 	return proxyConfig, fmt.Errorf("invalid install-config.yaml in install config secret %v. err: %v", installConfigSecret.Name, err)
	// }

	return getGlobalProxyInInstallConfig(installConfigData)
}

// getGlobalProxyInInstallConfig gets proxyConfig from install-config.yaml
// refer: https://github.com/openshift/installer/blob/master/docs/user/customization.md#proxy
func getGlobalProxyInInstallConfig(installConfig []byte) (agentv1.ProxyConfig, error) {
	proxyConfig := agentv1.ProxyConfig{}
	proxyConfigRaw := map[string]interface{}{}

	err := yaml.Unmarshal(installConfig, &proxyConfigRaw)
	if err != nil {
		return proxyConfig, err
	}

	// proxy defined in https://github.com/openshift/installer/blob/master/pkg/types/installconfig.go
	proxyConfig.HTTPSProxy, _, err = unstructured.NestedString(proxyConfigRaw, "proxy", "httpsProxy")
	if err != nil {
		return proxyConfig, err
	}
	proxyConfig.HTTPProxy, _, err = unstructured.NestedString(proxyConfigRaw, "proxy", "httpProxy")
	if err != nil {
		return proxyConfig, err
	}
	proxyConfig.NoProxy, _, err = unstructured.NestedString(proxyConfigRaw, "proxy", "noProxy")
	if err != nil {
		return proxyConfig, err
	}

	return proxyConfig, nil
}
