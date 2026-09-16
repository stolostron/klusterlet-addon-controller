# klusterlet-addon-controller Architecture

## Purpose

The controller runs on the ACM hub cluster and translates `KlusterletAddonConfig`
and `ManagedCluster` state into `ManagedClusterAddOn` resources. ACM's addon
management machinery then delivers the selected addon workloads to each managed
cluster, normally through the addon operator and ManifestWork resources.

## Runtime components

`cmd/manager/main.go` creates a controller-runtime manager, loads image data from
the hub, builds Kubernetes and dynamic clients, registers the ACM and local API
schemes, and starts the controllers with leader election. Metrics are exposed on
`0.0.0.0:8383`; the command-line `--metrics-addr` value is retained for the
legacy entrypoint but does not replace the manager metrics bind setting.

`pkg/controller/controller.go` registers three reconcilers:

- `pkg/controller/addon/`: reconciles addon enablement and global addon values.
- `pkg/controller/managedcluster/`: creates or recreates the default addon config and handles cluster-specific setup predicates.
- `pkg/controller/globalproxy/`: propagates relevant managed-cluster and addon-config changes for proxy configuration.

## Main data flow

1. A `ManagedCluster` and namespaced `KlusterletAddonConfig` are observed by the relevant watches.
2. The addon reconciler skips deleted or paused clusters, evaluates each known addon, and honors Placement-based `ClusterManagementAddOn` strategies.
3. For enabled addons it computes node selectors, image overrides, proxy settings, and hosted-mode placement. These values are serialized into the `addon.open-cluster-management.io/values` annotation on `ManagedClusterAddOn`.
4. Disabled addons are deleted. Enabled addons are created or updated in the managed-cluster namespace, with hosted addons using the hosting cluster's install namespace when configured.
5. Downstream ACM addon controllers consume the resulting `ManagedClusterAddOn` resources and deliver addon components to the managed cluster.

## API and deployment boundaries

The local API types live under `pkg/apis/agent/v1/`; the CRD and RBAC/deployment
manifests are under `deploy/`. Changes to API types require `make generate` and
usually `make manifests`, followed by checking the generated CRD and deepcopy
outputs. E2E fixtures under `test/e2e/resources/` are the concrete integration
contract for cluster behavior.

The controller depends on ACM APIs for `ManagedCluster`, `ManagedClusterAddOn`,
and ManifestWork-related integration, plus OpenShift infrastructure APIs for
cluster proxy and image registry behavior. It is hub-side code and expects access
to the hub API server; it does not directly manage a managed-cluster API server.

## Important design decisions

- Event predicates intentionally filter watches to avoid reconciling every cluster update.
- `KlusterletAddons` distinguishes addons managed by this controller from addons that are only observed or deprecated.
- The values annotation is merged rather than blindly replaced in helper paths so other addon values can survive controller updates.
- The `klusterletaddonconfig-pause` annotation is a deliberate development and recovery escape hatch; paused resources must not be reconciled.
- Generated API files and manifests are checked-in artifacts and should be regenerated rather than hand-edited.
