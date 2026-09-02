# Ladon

Kubernetes security/pentesting desktop tool. Wails v3 (Go) backend, Vue 3 + TypeScript + shadcn-vue / Tailwind v4 / DaisyUI frontend. Everything runs in-process — no Python, no sidecar, no external service.

## Current functionality

- **Projects** are first-class entities with CRUD UI. Two kinds:
  - `KubernetesClusterReview` — pinned to a kubeconfig context (or the active one when unset). Connect, run the rule set, see findings.
  - `ContainerImageReview` — pinned to one or more container image references. UI can list local Docker/Podman images; scanning logic is not implemented yet.
- Projects, rules, scans, and findings persist to an in-process SQLite database (`modernc.org/sqlite`) under the user config dir.
- Cluster connection is in-memory for the session, built from the system's default kubeconfig via client-go.
- Rules are stored in SQLite (seeded builtins plus user-authored). Import and export as a versioned YAML pack.
- A light/dark theme toggle lives in the sidebar footer. Default is dark; choice persists to `localStorage`.

### Hardcoded evaluate checks

| ID | Severity | Check |
|---|---|---|
| RBAC001 | critical | `default` ServiceAccount in a namespace bound to `cluster-admin` via a ClusterRoleBinding |
| RBAC002 | high | Role or ClusterRole grants the wildcard verb `*` |
| POD001 | critical | Container has `securityContext.privileged: true` |
| POD004 | medium | Container doesn't set `securityContext.runAsNonRoot: true` |

Stored declarative rules are also evaluated against the same resources.

## Development

```bash
bun install
bun run wails:dev
```

`bun run wails:dev` boots Vite + the Go backend together. For typecheck / compile verification only:

```bash
bun run build            # vue-tsc + vite
go test ./...
go build -o bin/ladon .
```

A real or local (`kind` / `minikube`) cluster is needed to exercise the scan flow end-to-end. Docker or Podman is needed for the local image picker.

## Project layout

```
src/                         Vue frontend
  App.vue                    sidebar + main layout
  components/
    AppSidebar.vue           shadcn-vue sidebar (project list + theme toggle)
    project/                 project feature
    rule/                    rule editor / detail
    ui/                      shadcn-vue components
  composables/               module-scoped reactive state
  bindings/ladon/            generated Wails TypeScript bindings
main.go                      Wails v3 app + service wiring
appdb/                       SQLite open + migrations
project/                     project CRUD
rule/                        rule CRUD + YAML pack import/export
cluster/                     kubeconfig contexts + in-memory client
scan/                        preview / persist scans and findings
images/                      list local Docker/Podman images
```

## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Vue - Official](https://marketplace.visualstudio.com/items?itemName=Vue.volar) + [Go](https://marketplace.visualstudio.com/items?itemName=golang.go)
