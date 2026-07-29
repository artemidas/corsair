//! Project CRUD, persisted to a SQLite database managed by tauri-plugin-sql.
//!
//! A "project" is a single security review. Two kinds are supported today:
//! - `KubernetesClusterReview`: pinned to a kubeconfig context (or the
//!   currently active one when none is set).
//! - `ContainerImageReview`: pinned to a container image reference.
//!
//! `kind` is the source of truth. The per-kind fields live on a flat
//! `ProjectConfig` struct; entries that don't apply to a given kind are
//! `None` (and validated as such on write).
//!
//! All persistence flows through the plugin's SQLite pool. Migrations are
//! applied on app startup via `tauri.conf.json` -> `plugins.sql.preload`,
//! so the pool is always available to commands below.

use chrono::Utc;
use serde::{Deserialize, Serialize};
use sqlx::SqlitePool;
use tauri_plugin_sql::{DbInstances, DbPool};
use uuid::Uuid;

const DB_KEY: &str = "sqlite:corsair.db";

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub enum ProjectKind {
    KubernetesClusterReview,
    ContainerImageReview,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProjectConfig {
    #[serde(skip_serializing_if = "Option::is_none", default)]
    pub context: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none", default)]
    pub image: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Project {
    pub id: String,
    pub name: String,
    pub kind: ProjectKind,
    pub config: ProjectConfig,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ProjectInput {
    pub name: String,
    pub kind: ProjectKind,
    pub config: ProjectConfig,
}

async fn pool(db: &DbInstances) -> Result<SqlitePool, String> {
    let map = db.0.read().await;
    match map.get(DB_KEY) {
        Some(DbPool::Sqlite(p)) => Ok(p.clone()),
        None => Err(format!("database '{DB_KEY}' is not loaded")),
    }
}

pub async fn list_projects(db: &DbInstances) -> Result<Vec<Project>, String> {
    let pool = pool(db).await?;
    let rows: Vec<(String, String, String, String, String, String)> = sqlx::query_as(
        "SELECT id, name, kind, config, created_at, updated_at \
         FROM projects ORDER BY created_at ASC",
    )
    .fetch_all(&pool)
    .await
    .map_err(|e| e.to_string())?;
    rows.into_iter().map(row_to_project).collect()
}

pub async fn get_project(db: &DbInstances, id: &str) -> Result<Option<Project>, String> {
    let pool = pool(db).await?;
    let row: Option<(String, String, String, String, String, String)> = sqlx::query_as(
        "SELECT id, name, kind, config, created_at, updated_at \
         FROM projects WHERE id = ?1",
    )
    .bind(id)
    .fetch_optional(&pool)
    .await
    .map_err(|e| e.to_string())?;
    row.map(row_to_project).transpose()
}

pub async fn create_project(db: &DbInstances, input: ProjectInput) -> Result<Project, String> {
    let name = input.name.trim().to_string();
    if name.is_empty() {
        return Err("name must not be empty".into());
    }
    let config = normalize_for_kind(input.kind, input.config)?;

    let id = Uuid::new_v4().to_string();
    let kind = kind_str(input.kind);
    let config_json = serde_json::to_string(&config).map_err(|e| e.to_string())?;
    let now = Utc::now().to_rfc3339();

    let pool = pool(db).await?;
    sqlx::query(
        "INSERT INTO projects (id, name, kind, config, created_at, updated_at) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6)",
    )
    .bind(&id)
    .bind(&name)
    .bind(kind)
    .bind(&config_json)
    .bind(&now)
    .bind(&now)
    .execute(&pool)
    .await
    .map_err(|e| e.to_string())?;

    Ok(Project {
        id,
        name,
        kind: input.kind,
        config,
        created_at: now.clone(),
        updated_at: now,
    })
}

pub async fn update_project(
    db: &DbInstances,
    id: String,
    input: ProjectInput,
) -> Result<Project, String> {
    let name = input.name.trim().to_string();
    if name.is_empty() {
        return Err("name must not be empty".into());
    }
    let config = normalize_for_kind(input.kind, input.config)?;

    let pool = pool(db).await?;
    let kind = kind_str(input.kind);
    let config_json = serde_json::to_string(&config).map_err(|e| e.to_string())?;
    let now = Utc::now().to_rfc3339();

    let affected = sqlx::query(
        "UPDATE projects SET name = ?1, kind = ?2, config = ?3, updated_at = ?4 \
         WHERE id = ?5",
    )
    .bind(&name)
    .bind(kind)
    .bind(&config_json)
    .bind(&now)
    .bind(&id)
    .execute(&pool)
    .await
    .map_err(|e| e.to_string())?
    .rows_affected();

    if affected == 0 {
        return Err(format!("project '{id}' not found"));
    }

    get_project(db, &id)
        .await?
        .ok_or_else(|| format!("project '{id}' disappeared after update"))
}

pub async fn delete_project(db: &DbInstances, id: &str) -> Result<(), String> {
    let pool = pool(db).await?;
    let affected = sqlx::query("DELETE FROM projects WHERE id = ?1")
        .bind(id)
        .execute(&pool)
        .await
        .map_err(|e| e.to_string())?
        .rows_affected();
    if affected == 0 {
        return Err(format!("project '{id}' not found"));
    }
    Ok(())
}

fn kind_str(kind: ProjectKind) -> &'static str {
    match kind {
        ProjectKind::KubernetesClusterReview => "kubernetesClusterReview",
        ProjectKind::ContainerImageReview => "containerImageReview",
    }
}

fn kind_from_str(s: &str) -> Result<ProjectKind, String> {
    match s {
        "kubernetesClusterReview" => Ok(ProjectKind::KubernetesClusterReview),
        "containerImageReview" => Ok(ProjectKind::ContainerImageReview),
        other => Err(format!("unknown project kind '{other}' in db")),
    }
}

fn normalize_for_kind(kind: ProjectKind, config: ProjectConfig) -> Result<ProjectConfig, String> {
    match kind {
        ProjectKind::KubernetesClusterReview => {
            if let Some(c) = &config.context {
                if c.trim().is_empty() {
                    return Err("context must not be empty when provided".into());
                }
            }
            Ok(ProjectConfig {
                context: config.context.map(|c| c.trim().to_string()),
                image: None,
            })
        }
        ProjectKind::ContainerImageReview => {
            let image = config
                .image
                .as_ref()
                .map(|s| s.trim().to_string())
                .filter(|s| !s.is_empty())
                .ok_or_else(|| "image must not be empty".to_string())?;
            Ok(ProjectConfig {
                context: None,
                image: Some(image),
            })
        }
    }
}

fn row_to_project(
    row: (String, String, String, String, String, String),
) -> Result<Project, String> {
    let (id, name, kind_str, config_str, created_at, updated_at) = row;
    let kind = kind_from_str(&kind_str)?;
    let config: ProjectConfig =
        serde_json::from_str(&config_str).map_err(|e| format!("invalid config in db: {e}"))?;
    Ok(Project {
        id,
        name,
        kind,
        config,
        created_at,
        updated_at,
    })
}
