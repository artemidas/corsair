mod cluster;
mod rules;

use k8s_openapi::api::core::v1::Namespace;
use kube::api::ListParams;
use kube::{Api, Client};
use rules::Finding;
use std::sync::Mutex;
use tauri::State;

/// Holds the connected cluster client for the session. In-memory only,
/// single cluster, no persistence — matches this iteration's scope.
#[derive(Default)]
struct AppState {
    client: Mutex<Option<Client>>,
}

#[tauri::command]
async fn connect_cluster(state: State<'_, AppState>) -> Result<(), String> {
    let client = cluster::connect().await.map_err(|e| e.to_string())?;

    // Verify the connection actually works by listing namespaces.
    let ns_api: Api<Namespace> = Api::all(client.clone());
    ns_api
        .list(&ListParams::default())
        .await
        .map_err(|e| e.to_string())?;

    *state.client.lock().unwrap() = Some(client);
    Ok(())
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

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .manage(AppState::default())
        .invoke_handler(tauri::generate_handler![connect_cluster, run_scan])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
