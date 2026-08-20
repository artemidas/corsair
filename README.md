# Ladon

Kubernetes security/pentesting desktop tool. Tauri v2 (Rust) backend, Vue 3 + TypeScript + shadcn-vue / Tailwind v4 / DaisyUI frontend. Everything runs in-process — no Python, no sidecar, no external service.

## Current functionality

- **Projects** are first-class entities with CRUD UI. Two kinds:
  - `KubernetesClusterReview` — pinned to a kubeconfig context (or the active one when unset). Connect, run the fixed rule set, see findings.
  - `ContainerImageReview` — pinned to a container image reference. UI is a placeholder; the scanning logic is not implemented yet.
- Projects are persisted to an in-process SQLite database (`tauri-plugin-sql`).
- Cluster connection is in-memory for the session; the in-memory `kube::Client` is built from the system's default kubeconfig, optionally with a specific context.
- A light/dark theme toggle lives in the sidebar footer. Default is dark; choice persists to `localStorage`.

### Rules implemented

| ID | Severity | Check |
|---|---|---|
| RBAC001 | critical | `default` ServiceAccount in a namespace bound to `cluster-admin` via a ClusterRoleBinding |
| RBAC002 | high | Role or ClusterRole grants the wildcard verb `*` |
| POD001 | critical | Container has `securityContext.privileged: true` |
| POD004 | medium | Container doesn't set `securityContext.runAsNonRoot: true` |

## Development

```bash
bun install
bun run tauri dev
```

`bun run tauri dev` boots Vite + the Rust backend together. For typecheck / build verification only:

```bash
bun run build            # vue-tsc + vite
cd src-tauri && cargo build
```

A real or local (`kind` / `minikube`) cluster is needed to exercise the scan flow end-to-end.

## Project layout

```
src/                         Vue frontend
  App.vue                    sidebar + main layout
  components/
    AppSidebar.vue           shadcn-vue sidebar (project list + theme toggle)
    project/                 project feature
      ProjectDetail.vue      connect / run scan / findings (k8s), placeholder (image)
      ProjectEditor.vue      shadcn-vue Form + vee-validate + zod
    ui/                      shadcn-vue components
  composables/
    useProjects.ts           module-scoped reactive state for projects
    useTheme.ts              module-scoped theme state (default dark, persisted to localStorage)
src-tauri/
  src/
    lib.rs                   Tauri commands, AppState, plugin wiring
    cluster.rs               kube::Client from default kubeconfig (optional context)
    projects.rs              Project CRUD over SQLite
    rules.rs                 Finding type, Rule trait, the 4 fixed rules
```

## Backend commands (IPC)

- `list_projects` / `get_project` / `create_project` / `update_project` / `delete_project` — project CRUD over SQLite.
- `active_context` — returns the last context the in-memory `kube::Client` was built for (or `null` if never connected).
- `connect_cluster(context?)` — connects to the default kubeconfig (or the named context), verifies by listing namespaces, stores the client for the session.
- `run_scan` — runs the 4 fixed rules against the connected cluster, returns `Finding[]`.

## Roadmap

Not yet implemented — don't expect to find any of these:

- Container image scanning logic (the `ContainerImageReview` kind is reserved; the UI is a placeholder).
- Multiple cluster sources (local YAML manifests, in-cluster config, etc.).
- Auth / multi-user.
- Report export.
- Custom rule management / UI.
- Automatic rule updates.

## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Vue - Official](https://marketplace.visualstudio.com/items?itemName=Vue.volar) + [Tauri](https://marketplace.visualstudio.com/items?itemName=tauri-apps.tauri-vscode) + [rust-analyzer](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)
