# klusterlet-addon-controller — Agent Instructions

This repository contains the ACM klusterlet addon controller, a Go Kubernetes
controller that manages addon configuration and `ManagedClusterAddOn` resources
for managed clusters.

## Repository layout

- `cmd/manager/`: controller-runtime manager entrypoint and client setup.
- `pkg/apis/agent/v1/`: `KlusterletAddonConfig` API types, addon constants, and generated deepcopy code.
- `pkg/controller/`: addon, managed-cluster, and global-proxy reconcilers.
- `deploy/`: CRDs, RBAC, deployment manifests, and development resources.
- `test/e2e/`: end-to-end tests that require a configured Kubernetes/OpenShift cluster.
- `build/`: dependency, test, lint, image, and cluster helper scripts.
- `vendor/`: vendored Go dependencies; normally excluded from review and edits.

## Development commands

Run commands from the repository root. The Makefile invokes `build/before-make.sh`,
so inspect that script if a target fails before reaching its stated command.

| Command | Purpose | Requirements |
|---|---|---|
| `make build` | Build the manager binary at `build/_output/manager`. | Go toolchain and dependencies |
| `make test` | Run unit tests through the repository test harness. | Go toolchain; downloads envtest assets |
| `make check` | Run the configured lint checks. | Network access for the lint harness |
| `make lint` | Run the configured Go linter. | Network access; shell tools |
| `make generate` | Regenerate API deepcopy code with `controller-gen`. | `controller-gen` or network access |
| `make manifests` | Regenerate the CRD manifest. | `controller-gen` or network access |
| `make build-e2e` | Compile the e2e test binary. | Go toolchain |
| `make test-e2e` | Build, prepare a cluster, deploy, and run e2e tests. | Kubernetes/OpenShift cluster and `KUBECONFIG` |
| `make deploy` | Apply the manifests in `deploy/` using `kubectl`. | Kubernetes/OpenShift cluster and `kubectl` |
| `make run` | Run locally with `operator-sdk`. | `operator-sdk` v0.18.1 and cluster access |

After changing `pkg/apis/agent/v1/*types.go`, run `make generate` and review the
generated `zz_generated.deepcopy.go` change. Keep generated files consistent with
their source types.

## Controller conventions

- The manager registers the addon, managed-cluster, and global-proxy controllers in `pkg/controller/controller.go`.
- Reconciliation reads hub-side `ManagedCluster` and namespaced `KlusterletAddonConfig` resources, then creates, updates, or deletes `ManagedClusterAddOn` resources.
- Addon values are propagated through the `addon.open-cluster-management.io/values` annotation; image overrides, node selectors, proxy settings, and hosted-mode behavior are significant inputs.
- Preserve controller-runtime event predicates and namespace/name mapping when changing watches; they intentionally limit reconciliation scope.
- Kubernetes API types and CRDs must remain compatible with the manifests under `deploy/` and the e2e fixtures under `test/e2e/`.

For system architecture, data flows, and module layout, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Personal configuration

Read personal config at the start of any task that needs an assignee, email, or project key.
Canonical path: ~/.config/user.local.md (tool-agnostic, global).
If the file does not exist, fall back to agent memory (`user-config`), then placeholders.
Run `make personalize` to generate or update the file (if this repo uses Fleet Engineering tooling).

## Tool integrations

- The `gh` CLI is not assumed to be installed in this environment. Use the configured GitHub MCP server for GitHub operations when available.
- Jira CLI is not assumed to be installed. Use the configured Jira MCP server for Jira operations when available.
- If a GitHub CLI is used in another environment, use `GH_TOKEN` and per-organization `GH_TOKEN_<ORG>` variables rather than embedding credentials.

## Fleet Engineering Skills

Fetch and apply the relevant skill when the task matches its domain.

| Skill | When to use |
|---|---|
| [backlog-grooming](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/backlog-grooming/SKILL.md) | Scan Jira work for grooming gaps and priorities |
| [breaking-changes](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/breaking-changes/SKILL.md) | Detect breaking API, configuration, behavior, or integration changes |
| [bug-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/bug-specialist/SKILL.md) | Triage bugs and plan reproductions or fixes |
| [ci-triage](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/ci-triage/SKILL.md) | Diagnose failing pull-request checks |
| [coderabbit-sync](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/coderabbit-sync/SKILL.md) | Maintain the Fleet reference CodeRabbit configuration |
| [cve-sustaining-handoff](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/cve-sustaining-handoff/SKILL.md) | Resolve or hand off CVE tracking work |
| [cve-triage](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/cve-triage/SKILL.md) | Gather vulnerability evidence and dispositions |
| [diagnosing-bugs](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/diagnosing-bugs/SKILL.md) | Reproduce and narrow unclear technical failures |
| [epic-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/epic-specialist/SKILL.md) | Plan multi-sprint Jira epics |
| [feature-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/feature-specialist/SKILL.md) | Define significant customer-facing capabilities |
| [finish-work](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/finish-work/SKILL.md) | Commit, push, open a PR, and update Jira when explicitly requested |
| [init-context-docs](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/init-context-docs/SKILL.md) | Assess and bootstrap repository context documentation |
| [jira-create](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/jira-create/SKILL.md) | Create Jira issues through the guided quality workflow |
| [jira-qe-readiness](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/jira-qe-readiness/SKILL.md) | Assess whether Jira work is ready for QE |
| [jira-report](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/jira-report/SKILL.md) | Produce Jira portfolio and quality reports |
| [jira-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/jira-specialist/SKILL.md) | Triage, search, link, and transition Jira work |
| [jira-type-audit](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/jira-type-audit/SKILL.md) | Audit Jira issue type correctness |
| [opencode-setup](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/opencode-setup/SKILL.md) | Configure OpenCode and its integrations |
| [org-repo-audit](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/org-repo-audit/SKILL.md) | Audit organization repositories for SDLC readiness |
| [pr-fix](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/pr-fix/SKILL.md) | Fix merge conflicts, CI failures, or review comments |
| [pr-hygiene](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/pr-hygiene/SKILL.md) | Manage stale pull-request lifecycle state |
| [pr-review](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/pr-review/SKILL.md) | Review GitHub pull requests with isolated context |
| [pr-review-detailed](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/pr-review-detailed/SKILL.md) | Perform layered checklist-based review analysis |
| [pr-review-fix](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/pr-review-fix/SKILL.md) | Iteratively review and fix local changes before commit |
| [release-notes](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/release-notes/SKILL.md) | Generate release notes from merged pull requests |
| [repo-content-audit](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/repo-content-audit/SKILL.md) | Find unlinked or orphaned repository content |
| [repo-setup](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/repo-setup/SKILL.md) | Onboard or refresh repository agentic SDLC files |
| [renovate-prs](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/renovate-prs/SKILL.md) | Manage dependency update pull requests |
| [rhacm-addon-wizard](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/rhacm-addon-wizard/SKILL.md) | Guide RHACM addon development |
| [risk-report](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/risk-report/SKILL.md) | Detect Jira risk signals and draft reports |
| [risk-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/risk-specialist/SKILL.md) | Plan mitigations and track delivery risks |
| [release-dod](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/release-dod/SKILL.md) | Create release Definition of Done checklists |
| [scored-code-review](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/scored-code-review/SKILL.md) | Use the deprecated scored review workflow when required by legacy work |
| [session-summary](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/session-summary/SKILL.md) | Summarize session work across Jira and GitHub |
| [spike-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/spike-specialist/SKILL.md) | Conduct time-boxed research or proof-of-concepts |
| [start-work](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/start-work/SKILL.md) | Create a Jira sub-task for current work |
| [story-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/story-specialist/SKILL.md) | Define user stories and acceptance criteria |
| [supportex-review](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/supportex-review/SKILL.md) | Review support exception requests |
| [task-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/task-specialist/SKILL.md) | Plan internal technical tasks |
| [test-coverage-gap](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/test-coverage-gap/SKILL.md) | Analyze risk-prioritized test coverage gaps |
| [ticket-specialist](https://raw.githubusercontent.com/OpenShift-Fleet/agentic-sdlc/main/skills/ticket-specialist/SKILL.md) | Intake and triage stakeholder requests |

The complete catalog is maintained in the Fleet source repository's `skills/README.md`.
