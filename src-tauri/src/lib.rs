mod builtin_rules;
mod cluster;
mod custom_rule;
mod projects;
mod rule_pack;
mod rules;

use k8s_openapi::api::core::v1::Namespace;
use kube::api::ListParams;
use kube::{Api, Client};
use cluster::{ClusterStatus, KubeContexts};
use custom_rule::{CustomRule, CustomRuleInput};
use projects::{Project, ProjectInput};
use rule_pack::{ExportScope, ImportMode, ImportSummary};
use rules::Finding;
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

#[tauri::command]
async fn run_scan(
    state: State<'_, AppState>,
    db: State<'_, DbInstances>,
) -> Result<Vec<Finding>, String> {
    let client = state
        .client
        .lock()
        .unwrap()
        .clone()
        .ok_or_else(|| "not connected to a cluster".to_string())?;

    let data = rules::fetch_cluster_data(&client)
        .await
        .map_err(|e| e.to_string())?;

    let mut findings = rules::run_rules(&data);

    let custom_rules = custom_rule::all_rules(&db).await?;
    findings.extend(rules::evaluate_custom_rules(&client, &custom_rules).await);

    Ok(findings)
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
    projects::delete_project(&db, &id).await
}

#[tauri::command]
async fn list_custom_rules(db: State<'_, DbInstances>) -> Result<Vec<CustomRule>, String> {
    let mut out = custom_rule::builtin();
    out.extend(custom_rule::list_rules(&db).await?);
    Ok(out)
}

#[tauri::command]
async fn create_custom_rule(
    db: State<'_, DbInstances>,
    input: CustomRuleInput,
) -> Result<CustomRule, String> {
    custom_rule::create_rule(&db, input).await
}

#[tauri::command]
async fn update_custom_rule(
    db: State<'_, DbInstances>,
    id: String,
    input: CustomRuleInput,
) -> Result<CustomRule, String> {
    custom_rule::update_rule(&db, id, input).await
}

#[tauri::command]
async fn delete_custom_rule(db: State<'_, DbInstances>, id: String) -> Result<(), String> {
    custom_rule::delete_rule(&db, &id).await
}

#[tauri::command]
async fn export_rules(
    db: State<'_, DbInstances>,
    path: String,
    scope: ExportScope,
) -> Result<usize, String> {
    rule_pack::export_to_path(&db, path, scope).await
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
            run_scan,
            list_projects,
            get_project,
            create_project,
            update_project,
            delete_project,
            list_custom_rules,
            create_custom_rule,
            update_custom_rule,
            delete_custom_rule,
            import_rules,
            export_rules,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
