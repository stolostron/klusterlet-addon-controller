// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package managedclusteraddon

import (
	"context"
	"fmt"
	"reflect"
	"time"

	addonv1alpha1 "github.com/open-cluster-management/api/addon/v1alpha1"
	manifestworkv1 "github.com/open-cluster-management/api/work/v1"
	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_managedclusteraddon")

const (

	// types of condition
	addonDegraded    = "Degraded"
	addonProgressing = "Progressing"

	// reasons of condition
	progressingReasonMissing   = "ManifestWorkCreating"
	progressingReasonCreated   = "ManifestWorkCreated"
	progressingReasonApplied   = "ManifestWorkApplied"
	progressingReasonDeleting  = "AddonTerminating"
	degradedReasonInstallError = "AddonInstallationError"

	// messages of condition
	progressingMSGMissing           = "Creating manifests for add-on installation." // message will show when we are waiting to create the manifests of addons
	progressingMSGCreated           = "Installing manifests."                       // message when we are still in installation
	progressingMSGApplied           = "All manifests are installed."                // message when the manifestwork is applied (manifest is installed)
	progressingMsgDeleting          = "Add-on is being deleted."                    // message when addon is in deletion
	degradedMSGInstallErrorTemplate = "Failed to complete add-on installation: %s." // message when we detect error in addon's manifests installation, %s is errorFailedApplyTemplate

	// possible error messages
	errorFailedApplyTemplate = "%d of %d manifests failed to apply"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ManagedClusterAddOn Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedClusterAddOn{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("managedclusteraddon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ManagedClusterAddOn
	err = c.Watch(&source.Kind{Type: &addonv1alpha1.ManagedClusterAddOn{}}, &handler.EnqueueRequestForObject{},
		addons.NewAddonNamePredicate())
	if err != nil {
		return err
	}
	// Watch for changes to manifestwork, and will check if manifestwork's name is matching one of the addons we own
	err = c.Watch(
		&source.Kind{Type: &manifestworkv1.ManifestWork{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				name := ""
				if addon, _ := addons.GetAddonFromManifestWorkName(obj.Meta.GetName()); addon != nil {
					name = addon.GetManagedClusterAddOnName()
				}
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      name,
							Namespace: obj.Meta.GetNamespace(),
						},
					},
				}
			},
		)},
		newManifestWorkPredicate(),
	)
	if err != nil {
		return err
	}
	// Watch for changes to lease
	err = c.Watch(
		&source.Kind{Type: &coordinationv1.Lease{}},
		&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(
			func(obj handler.MapObject) []reconcile.Request {
				return []reconcile.Request{
					{
						NamespacedName: types.NamespacedName{
							Name:      obj.Meta.GetName(),
							Namespace: obj.Meta.GetNamespace(),
						},
					},
				}
			},
		)},
		addons.NewAddonNamePredicate(),
	)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManagedClusterAddOn implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManagedClusterAddOn{}

// ReconcileManagedClusterAddOn reconciles a ManagedClusterAddOn object
type ReconcileManagedClusterAddOn struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManagedClusterAddOn object and makes changes based on the state read
// and status in the ManagedClusterAddOn
func (r *ReconcileManagedClusterAddOn) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ManagedClusterAddOn")
	// Fetch the related addon
	addon, err := addons.GetAddonFromManagedClusterAddonName(request.Name)
	if err != nil {
		log.V(2).Info("The given ManagedClusterAddon has no supported name. Skipping", request)
		return reconcile.Result{}, nil
	}
	// Fetch the ManagedClusterAddOn instance
	managedClusterAddOn := &addonv1alpha1.ManagedClusterAddOn{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, managedClusterAddOn); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// make sure we are handling the ManagedClusterAddOns with status.addonResource=klustterletaddonconfigs
	klusterletAddonConfigName, isReferenceCorrect := checkAddOnResourceGVR(
		managedClusterAddOn.Status.RelatedObjects,
		&schema.GroupVersionResource{
			Group:    agentv1.SchemeGroupVersion.Group,
			Version:  agentv1.SchemeGroupVersion.Version,
			Resource: "klusterletaddonconfigs",
		},
	)
	if !isReferenceCorrect {
		log.V(2).Info(fmt.Sprintf("ManagedClusterAddon %s has AddOn reference %v. Skipping.",
			managedClusterAddOn.Name, managedClusterAddOn.Status.RelatedObjects))
		return reconcile.Result{}, nil
	}

	// store oldstatus for compare in future to see if we need to update
	oldstatus := managedClusterAddOn.Status.DeepCopy()

	// Fetch the klusterletaddonconfig instance for enable/disable settings
	klusterletaddonconfig := &agentv1.KlusterletAddonConfig{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      klusterletAddonConfigName,
		Namespace: request.Namespace,
	}, klusterletaddonconfig); err != nil {
		// if klusterletaddonconfig is not found we should delete ManagedClusterAddOn and hub kubeconfig resources
		if errors.IsNotFound(err) {
			log.Error(err, "klusterletaddonconfig not found, deleting ManagedClusterAddOn "+managedClusterAddOn.Name)
			delErr := deleteAll(r.client, addon, managedClusterAddOn)
			return reconcile.Result{}, delErr
		}
		log.Error(err, "failed to get klusterletaddonconfig")
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fetch the manifestwork instance
	manifestWork := &manifestworkv1.ManifestWork{}
	manifestWorkIsNotFound := false
	if err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      addons.ConstructManifestWorkName(klusterletaddonconfig, addon),
		Namespace: request.Namespace,
	}, manifestWork); err != nil && !errors.IsNotFound(err) {
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	} else if errors.IsNotFound(err) {
		manifestWorkIsNotFound = true
	}

	// manifesttwork not found means deletion complete, we can delete the resource now
	if manifestWorkIsNotFound &&
		(!addon.IsEnabled(klusterletaddonconfig) || klusterletaddonconfig.DeletionTimestamp != nil) {
		// delete all
		if err := deleteAll(r.client, addon, managedClusterAddOn); err != nil {
			log.Error(err, "failed to delete ManagedClusterAddOn %s and its hub kubeconfig resourcers"+managedClusterAddOn.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// check & set addon install namespace
	if managedClusterAddOn.Spec.InstallNamespace != addonoperator.KlusterletAddonNamespace {
		managedClusterAddOn.Spec.InstallNamespace = addonoperator.KlusterletAddonNamespace
		if err := r.client.Update(context.TODO(), managedClusterAddOn); err != nil {
			return reconcile.Result{}, err
		}
	}

	if addon.CheckHubKubeconfigRequired() {
		// creates necessary rolebinding for addon
		if err := createOrUpdateHubKubeConfigResources(klusterletaddonconfig, r, addon); err != nil {
			return reconcile.Result{}, fmt.Errorf("Failed to create rolebinding for componnet %s: %w", addon.GetAddonName(), err)
		}

		// set registration configuration
		managedClusterAddOn.Status.Registrations = []addonv1alpha1.RegistrationConfig{
			{SignerName: certificatesv1.KubeAPIServerClientSignerName},
		}
	}

	// update progressing status base on manifestwork
	_, errProgressing := updateProgressingStatus(
		managedClusterAddOn,
		addon.IsEnabled(klusterletaddonconfig) && klusterletaddonconfig.DeletionTimestamp == nil,
		manifestWorkIsNotFound,
		manifestWork,
	)

	// check & set degraded information
	_ = updateDegradedStatus(managedClusterAddOn, errProgressing)

	// write managedClusterAddOn status if needed
	if !reflect.DeepEqual(*oldstatus, managedClusterAddOn.Status) {
		err := r.client.Status().Update(context.TODO(), managedClusterAddOn)
		if err != nil {
			log.Error(err, "Failed to update status of ManagedClusterAddOn "+managedClusterAddOn.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// filterConditions removes conditions if they match the type
func filterConditions(conditions *[]metav1.Condition, excludeType string) {
	newConditions := []metav1.Condition{}
	for _, c := range *conditions {
		if c.Type != excludeType {
			newConditions = append(newConditions, c)
		}
	}
	if len(newConditions) == len(*conditions) {
		return
	}
	(*conditions) = (*conditions)[:0]
	*conditions = append(*conditions, newConditions...)
}

// updateDegradedStatus updates ManagedClusterAddOn.status's degraded type condition based on former errors
// will remove degraded condition if nothing is wrong
func updateDegradedStatus(mca *addonv1alpha1.ManagedClusterAddOn,
	errProgressing error) metav1.ConditionStatus {
	if errProgressing == nil {
		// filter out degraded
		filterConditions(&mca.Status.Conditions, addonDegraded)
		return metav1.ConditionFalse
	}
	var conditionReason string
	var conditionMsg string
	conditionType := addonDegraded
	conditionStatus := metav1.ConditionTrue
	// show progressing issues as higher priority
	if errProgressing != nil {
		conditionReason = degradedReasonInstallError
		conditionMsg = fmt.Sprintf(degradedMSGInstallErrorTemplate, errProgressing.Error())
	}
	condition := createCondition(conditionType, conditionStatus, conditionReason, conditionMsg)
	setStatusCondition(&mca.Status.Conditions, condition)
	return conditionStatus
}

// updateProgressingStatus updates ManagedClusterAddOn.status's processing type condition based on given manifestwork
// if manifestwork is not created/still waiting for complete, will show processing=true
// if manifestwork is finished apply (with or without errors), will show processing=false
// if there are any manifests applied failed, will return an error to indicate failed to apply the manifestwork
func updateProgressingStatus(
	mca *addonv1alpha1.ManagedClusterAddOn,
	isEnabled bool,
	manifestWorkIsNotFound bool,
	mw *manifestworkv1.ManifestWork,
) (status metav1.ConditionStatus, err error) {
	var conditionReason string
	var conditionMsg string
	conditionType := addonProgressing
	conditionStatus := metav1.ConditionTrue
	err = nil

	if !isEnabled {
		// when disabled, until completely deleted, should always show terminating
		conditionReason = progressingReasonDeleting
		conditionMsg = progressingMsgDeleting
	} else if manifestWorkIsNotFound {
		// when waiting for manifestwork to create
		conditionReason = progressingReasonMissing
		conditionMsg = progressingMSGMissing
	} else {
		numFailed, numSucceeded, numTotal := checkManifestWorkStatus(mw)
		// check if it's done, if applied > total, then it's done, otherwise it's not
		if numFailed+numSucceeded >= numTotal {
			conditionStatus = metav1.ConditionFalse
			conditionReason = progressingReasonApplied
			conditionMsg = progressingMSGApplied
		} else {
			conditionReason = progressingReasonCreated
			conditionMsg = progressingMSGCreated
		}
		if numFailed > 0 {
			err = fmt.Errorf(errorFailedApplyTemplate, numFailed, numTotal)
		}
	}
	// update condition
	condition := createCondition(conditionType, conditionStatus, conditionReason, conditionMsg)
	setStatusCondition(&mca.Status.Conditions, condition)
	return conditionStatus, err
}

// createCondition returns a condition based on given information
func createCondition(
	conditionType string,
	status metav1.ConditionStatus,
	reason string,
	msg string,
) *metav1.Condition {
	return &metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             reason,
		Message:            msg,
	}
}

// setStatusCondition appends new if there is no existed condition with same type
// will override a condition if it is with the same type, will do no changes if type & status & reason are the same
// this method assumes the given array of conditions don't have any two conditions with the same type
func setStatusCondition(conditions *[]metav1.Condition, condition *metav1.Condition) {
	for i, c := range *conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status || c.Reason != condition.Reason {
				(*conditions)[i] = *condition
			}
			return
		}
	}
	*conditions = append(*conditions, *condition)
}

// checkManifestWorkStatus checks the manifsetwork.status to report manifestwork apply status
func checkManifestWorkStatus(m *manifestworkv1.ManifestWork) (numFailed, numSucceeded, numTotal int) {
	// input should never be nil
	if m == nil {
		return
	}
	numFailed = 0
	numSucceeded = 0
	numTotal = len(m.Spec.Workload.Manifests)

	manifestIsAppliedArray := make([]bool, len(m.Spec.Workload.Manifests))
	manifestFailedArray := make([]bool, len(m.Spec.Workload.Manifests))

	for _, mc := range m.Status.ResourceStatus.Manifests {
		isApplied := false
		hasError := false
		ordinal := int(mc.ResourceMeta.Ordinal)
		// will search for applied=true in conditions, and record error messages if not applied
		for _, c := range mc.Conditions {
			if manifestworkv1.ManifestConditionType(c.Type) == manifestworkv1.ManifestApplied &&
				c.Status == metav1.ConditionTrue {
				isApplied = true
			}
			// applied & false means error based on implementation of manifestwork:
			// https://github.com/open-cluster-management/work/blob/1fa05673bdbca451c8c99624ad9a91c33950018f/pkg/spoke/controllers/manifestcontroller/manifestwork_controller.go#L363
			if manifestworkv1.ManifestConditionType(c.Type) == manifestworkv1.ManifestApplied &&
				c.Status == metav1.ConditionFalse {
				hasError = true
			}
		}
		if ordinal >= 0 && ordinal < len(manifestIsAppliedArray) {
			if hasError {
				manifestFailedArray[ordinal] = true
			} else if isApplied {
				manifestIsAppliedArray[ordinal] = true
			}
		}
	}
	// count how many manifests are applied
	for _, b := range manifestIsAppliedArray {
		if b {
			numSucceeded++
		}
	}
	for _, b := range manifestFailedArray {
		if b {
			numFailed++
		}
	}

	return numFailed, numSucceeded, numTotal
}

// checkAddOnResourceGVR checks if ManagedClusterAddOn's addon referernce is the same as the given gvr.
// returns the resource name and true if is the same.
func checkAddOnResourceGVR(refs []addonv1alpha1.ObjectReference, gvr *schema.GroupVersionResource) (string, bool) {
	if len(refs) == 0 || gvr == nil {
		return "", false
	}
	for _, ref := range refs {
		if ref.Group == gvr.Group && ref.Resource == gvr.Resource {
			return ref.Name, true
		}
	}
	return "", false
}

// deleteAll deletes given ManagedClusterAddOn and hub kubeconfig resources, returns nil if
// deleted or not found
func deleteAll(
	c client.Client,
	addon addons.KlusterletAddon,
	mca *addonv1alpha1.ManagedClusterAddOn,
) error {
	if mca != nil {
		if err := c.Delete(context.TODO(), mca); err != nil && !errors.IsNotFound(err) {
			return err
		}
	}

	// delete the hub kubeconfig resources, like clusterrolebinding
	return deleteHubKubeConfigResources(addon, mca.Namespace, c)
}
