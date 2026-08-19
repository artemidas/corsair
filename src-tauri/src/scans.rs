//! Persisted scan runs and their findings.
//!
//! Each `run_scan` for a project writes a `scans` row and one `findings` row
//! per hit. Failed evaluations are stored too (`status = failed`) so the
//! project keeps a history even when the cluster call blows up.

use chrono::Utc;
use serde::{Deserialize, Serialize};
use sqlx::SqlitePool;
use tauri_plugin_sql::{DbInstances, DbPool};
use uuid::Uuid;

use crate::rules::{Finding, Severity};

const DB_KEY: &str = "sqlite:corsair.db";

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ScanStatus {
    Completed,
    Failed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Scan {
    pub id: String,
    pub project_id: String,
    pub status: ScanStatus,
    pub context: Option<String>,
    pub error: Option<String>,
    pub finding_count: i64,
    pub started_at: String,
    pub finished_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct StoredFinding {
    pub id: String,
    pub scan_id: String,
    pub rule_id: String,
    pub rule_title: String,
    pub severity: Severity,
    pub resource_kind: String,
    pub resource_name: String,
    pub namespace: Option<String>,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ScanResult {
    pub scan: Scan,
    pub findings: Vec<StoredFinding>,
}

type ScanRow = (
    String,
    String,
    String,
    Option<String>,
    Option<String>,
    i64,
    String,
    String,
);

type FindingRow = (
    String,
    String,
    String,
    String,
    String,
    String,
    String,
    Option<String>,
    String,
);

async fn pool(db: &DbInstances) -> Result<SqlitePool, String> {
    let map = db.0.read().await;
    match map.get(DB_KEY) {
        Some(DbPool::Sqlite(p)) => Ok(p.clone()),
        None => Err(format!("database '{DB_KEY}' is not loaded")),
    }
}

pub async fn persist_scan(
    db: &DbInstances,
    project_id: String,
    context: Option<String>,
    outcome: Result<Vec<Finding>, String>,
) -> Result<ScanResult, String> {
    let pool = pool(db).await?;
    let id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();

    match outcome {
        Ok(findings) => {
            let stored: Vec<StoredFinding> = findings
                .into_iter()
                .map(|finding| StoredFinding {
                    id: Uuid::new_v4().to_string(),
                    scan_id: id.clone(),
                    rule_id: finding.rule_id,
                    rule_title: finding.rule_title,
                    severity: finding.severity,
                    resource_kind: finding.resource_kind,
                    resource_name: finding.resource_name,
                    namespace: finding.namespace,
                    message: finding.message,
                })
                .collect();

            let mut tx = pool.begin().await.map_err(|e| e.to_string())?;
            sqlx::query(
                "INSERT INTO scans \
                 (id, project_id, status, context, error, finding_count, started_at, finished_at) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
            )
            .bind(&id)
            .bind(&project_id)
            .bind("completed")
            .bind(&context)
            .bind(None::<String>)
            .bind(stored.len() as i64)
            .bind(&now)
            .bind(&now)
            .execute(&mut *tx)
            .await
            .map_err(|e| e.to_string())?;

            for finding in &stored {
                sqlx::query(
                    "INSERT INTO findings \
                     (id, scan_id, rule_id, rule_title, severity, resource_kind, \
                      resource_name, namespace, message) \
                     VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)",
                )
                .bind(&finding.id)
                .bind(&finding.scan_id)
                .bind(&finding.rule_id)
                .bind(&finding.rule_title)
                .bind(finding.severity.as_str())
                .bind(&finding.resource_kind)
                .bind(&finding.resource_name)
                .bind(&finding.namespace)
                .bind(&finding.message)
                .execute(&mut *tx)
                .await
                .map_err(|e| e.to_string())?;
            }

            tx.commit().await.map_err(|e| e.to_string())?;

            Ok(ScanResult {
                scan: Scan {
                    id,
                    project_id,
                    status: ScanStatus::Completed,
                    context,
                    error: None,
                    finding_count: stored.len() as i64,
                    started_at: now.clone(),
                    finished_at: now,
                },
                findings: stored,
            })
        }
        Err(error) => {
            sqlx::query(
                "INSERT INTO scans \
                 (id, project_id, status, context, error, finding_count, started_at, finished_at) \
                 VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8)",
            )
            .bind(&id)
            .bind(&project_id)
            .bind("failed")
            .bind(&context)
            .bind(&error)
            .bind(0_i64)
            .bind(&now)
            .bind(&now)
            .execute(&pool)
            .await
            .map_err(|e| e.to_string())?;

            Ok(ScanResult {
                scan: Scan {
                    id,
                    project_id,
                    status: ScanStatus::Failed,
                    context,
                    error: Some(error),
                    finding_count: 0,
                    started_at: now.clone(),
                    finished_at: now,
                },
                findings: vec![],
            })
        }
    }
}

pub async fn list_scans(db: &DbInstances, project_id: &str) -> Result<Vec<Scan>, String> {
    let pool = pool(db).await?;
    let rows: Vec<ScanRow> = sqlx::query_as(
        "SELECT id, project_id, status, context, error, finding_count, started_at, finished_at \
         FROM scans WHERE project_id = ?1 ORDER BY started_at DESC",
    )
    .bind(project_id)
    .fetch_all(&pool)
    .await
    .map_err(|e| e.to_string())?;
    rows.into_iter().map(row_to_scan).collect()
}

pub async fn get_scan(db: &DbInstances, id: &str) -> Result<Option<Scan>, String> {
    let pool = pool(db).await?;
    let row: Option<ScanRow> = sqlx::query_as(
        "SELECT id, project_id, status, context, error, finding_count, started_at, finished_at \
         FROM scans WHERE id = ?1",
    )
    .bind(id)
    .fetch_optional(&pool)
    .await
    .map_err(|e| e.to_string())?;
    row.map(row_to_scan).transpose()
}

pub async fn list_scan_findings(
    db: &DbInstances,
    scan_id: &str,
) -> Result<Vec<StoredFinding>, String> {
    let pool = pool(db).await?;
    let rows: Vec<FindingRow> = sqlx::query_as(
        "SELECT id, scan_id, rule_id, rule_title, severity, resource_kind, \
         resource_name, namespace, message \
         FROM findings WHERE scan_id = ?1 ORDER BY rule_id, resource_name",
    )
    .bind(scan_id)
    .fetch_all(&pool)
    .await
    .map_err(|e| e.to_string())?;
    rows.into_iter().map(row_to_finding).collect()
}

pub async fn delete_for_project(db: &DbInstances, project_id: &str) -> Result<(), String> {
    let pool = pool(db).await?;
    sqlx::query(
        "DELETE FROM findings WHERE scan_id IN (SELECT id FROM scans WHERE project_id = ?1)",
    )
    .bind(project_id)
    .execute(&pool)
    .await
    .map_err(|e| e.to_string())?;
    sqlx::query("DELETE FROM scans WHERE project_id = ?1")
        .bind(project_id)
        .execute(&pool)
        .await
        .map_err(|e| e.to_string())?;
    Ok(())
}

fn row_to_scan(row: ScanRow) -> Result<Scan, String> {
    let (id, project_id, status, context, error, finding_count, started_at, finished_at) = row;
    Ok(Scan {
        id,
        project_id,
        status: status_from_str(&status)?,
        context,
        error,
        finding_count,
        started_at,
        finished_at,
    })
}

fn row_to_finding(row: FindingRow) -> Result<StoredFinding, String> {
    let (
        id,
        scan_id,
        rule_id,
        rule_title,
        severity,
        resource_kind,
        resource_name,
        namespace,
        message,
    ) = row;
    Ok(StoredFinding {
        id,
        scan_id,
        rule_id,
        rule_title,
        severity: Severity::parse(&severity)?,
        resource_kind,
        resource_name,
        namespace,
        message,
    })
}

fn status_from_str(s: &str) -> Result<ScanStatus, String> {
    match s {
        "completed" => Ok(ScanStatus::Completed),
        "failed" => Ok(ScanStatus::Failed),
        other => Err(format!("unknown scan status '{other}' in db")),
    }
}
