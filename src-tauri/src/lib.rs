mod cluster;
mod custom_rule;
mod projects;
mod rule_pack;
mod rules;
mod scans;

use k8s_openapi::api::core::v1::Namespace;
use kube::api::ListParams;
use kube::{Api, Client};
use cluster::{ClusterStatus, KubeContexts};
use custom_rule::{Rule, RuleInput};
use projects::{Project, ProjectInput};
use rule_pack::{ImportMode, ImportSummary};
use rules::Finding;
use scans::{Scan, ScanResult, StoredFinding};
use std::sync::Mutex;
use tauri::State;
use tauri_plugin_sql::{DbInstances, Migration, MigrationKind};

const DB_KEY: &str = "sqlite:corsair.db";

/// Holds the connected cluster client for the session. In-memory only,
/// single cluster, no persistence — matches this iteration's scope.
#[derive(Default)]
struct AppState {
    client: Mutex<Option<Client>>,
    /// Resolved kubeconfig context name the client was built for.
    last_context: Mutex<Option<String>>,
}

#[tauri::command]
async fn connect_cluster(
    state: State<'_, AppState>,
    context: Option<String>,
) -> Result<ClusterStatus, String> {
    let requested = context.filter(|s| !s.is_empty());
    let resolved = cluster::resolve_context_name(requested.as_deref());
    let client = cluster::connect(requested.as_deref())
        .await
        .map_err(|e| e.to_string())?;

    // Verify the connection actually works by listing namespaces.
    let ns_api: Api<Namespace> = Api::all(client.clone());
    ns_api
        .list(&ListParams::default())
        .await
        .map_err(|e| e.to_string())?;

    *state.client.lock().unwrap() = Some(client);
    *state.last_context.lock().unwrap() = resolved.clone();
    Ok(ClusterStatus::connected(resolved))
}

#[tauri::command]
fn disconnect_cluster(state: State<'_, AppState>) -> ClusterStatus {
    *state.client.lock().unwrap() = None;
    *state.last_context.lock().unwrap() = None;
    ClusterStatus::disconnected()
}

#[tauri::command]
async fn probe_cluster(state: State<'_, AppState>) -> Result<ClusterStatus, String> {
    let client = state.client.lock().unwrap().clone();
    let context = state.last_context.lock().unwrap().clone();
    let Some(client) = client else {
        return Ok(ClusterStatus::disconnected());
    };

    Ok(match cluster::probe(&client).await {
        Ok(()) => ClusterStatus::connected(context),
        Err(e) => ClusterStatus::unreachable(context, e.to_string()),
    })
}

#[tauri::command]
fn list_kube_contexts() -> Result<KubeContexts, String> {
    cluster::list_contexts()
}

fn connected_client(state: &AppState) -> Result<(Client, Option<String>), String> {
    let client = state
        .client
        .lock()
        .unwrap()
        .clone()
        .ok_or_else(|| "not connected to a cluster".to_string())?;
    let context = state.last_context.lock().unwrap().clone();
    Ok((client, context))
}

async fn evaluate_cluster(
    client: &Client,
    db: &DbInstances,
) -> Result<Vec<Finding>, String> {
    let data = rules::fetch_cluster_data(client)
        .await
        .map_err(|e| e.to_string())?;

    let mut findings = rules::run_rules(&data);

    let stored = custom_rule::list_rules(db).await?;
    findings.extend(rules::evaluate_rules(client, &stored).await);

    Ok(findings)
}

/// Live check against the connected cluster. Does not persist a scan.
#[tauri::command]
async fn preview_scan(
    state: State<'_, AppState>,
    db: State<'_, DbInstances>,
) -> Result<Vec<Finding>, String> {
    let (client, _) = connected_client(&state)?;
    evaluate_cluster(&client, &db).await
}

#[tauri::command]
async fn run_scan(
    state: State<'_, AppState>,
    db: State<'_, DbInstances>,
    project_id: String,
) -> Result<ScanResult, String> {
    projects::get_project(&db, &project_id)
        .await?
        .ok_or_else(|| format!("project '{project_id}' not found"))?;

    let (client, context) = connected_client(&state)?;
    let outcome = evaluate_cluster(&client, &db).await;
    scans::persist_scan(&db, project_id, context, outcome).await
}

#[tauri::command]
async fn list_scans(db: State<'_, DbInstances>, project_id: String) -> Result<Vec<Scan>, String> {
    scans::list_scans(&db, &project_id).await
}

#[tauri::command]
async fn get_scan(db: State<'_, DbInstances>, id: String) -> Result<Option<Scan>, String> {
    scans::get_scan(&db, &id).await
}

#[tauri::command]
async fn list_scan_findings(
    db: State<'_, DbInstances>,
    scan_id: String,
) -> Result<Vec<StoredFinding>, String> {
    scans::list_scan_findings(&db, &scan_id).await
}

#[tauri::command]
async fn list_projects(db: State<'_, DbInstances>) -> Result<Vec<Project>, String> {
    projects::list_projects(&db).await
}

#[tauri::command]
async fn get_project(db: State<'_, DbInstances>, id: String) -> Result<Option<Project>, String> {
    projects::get_project(&db, &id).await
}

#[tauri::command]
async fn create_project(
    db: State<'_, DbInstances>,
    input: ProjectInput,
) -> Result<Project, String> {
    projects::create_project(&db, input).await
}

#[tauri::command]
async fn update_project(
    db: State<'_, DbInstances>,
    id: String,
    input: ProjectInput,
) -> Result<Project, String> {
    projects::update_project(&db, id, input).await
}

#[tauri::command]
async fn delete_project(db: State<'_, DbInstances>, id: String) -> Result<(), String> {
    scans::delete_for_project(&db, &id).await?;
    projects::delete_project(&db, &id).await
}

#[tauri::command]
async fn list_rules(db: State<'_, DbInstances>) -> Result<Vec<Rule>, String> {
    custom_rule::list_rules(&db).await
}

#[tauri::command]
async fn create_rule(
    db: State<'_, DbInstances>,
    input: RuleInput,
) -> Result<Rule, String> {
    custom_rule::create_rule(&db, input).await
}

#[tauri::command]
async fn update_rule(
    db: State<'_, DbInstances>,
    id: String,
    input: RuleInput,
) -> Result<Rule, String> {
    custom_rule::update_rule(&db, id, input).await
}

#[tauri::command]
async fn delete_rule(db: State<'_, DbInstances>, id: String) -> Result<(), String> {
    custom_rule::delete_rule(&db, &id).await
}

#[tauri::command]
async fn export_rules(
    db: State<'_, DbInstances>,
    path: String,
) -> Result<usize, String> {
    rule_pack::export_to_path(&db, path).await
}

#[tauri::command]
async fn import_rules(
    db: State<'_, DbInstances>,
    path: String,
    mode: ImportMode,
) -> Result<ImportSummary, String> {
    rule_pack::import_from_path(&db, path, mode).await
}

fn migrations() -> Vec<Migration> {
    vec![
        Migration {
            version: 1,
            description: "create projects table",
            sql: "CREATE TABLE IF NOT EXISTS projects (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                kind TEXT NOT NULL,
                config TEXT NOT NULL,
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL
            )",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 2,
            description: "create custom_rules table",
            sql: "CREATE TABLE IF NOT EXISTS custom_rules (
                id TEXT PRIMARY KEY,
                title TEXT NOT NULL,
                description TEXT NOT NULL,
                severity TEXT NOT NULL,
                resource_type TEXT NOT NULL,
                field_path TEXT NOT NULL,
                expected_value TEXT NOT NULL,
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL
            )",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 3,
            description: "add operator to custom_rules",
            sql: "ALTER TABLE custom_rules ADD COLUMN operator TEXT NOT NULL DEFAULT 'equals'",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 4,
            description: "add import_id to custom_rules",
            sql: "ALTER TABLE custom_rules ADD COLUMN import_id TEXT",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 5,
            description: "unique index on custom_rules.import_id",
            sql: "CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_rules_import_id \
                  ON custom_rules(import_id) WHERE import_id IS NOT NULL",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 6,
            description: "create scans table",
            sql: "CREATE TABLE IF NOT EXISTS scans (
                id TEXT PRIMARY KEY,
                project_id TEXT NOT NULL,
                status TEXT NOT NULL,
                context TEXT,
                error TEXT,
                finding_count INTEGER NOT NULL DEFAULT 0,
                started_at TEXT NOT NULL,
                finished_at TEXT NOT NULL
            )",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 7,
            description: "index scans by project and time",
            sql: "CREATE INDEX IF NOT EXISTS idx_scans_project_started \
                  ON scans(project_id, started_at)",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 8,
            description: "create findings table",
            sql: "CREATE TABLE IF NOT EXISTS findings (
                id TEXT PRIMARY KEY,
                scan_id TEXT NOT NULL,
                rule_id TEXT NOT NULL,
                rule_title TEXT NOT NULL,
                severity TEXT NOT NULL,
                resource_kind TEXT NOT NULL,
                resource_name TEXT NOT NULL,
                namespace TEXT,
                message TEXT NOT NULL
            )",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 9,
            description: "index findings by scan and rule",
            sql: "CREATE INDEX IF NOT EXISTS idx_findings_scan_rule \
                  ON findings(scan_id, rule_id)",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 10,
            description: "seed former built-in rules as regular rows",
            sql: "INSERT OR IGNORE INTO custom_rules \
                  (id, title, description, severity, resource_type, field_path, operator, expected_value, import_id, created_at, updated_at) \
                  VALUES \
                  ('BUILTIN-001', 'Privileged container', 'Containers running with securityContext.privileged=true can access all host devices and capabilities.', 'critical', 'Pod', 'spec.containers[*].securityContext.privileged', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-002', 'Host network', 'Pods with spec.hostNetwork=true share the host''s network namespace and can listen on any interface.', 'high', 'Pod', 'spec.hostNetwork', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-003', 'Host PID namespace', 'Pods with spec.hostPID=true can see and signal all host processes.', 'high', 'Pod', 'spec.hostPID', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-004', 'Host IPC namespace', 'Pods with spec.hostIPC=true share the host''s IPC namespace.', 'medium', 'Pod', 'spec.hostIPC', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-005', 'Default ServiceAccount in use', 'Pods or workload bindings that rely on the ''default'' ServiceAccount inherit its permissive token by default.', 'medium', 'ServiceAccount', 'metadata.name', 'equals', 'default', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-006', 'Role grants wildcard verb', 'A Role granting the ''*'' verb allows every action on the listed resources.', 'high', 'Role', 'rules[*].verbs[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-007', 'ClusterRole grants wildcard verb', 'A ClusterRole granting the ''*'' verb allows every action cluster-wide.', 'high', 'ClusterRole', 'rules[*].verbs[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-008', 'Role grants wildcard API group', 'A Role granting the ''*'' apiGroup effectively grants every API.', 'high', 'Role', 'rules[*].apiGroups[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'), \
                  ('BUILTIN-009', 'ClusterRole grants wildcard API group', 'A ClusterRole granting the ''*'' apiGroup effectively grants every API cluster-wide.', 'high', 'ClusterRole', 'rules[*].apiGroups[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00')",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 11,
            description: "add public rule_id to custom_rules",
            sql: "ALTER TABLE custom_rules ADD COLUMN rule_id TEXT",
            kind: MigrationKind::Up,
        },
        Migration {
            version: 12,
            description: "unique index on custom_rules.rule_id",
            sql: "CREATE UNIQUE INDEX IF NOT EXISTS idx_custom_rules_rule_id \
                  ON custom_rules(rule_id) WHERE rule_id IS NOT NULL AND rule_id != ''",
            kind: MigrationKind::Up,
        },
    ]
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(
            tauri_plugin_sql::Builder::default()
                .add_migrations(DB_KEY, migrations())
                .build(),
        )
        .manage(AppState::default())
        .invoke_handler(tauri::generate_handler![
            connect_cluster,
            disconnect_cluster,
            probe_cluster,
            list_kube_contexts,
            preview_scan,
            run_scan,
            list_scans,
            get_scan,
            list_scan_findings,
            list_projects,
            get_project,
            create_project,
            update_project,
            delete_project,
            list_rules,
            create_rule,
            update_rule,
            delete_rule,
            import_rules,
            export_rules,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
