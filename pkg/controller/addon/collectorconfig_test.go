// Copyright Contributors to the Open Cluster Management project

package addon

import (
	"context"
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	mcv1 "open-cluster-management.io/api/cluster/v1"

	"github.com/stolostron/klusterlet-addon-controller/pkg/apis"
	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

var searchGVK = schema.GroupVersionKind{
	Group:   "search.open-cluster-management.io",
	Version: "v1alpha1",
	Kind:    "Search",
}

// newSearchCR returns a minimal unstructured Search CR named searchInstanceName in the given
// namespace -- the only thing resolveSearchNamespace/getMergedCollectorConfigSpec actually need.
func newSearchCR(namespace string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(searchGVK)
	obj.SetName(searchInstanceName)
	obj.SetNamespace(namespace)
	return obj
}

// newMergedCollectorConfig returns an unstructured CollectorConfig CR named
// mergedCollectorConfigName in the given namespace. A nil spec omits the "spec" field entirely
// (simulating a CR with no spec set); pass map[string]interface{}{} for an explicit empty spec.
func newMergedCollectorConfig(namespace string, spec map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(collectorConfigGVK)
	obj.SetName(mergedCollectorConfigName)
	obj.SetNamespace(namespace)
	if spec != nil {
		_ = unstructured.SetNestedMap(obj.Object, spec, "spec")
	}
	return obj
}

func newFakeHubClient(objs ...client.Object) client.WithWatch {
	return fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(objs...).Build()
}

// TestResolveSearchNamespace_KAC covers resolveSearchNamespace directly: not-found (no error),
// found, ambiguous (multiple CRs named searchInstanceName -- must error rather than guess), and a
// CR present under a different name (must be ignored, not matched). Named with a _KAC suffix to
// avoid clashing with any equivalently-named test in a future shared-test-helpers refactor.
func TestResolveSearchNamespace_KAC(t *testing.T) {
	tests := []struct {
		name          string
		objects       []client.Object
		wantNamespace string
		wantFound     bool
		wantErr       bool
	}{
		{
			name:      "no Search CR at all",
			objects:   nil,
			wantFound: false,
		},
		{
			name:          "exactly one Search CR",
			objects:       []client.Object{newSearchCR("open-cluster-management")},
			wantNamespace: "open-cluster-management",
			wantFound:     true,
		},
		{
			name: "two Search CRs with the well-known name, in different namespaces -- ambiguous",
			objects: []client.Object{
				newSearchCR("ns1"),
				newSearchCR("ns2"),
			},
			wantErr: true,
		},
		{
			name: "a Search CR with a different name is ignored, not matched",
			objects: func() []client.Object {
				other := &unstructured.Unstructured{}
				other.SetGroupVersionKind(searchGVK)
				other.SetName("not-the-operator-cr")
				other.SetNamespace("ns3")
				return []client.Object{other}
			}(),
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hubClient := newFakeHubClient(tc.objects...)
			ns, found, err := resolveSearchNamespace(context.TODO(), hubClient)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found != tc.wantFound {
				t.Errorf("found = %v, want %v", found, tc.wantFound)
			}
			if found && ns != tc.wantNamespace {
				t.Errorf("namespace = %q, want %q", ns, tc.wantNamespace)
			}
		})
	}
}

// TestGetMergedCollectorConfigSpec covers the "not found" (both variants), populated, and empty
// spec cases end to end through getMergedCollectorConfigSpec.
func TestGetMergedCollectorConfigSpec(t *testing.T) {
	populated := map[string]interface{}{
		"collectionRules": []interface{}{
			map[string]interface{}{
				"action": "include",
				"resourceSelector": map[string]interface{}{
					"apiGroups": []interface{}{"example.io"},
					"kinds":     []interface{}{"Foo"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		objects  []client.Object
		wantSpec map[string]interface{}
	}{
		{
			name:     "no Search CR yet -- nil, no error",
			objects:  nil,
			wantSpec: nil,
		},
		{
			name:     "Search CR exists but merged-collector-config does not -- nil, no error",
			objects:  []client.Object{newSearchCR("open-cluster-management")},
			wantSpec: nil,
		},
		{
			name: "populated spec is returned as-is",
			objects: []client.Object{
				newSearchCR("open-cluster-management"),
				newMergedCollectorConfig("open-cluster-management", populated),
			},
			wantSpec: populated,
		},
		{
			name: "CR with no spec field set at all -- explicit empty map, not nil",
			objects: []client.Object{
				newSearchCR("open-cluster-management"),
				newMergedCollectorConfig("open-cluster-management", nil),
			},
			wantSpec: map[string]interface{}{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hubClient := newFakeHubClient(tc.objects...)
			spec, err := getMergedCollectorConfigSpec(context.TODO(), hubClient)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if (spec == nil) != (tc.wantSpec == nil) {
				t.Fatalf("spec = %#v, want %#v", spec, tc.wantSpec)
			}
			if tc.wantSpec != nil {
				gotRules, _, _ := unstructured.NestedSlice(map[string]interface{}{"x": spec}, "x", "collectionRules")
				wantRules, _, _ := unstructured.NestedSlice(map[string]interface{}{"x": tc.wantSpec}, "x", "collectionRules")
				if len(gotRules) != len(wantRules) {
					t.Errorf("collectionRules length = %d, want %d (spec=%#v)", len(gotRules), len(wantRules), spec)
				}
			}
		})
	}
}

// TestGetMergedCollectorConfigSpec_AmbiguousSearchCR_ReturnsError covers the "refuse to guess"
// case: getMergedCollectorConfigSpec must propagate resolveSearchNamespace's error.
func TestGetMergedCollectorConfigSpec_AmbiguousSearchCR_ReturnsError(t *testing.T) {
	hubClient := newFakeHubClient(newSearchCR("ns1"), newSearchCR("ns2"))
	_, err := getMergedCollectorConfigSpec(context.TODO(), hubClient)
	if err == nil {
		t.Fatal("expected an error when multiple Search CRs exist, got nil")
	}
}

// TestGetMergedCollectorConfigSpec_TransientGetError_ReturnsError is the key correctness
// regression test: a transient (non-NotFound) error from the hub Get must be propagated, not
// swallowed into (nil, nil) -- swallowing it would make buildAddonValuesAnnotation drop a
// previously-distributed collectorConfig annotation entry on the very next reconcile, turning a
// transient hub API hiccup into search-collector losing its collection rules fleet-wide.
func TestGetMergedCollectorConfigSpec_TransientGetError_ReturnsError(t *testing.T) {
	base := newFakeHubClient(newSearchCR("open-cluster-management"))
	failingClient := interceptor.NewClient(base, interceptor.Funcs{
		Get: func(
			ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption,
		) error {
			if u, ok := obj.(*unstructured.Unstructured); ok && u.GetKind() == "CollectorConfig" {
				return errors.New("simulated API server error")
			}
			return c.Get(ctx, key, obj, opts...)
		},
	})

	_, err := getMergedCollectorConfigSpec(context.TODO(), failingClient)
	if err == nil {
		t.Fatal("expected an error when the hub Get fails transiently, got nil -- must not be swallowed")
	}
}

// TestBuildAddonValuesAnnotation covers buildAddonValuesAnnotation directly: the search-collector
// addon must get a "collectorConfig" key when one is available on the hub; every other addon
// must never get one, even if the hub has a populated merged-collector-config (it's irrelevant to
// them); and an empty/absent config for search-collector must behave exactly like before this
// change for addons with no global values either (empty annotation, not written at all).
func TestBuildAddonValuesAnnotation(t *testing.T) {
	populatedSpec := map[string]interface{}{
		"collectionRules": []interface{}{
			map[string]interface{}{"action": "include"},
		},
	}
	hubWithConfig := newFakeHubClient(
		newSearchCR("open-cluster-management"),
		newMergedCollectorConfig("open-cluster-management", populatedSpec),
	)
	hubEmpty := newFakeHubClient()

	t.Run("search-collector addon with a populated hub config gets collectorConfig", func(t *testing.T) {
		r := &ReconcileKlusterletAddOn{client: hubWithConfig}
		out, err := r.buildAddonValuesAnnotation(context.TODO(), globalValues{}, agentv1.SearchAddonName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out == "" {
			t.Fatal("expected a non-empty annotation value")
		}
		if err := validateValues(out, `{"collectorConfig":{"spec":{"collectionRules":[{"action":"include"}]}}}`); err != nil {
			t.Errorf("unexpected annotation content: %v (got %s)", err, out)
		}
	})

	t.Run("non-search addon never gets collectorConfig, even with a populated hub config", func(t *testing.T) {
		r := &ReconcileKlusterletAddOn{client: hubWithConfig}
		out, err := r.buildAddonValuesAnnotation(context.TODO(), globalValues{}, agentv1.ApplicationAddonName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "" {
			t.Errorf("expected empty annotation for a non-search addon, got %q", out)
		}
	})

	t.Run("search-collector addon with no hub config and no global values -- empty annotation, unchanged behavior", func(t *testing.T) {
		r := &ReconcileKlusterletAddOn{client: hubEmpty}
		out, err := r.buildAddonValuesAnnotation(context.TODO(), globalValues{}, agentv1.SearchAddonName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out != "" {
			t.Errorf("expected empty annotation, got %q", out)
		}
	})

	t.Run("search-collector addon with global values but no hub config -- only global is written", func(t *testing.T) {
		r := &ReconcileKlusterletAddOn{client: hubEmpty}
		gv := globalValues{Global: global{NodeSelector: map[string]string{"foo": "bar"}}}
		out, err := r.buildAddonValuesAnnotation(context.TODO(), gv, agentv1.SearchAddonName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := validateValues(out, `{"global":{"nodeSelector":{"foo":"bar"}}}`); err != nil {
			t.Errorf("unexpected annotation content: %v (got %s)", err, out)
		}
	})

	t.Run("search-collector addon with both global values and hub config -- both present", func(t *testing.T) {
		r := &ReconcileKlusterletAddOn{client: hubWithConfig}
		gv := globalValues{Global: global{NodeSelector: map[string]string{"foo": "bar"}}}
		out, err := r.buildAddonValuesAnnotation(context.TODO(), gv, agentv1.SearchAddonName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := validateValues(out,
			`{"global":{"nodeSelector":{"foo":"bar"}},"collectorConfig":{"spec":{"collectionRules":[{"action":"include"}]}}}`,
		); err != nil {
			t.Errorf("unexpected annotation content: %v (got %s)", err, out)
		}
	})

	t.Run("transient hub error is propagated, not swallowed into an empty annotation", func(t *testing.T) {
		failingClient := interceptor.NewClient(hubWithConfig, interceptor.Funcs{
			Get: func(
				ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption,
			) error {
				if u, ok := obj.(*unstructured.Unstructured); ok && u.GetKind() == "CollectorConfig" {
					return errors.New("simulated API server error")
				}
				return c.Get(ctx, key, obj, opts...)
			},
		})
		r := &ReconcileKlusterletAddOn{client: failingClient}
		_, err := r.buildAddonValuesAnnotation(context.TODO(), globalValues{}, agentv1.SearchAddonName)
		if err == nil {
			t.Fatal("expected an error, got nil")
		}
	})
}

// TestReconcile_SearchCollectorAddonGetsCollectorConfigAnnotation is an end-to-end test through
// the real Reconcile() flow: given a ManagedCluster + KlusterletAddonConfig (both required for any
// addon to be created at all) plus a hub Search + merged-collector-config, the created
// search-collector ManagedClusterAddOn must carry the collectorConfig annotation, while every
// other addon created in the same reconcile must not.
func TestReconcile_SearchCollectorAddonGetsCollectorConfigAnnotation(t *testing.T) {
	testscheme := clientgoscheme.Scheme
	_ = mcv1.AddToScheme(testscheme)
	_ = addonv1alpha1.AddToScheme(testscheme)
	_ = apis.AddToScheme(testscheme)

	populatedSpec := map[string]interface{}{
		"collectionRules": []interface{}{
			map[string]interface{}{"action": "include"},
		},
	}

	objs := []client.Object{
		newManagedCluster("cluster1", nil, nil),
		newKlusterletAddonConfig("cluster1"),
		newSearchCR("open-cluster-management"),
		newMergedCollectorConfig("open-cluster-management", populatedSpec),
	}

	reconciler := &ReconcileKlusterletAddOn{
		client: fake.NewClientBuilder().WithScheme(testscheme).WithObjects(objs...).Build(),
	}

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "cluster1", Namespace: "cluster1"},
	}
	_, err := reconciler.Reconcile(context.TODO(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	searchAddon := &addonv1alpha1.ManagedClusterAddOn{}
	if err := reconciler.client.Get(context.TODO(),
		types.NamespacedName{Name: agentv1.SearchAddonName, Namespace: "cluster1"}, searchAddon); err != nil {
		t.Fatalf("failed to get search-collector ManagedClusterAddOn: %v", err)
	}
	gotValues := searchAddon.GetAnnotations()[annotationValues]
	if err := validateValues(gotValues, `{"collectorConfig":{"spec":{"collectionRules":[{"action":"include"}]}}}`); err != nil {
		t.Errorf("search-collector addon annotation mismatch: %v (got %q)", err, gotValues)
	}

	appAddon := &addonv1alpha1.ManagedClusterAddOn{}
	if err := reconciler.client.Get(context.TODO(),
		types.NamespacedName{Name: agentv1.ApplicationAddonName, Namespace: "cluster1"}, appAddon); err != nil {
		t.Fatalf("failed to get application-manager ManagedClusterAddOn: %v", err)
	}
	if _, ok := appAddon.GetAnnotations()[annotationValues]; ok {
		t.Errorf("application-manager addon must not carry a values annotation here, got %q",
			appAddon.GetAnnotations()[annotationValues])
	}
}
