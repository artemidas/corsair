# Corsair

Kubernetes security/pentesting desktop tool. Tauri v2 (Rust) backend, Vue 3 + TypeScript + Tailwind/DaisyUI frontend. No Python, no sidecar process — everything runs in-process in Rust.

## Current functionality

- Connect to a Kubernetes cluster via the local kubeconfig's active context (`kube` crate).
- Run a fixed set of security rules against the connected cluster.
- View findings in a table, color-coded by severity.

Results live in memory for the session only — nothing is persisted, and there's no support yet for multiple cluster sources, container image scanning, auth, or report export.

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

Requires a `~/.kube/config` (or `$KUBECONFIG`) with a usable active context — e.g. a local `kind` or `minikube` cluster.

### Project layout

```
src/                  Vue frontend
  components/
    Dashboard.vue     Connect / Run Scan / findings table
src-tauri/
  src/
    lib.rs            Tauri commands + app state
    cluster.rs         kube::Client from default kubeconfig
    rules.rs           Finding type, Rule trait, the 4 fixed rules
```

### Backend commands (IPC)

- `connect_cluster` — connects via kubeconfig, verifies by listing namespaces.
- `run_scan` — runs the fixed rule set against the connected cluster, returns `Finding[]`.

## Recommended IDE Setup

- [VS Code](https://code.visualstudio.com/) + [Vue - Official](https://marketplace.visualstudio.com/items?itemName=Vue.volar) + [Tauri](https://marketplace.visualstudio.com/items?itemName=tauri-apps.tauri-vscode) + [rust-analyzer](https://marketplace.visualstudio.com/items?itemName=rust-lang.rust-analyzer)
