// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	"github.com/open-cluster-management/klusterlet-addon-controller/pkg/utils"
)

// const for addon operator
const (
	KlusterletAddonOperatorPostfix = "-klusterlet-addon-operator"
)

// createManifestWorkComponentOperator - creates manifest work for klusterlet addon operator
func createManifestWorkComponentOperator(
	klusterletaddoncfg *agentv1.KlusterletAddonConfig,
	r *ReconcileKlusterletAddon) error {

	var manifests []manifestworkv1.Manifest

	// create namespace
	klusterletaddonNamespace := addonoperator.NewNamespace()

	// Create Component Operator ClusteRole
	clusterRole := addonoperator.NewClusterRole(klusterletaddoncfg)

	// create cluster role binding
	clusterRoleBinding := addonoperator.NewClusterRoleBinding(klusterletaddoncfg)

	// create service account
	serviceAccount := addonoperator.NewServiceAccount(klusterletaddoncfg, addonoperator.KlusterletAddonNamespace)

	// create imagePullSecret
	imagePullSecret, err := addonoperator.NewImagePullSecret(klusterletaddoncfg, r.client)
	if err != nil {
		log.Error(err, "Fail to create imagePullSecret")
		return err
	}

	// create deployment for klusterlet addon operator
	deployment, err := addonoperator.NewDeployment(klusterletaddoncfg, addonoperator.KlusterletAddonNamespace)
	if err != nil {
		log.Error(err, "Fail to crreate desired klusterlet addon operator deployment")
		return err
	}
	// add namespace, clusterrole, clusterrolebinding, serviceaccount
	nsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: klusterletaddonNamespace}}
	crManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRole}}
	crbManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRoleBinding}}
	saManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: serviceAccount}}
	manifests = append(manifests, nsManifest, crManifest, crbManifest, saManifest)
	// add imagePullSecret
	if imagePullSecret != nil {
		ipsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: imagePullSecret}}
		manifests = append(manifests, ipsManifest)
	}
	// add deployment
	dplManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: deployment}}
	manifests = append(manifests, dplManifest)

	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      klusterletaddoncfg.Name + KlusterletAddonOperatorPostfix,
			Namespace: klusterletaddoncfg.Namespace,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}

	if err := utils.CreateOrUpdateManifestWork(manifestWork, r.client, klusterletaddoncfg, r.scheme); err != nil {
		log.Error(err, "Failed to create manifest work for component")
		return err
	}

	return nil
}
