// Copyright Contributors to the Open Cluster Management project

package addon

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// This file lets the search-collector ManagedClusterAddOn's values annotation carry the hub's
// merged-collector-config CollectorConfig Spec, so search-collector on every managed cluster
// picks up the hub's search collection rules without any ACM Policy. This file implements the
// "reuse the existing values-annotation transport" side of that, in the controller that actually
// owns the search-collector ManagedClusterAddOn object.
//
// Deliberately implemented against unstructured.Unstructured with a manually-set GVK, rather than
// importing search-v2-operator's api/v1alpha1 Go package: this avoids a hard cross-repo module
// dependency between two independently released operators. We only need the GVK and two
// well-known names — not the full typed API — and the search component may not even be installed
// on every hub.
var (
	searchListGVK = schema.GroupVersionKind{
		Group:   "search.open-cluster-management.io",
		Version: "v1alpha1",
		Kind:    "SearchList",
	}
	collectorConfigGVK = schema.GroupVersionKind{
		Group:   "search.open-cluster-management.io",
		Version: "v1alpha1",
		Kind:    "CollectorConfig",
	}
)

const (
	// searchInstanceName mirrors search-v2-operator's OperatorName constant — there is always
	// supposed to be exactly one Search CR cluster-wide, named this.
	searchInstanceName = "search-v2-operator"

	// mergedCollectorConfigName mirrors search-v2-operator's own constant of the same name — the
	// single, operator-computed CollectorConfig CR whose Spec is distributed to managed clusters.
	mergedCollectorConfigName = "merged-collector-config"
)

// resolveSearchNamespace returns the namespace of the live Search CR (there is always supposed
// to be exactly one, named searchInstanceName), via the hub client. found is false (with err nil)
// if no such CR exists — either the search component isn't installed on this hub (the Search CRD
// itself doesn't exist, surfaced by the List call failing with a NoKindMatchError/NotFound-style
// error, which is deliberately treated the same as "no CR found" here) or the CR simply hasn't
// been created yet. Callers must treat "not found" as "nothing to distribute yet", not as an
// error. Mirrors search-v2-operator/addon/addon.go's resolveSearchNamespace and its rationale for
// refusing to guess when more than one match is found.
func resolveSearchNamespace(ctx context.Context, hubClient client.Client) (namespace string, found bool, err error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(searchListGVK)
	if err := hubClient.List(ctx, list); err != nil {
		if isMissingKindOrCRDError(err) {
			return "", false, nil
		}
		return "", false, err
	}

	var matchNamespace string
	matchCount := 0
	for i := range list.Items {
		if list.Items[i].GetName() == searchInstanceName {
			matchNamespace = list.Items[i].GetNamespace()
			matchCount++
		}
	}
	switch {
	case matchCount > 1:
		return "", false, fmt.Errorf(
			"found %d Search CRs named %q — refusing to guess which one to use", matchCount, searchInstanceName)
	case matchCount == 0:
		return "", false, nil
	default:
		return matchNamespace, true, nil
	}
}

// getMergedCollectorConfigSpec fetches the .spec of the hub's merged-collector-config
// CollectorConfig CR as a plain map, for injecting into the search-collector addon's values
// annotation. Returns (nil, nil) when there is nothing to distribute yet: the search component
// isn't installed, the Search CR doesn't exist yet, or the operator hasn't computed
// merged-collector-config yet (all legitimate, e.g. during a fresh install race) — callers must
// treat that as "omit", not as an error.
//
// Any other error is returned, not swallowed: a caller that swallowed a transient error into "no
// config" here would end up writing an annotation that drops previously-distributed collection
// rules, the same correctness hazard documented in search-v2-operator's getCollectorConfigValue.
func getMergedCollectorConfigSpec(ctx context.Context, hubClient client.Client) (map[string]interface{}, error) {
	namespace, found, err := resolveSearchNamespace(ctx, hubClient)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	cc := &unstructured.Unstructured{}
	cc.SetGroupVersionKind(collectorConfigGVK)
	if err := hubClient.Get(ctx, types.NamespacedName{Name: mergedCollectorConfigName, Namespace: namespace}, cc); err != nil {
		if apierrors.IsNotFound(err) || isMissingKindOrCRDError(err) {
			return nil, nil
		}
		return nil, err
	}

	spec, found, err := unstructured.NestedMap(cc.Object, "spec")
	if err != nil {
		return nil, err
	}
	if !found {
		// No spec set at all — shouldn't normally happen, but an explicit empty spec is a
		// meaningful, valid state (consistent with search-v2-operator's own handling of an empty
		// CollectionRules list), not an error.
		return map[string]interface{}{}, nil
	}
	return spec, nil
}

// isMissingKindOrCRDError returns true for errors that indicate the requested CRD/kind isn't
// installed on this hub at all (as opposed to the specific object just not existing yet, which
// apierrors.IsNotFound already covers for a Get). A List or Get against a GVK with no matching
// CRD surfaces as a NoKindMatchError/NoResourceMatchError from the RESTMapper — both are treated
// the same way here (search component not installed → nothing to distribute, not an error).
func isMissingKindOrCRDError(err error) bool {
	if err == nil {
		return false
	}
	return apierrors.IsNotFound(err) || apimeta.IsNoMatchError(err)
}
