mod cluster;
mod projects;
mod rules;

use k8s_openapi::api::core::v1::Namespace;
use kube::api::ListParams;
use kube::{Api, Client};
use projects::{Project, ProjectInput};
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
    /// Last context the client was built for. Lets the UI tell whether the
    /// active connection still matches the selected project.
    last_context: Mutex<Option<Option<String>>>,
}

#[tauri::command]
async fn connect_cluster(
    state: State<'_, AppState>,
    context: Option<String>,
) -> Result<(), String> {
    let client = cluster::connect(context.as_deref())
        .await
        .map_err(|e| e.to_string())?;

    // Verify the connection actually works by listing namespaces.
    let ns_api: Api<Namespace> = Api::all(client.clone());
    ns_api
        .list(&ListParams::default())
        .await
        .map_err(|e| e.to_string())?;

    *state.client.lock().unwrap() = Some(client);
    *state.last_context.lock().unwrap() = Some(context);
    Ok(())
}

#[tauri::command]
fn active_context(state: State<'_, AppState>) -> Option<Option<String>> {
    state.last_context.lock().unwrap().clone()
}

#[tauri::command]
async fn run_scan(state: State<'_, AppState>) -> Result<Vec<Finding>, String> {
    let client = state
        .client
        .lock()
        .unwrap()
        .clone()
        .ok_or_else(|| "not connected to a cluster".to_string())?;

    let data = rules::fetch_cluster_data(&client)
        .await
        .map_err(|e| e.to_string())?;

    Ok(rules::run_rules(&data))
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

fn migrations() -> Vec<Migration> {
    vec![Migration {
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
    }]
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(
            tauri_plugin_sql::Builder::default()
                .add_migrations(DB_KEY, migrations())
                .build(),
        )
        .manage(AppState::default())
        .invoke_handler(tauri::generate_handler![
            connect_cluster,
            active_context,
            run_scan,
            list_projects,
            get_project,
            create_project,
            update_project,
            delete_project,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
