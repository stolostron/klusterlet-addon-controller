// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedclusteraddon

import (
	"context"
	"reflect"
	"time"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/bindata"
	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	"github.com/open-cluster-management/library-go/pkg/applier"
	"github.com/open-cluster-management/library-go/pkg/templateprocessor"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var merger applier.Merger = func(current,
	new *unstructured.Unstructured,
) (
	future *unstructured.Unstructured,
	update bool,
) {
	if spec, ok := new.Object["spec"]; ok &&
		!reflect.DeepEqual(spec, current.Object["spec"]) {
		update = true
		current.Object["spec"] = spec
	}
	if rules, ok := new.Object["rules"]; ok &&
		!reflect.DeepEqual(rules, current.Object["rules"]) {
		update = true
		current.Object["rules"] = rules
	}
	if roleRef, ok := new.Object["roleRef"]; ok &&
		!reflect.DeepEqual(roleRef, current.Object["roleRef"]) {
		update = true
		current.Object["roleRef"] = roleRef
	}
	if subjects, ok := new.Object["subjects"]; ok &&
		!reflect.DeepEqual(subjects, current.Object["subjects"]) {
		update = true
		current.Object["subjects"] = subjects
	}
	return current, update
}

func createOrUpdateHubKubeConfigResources(
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	r *ReconcileManagedClusterAddOn,
	addon addons.KlusterletAddon) error {
	componentName := addon.GetAddonName()

	//Create the values for the yamls
	config := struct {
		ManagedClusterName      string
		ManagedClusterNamespace string
		ComponentName           string
		ManagedClusterAddOnName string
		ClusterRoleName         string
	}{
		ManagedClusterName:      klusterletaddonconfig.Name,
		ManagedClusterNamespace: klusterletaddonconfig.Name,
		ComponentName:           componentName,
		ManagedClusterAddOnName: addon.GetManagedClusterAddOnName(),
		ClusterRoleName:         addons.GetAddonClusterRolePrefix() + addon.GetManagedClusterAddOnName(),
	}

	newApplier, err := applier.NewApplier(
		bindata.NewBindataReader(),
		&templateprocessor.Options{},
		r.client,
		klusterletaddonconfig,
		r.scheme,
		merger,
		&applier.Options{
			Backoff: &wait.Backoff{
				Steps:    1,
				Duration: 10 * time.Millisecond,
				Factor:   1.0,
			},
		},
	)
	if err != nil {
		return err
	}

	err = newApplier.CreateOrUpdateInPath(
		"resources/hub/common",
		nil,
		false,
		config,
	)
	if err != nil {
		return err
	}
	// delete old role & rolebindings created in previous releases
	if err := deleteOutDatedRoleRoleBinding(addon, klusterletaddonconfig, r.client); err != nil {
		log.Info("Failed to delete outdated role/rolebinding. Skipping.", "error message:", err)
	}
	return nil
}

//deleteOutDatedRoleRoleBindings deletes old role/rolebinding with klusterletaddonconfig ownerRef (controller).
//it returns nil if no role/rolebinding exist, and it returns error when failed to delete the role/rolebinding
func deleteOutDatedRoleRoleBinding(
	addon addons.KlusterletAddon,
	klusterletaddonconfig *agentv1.KlusterletAddonConfig,
	client client.Client,
) error {
	if klusterletaddonconfig == nil {
		return nil
	}
	// name used in previous addon role & rolebinding
	name := klusterletaddonconfig.Name + "-" + addon.GetAddonName()
	// check if the role/rolebinding exist
	role := &rbacv1.Role{}
	rolebinding := &rbacv1.RoleBinding{}
	objs := make([]runtime.Object, 0)
	objs = append(objs, role)
	objs = append(objs, rolebinding)
	var retErr error
	retErr = nil
	for _, o := range objs {
		if err := client.Get(
			context.TODO(),
			types.NamespacedName{
				Name:      name,
				Namespace: klusterletaddonconfig.Namespace,
			}, o); err != nil && errors.IsNotFound(err) {
			continue
		} else if err != nil {
			retErr = err
			continue
		}

		// verify ownerRef
		if objMetaAccessor, ok := o.(metav1.ObjectMetaAccessor); !ok {
			log.V(2).Info("Failed to get ObjectMeta")
			continue
		} else {
			ownerRef := metav1.GetControllerOf(objMetaAccessor.GetObjectMeta())
			if ownerRef == nil {
				log.V(2).Info("No controller reference of the role, skipping")
				continue
			}
			ownerGV, err := schema.ParseGroupVersion(ownerRef.APIVersion)
			if err != nil {
				log.V(2).Info("Failed to get group from object ownerRef")
				continue
			}
			if ownerRef.Kind != klusterletaddonconfig.Kind ||
				ownerRef.Name != klusterletaddonconfig.Name ||
				ownerGV.Group != klusterletaddonconfig.GroupVersionKind().Group {
				log.V(2).Info("Object is not owned by klusterletaddonconfig. Skipping")
				continue
			}
		}

		if err := client.Delete(context.TODO(), o); err != nil && !errors.IsNotFound(err) {
			retErr = err
			continue
		}
	}

	return retErr
}

func deleteHubKubeConfigResources(
	addon addons.KlusterletAddon,
	managedClusterName string,
	client client.Client) error {
	componentName := addon.GetAddonName()

	//Create the values for the yamls
	config := struct {
		ManagedClusterName      string
		ManagedClusterNamespace string
		ComponentName           string
		ClusterRoleName         string
	}{
		ManagedClusterName:      managedClusterName,
		ManagedClusterNamespace: managedClusterName,
		ComponentName:           componentName,
		ClusterRoleName:         addons.GetAddonClusterRolePrefix() + addon.GetManagedClusterAddOnName(),
	}

	newApplier, err := applier.NewApplier(
		bindata.NewBindataReader(),
		&templateprocessor.Options{},
		client,
		nil,
		nil,
		nil,
		&applier.Options{
			Backoff: &wait.Backoff{
				Steps:    1,
				Duration: 10 * time.Millisecond,
				Factor:   1.0,
			},
		},
	)
	if err != nil {
		return err
	}

	err = newApplier.DeleteInPath(
		"resources/hub/common",
		nil,
		false,
		config,
	)
	if err != nil {
		return err
	}

	return nil
}
