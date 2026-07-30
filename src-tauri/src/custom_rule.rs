//! User-authored custom rules stored in SQLite.
//!
//! A rule is a simple matcher: given a Kubernetes resource type, a field
//! path (e.g. `spec.containers[*].securityContext.privileged`), and an
//! expected value, flag any resource that has a value equal to the
//! expected one at that path. See `evaluate_field_path` and
//! `value_matches` for the path syntax.

use chrono::Utc;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sqlx::SqlitePool;
use tauri_plugin_sql::{DbInstances, DbPool};
use uuid::Uuid;

use crate::builtin_rules::builtin_rules;
use crate::rules::Severity;

const DB_KEY: &str = "sqlite:corsair.db";

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CustomRule {
    pub id: String,
    pub title: String,
    pub description: String,
    pub severity: Severity,
    pub resource_type: String,
    pub field_path: String,
    pub expected_value: String,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct CustomRuleInput {
    pub title: String,
    pub description: String,
    pub severity: Severity,
    pub resource_type: String,
    pub field_path: String,
    pub expected_value: String,
}

async fn pool(db: &DbInstances) -> Result<SqlitePool, String> {
    let map = db.0.read().await;
    match map.get(DB_KEY) {
        Some(DbPool::Sqlite(p)) => Ok(p.clone()),
        None => Err(format!("database '{DB_KEY}' is not loaded")),
    }
}

pub async fn list_rules(db: &DbInstances) -> Result<Vec<CustomRule>, String> {
    let pool = pool(db).await?;
    let rows: Vec<(
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
    )> = sqlx::query_as(
        "SELECT id, title, description, severity, resource_type, field_path, expected_value, created_at, updated_at \
         FROM custom_rules ORDER BY created_at ASC",
    )
    .fetch_all(&pool)
    .await
    .map_err(|e| e.to_string())?;
    rows.into_iter().map(row_to_rule).collect()
}

pub async fn create_rule(
    db: &DbInstances,
    input: CustomRuleInput,
) -> Result<CustomRule, String> {
    let title = input.title.trim();
    let description = input.description.trim();
    let resource_type = input.resource_type.trim();
    let field_path = input.field_path.trim();
    let expected_value = input.expected_value;

    if title.is_empty() {
        return Err("title must not be empty".into());
    }
    if resource_type.is_empty() {
        return Err("resource_type must not be empty".into());
    }
    if field_path.is_empty() {
        return Err("field_path must not be empty".into());
    }

    let id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let severity = severity_str(input.severity);

    let pool = pool(db).await?;
    sqlx::query(
        "INSERT INTO custom_rules \
         (id, title, description, severity, resource_type, field_path, expected_value, created_at, updated_at) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)",
    )
    .bind(&id)
    .bind(title)
    .bind(description)
    .bind(severity)
    .bind(resource_type)
    .bind(field_path)
    .bind(&expected_value)
    .bind(&now)
    .bind(&now)
    .execute(&pool)
    .await
    .map_err(|e| e.to_string())?;

    Ok(CustomRule {
        id,
        title: title.to_string(),
        description: description.to_string(),
        severity: input.severity,
        resource_type: resource_type.to_string(),
        field_path: field_path.to_string(),
        expected_value,
        created_at: now.clone(),
        updated_at: now,
    })
}

pub async fn update_rule(
    db: &DbInstances,
    id: String,
    input: CustomRuleInput,
) -> Result<CustomRule, String> {
    let title = input.title.trim();
    let description = input.description.trim();
    let resource_type = input.resource_type.trim();
    let field_path = input.field_path.trim();
    let expected_value = input.expected_value;

    if title.is_empty() {
        return Err("title must not be empty".into());
    }
    if resource_type.is_empty() {
        return Err("resource_type must not be empty".into());
    }
    if field_path.is_empty() {
        return Err("field_path must not be empty".into());
    }

    let pool = pool(db).await?;
    let now = Utc::now().to_rfc3339();
    let severity = severity_str(input.severity);

    let affected = sqlx::query(
        "UPDATE custom_rules SET \
         title = ?1, description = ?2, severity = ?3, resource_type = ?4, \
         field_path = ?5, expected_value = ?6, updated_at = ?7 \
         WHERE id = ?8",
    )
    .bind(title)
    .bind(description)
    .bind(severity)
    .bind(resource_type)
    .bind(field_path)
    .bind(&expected_value)
    .bind(&now)
    .bind(&id)
    .execute(&pool)
    .await
    .map_err(|e| e.to_string())?
    .rows_affected();

    if affected == 0 {
        return Err(format!("custom rule '{id}' not found"));
    }

    get_rule(db, &id)
        .await?
        .ok_or_else(|| format!("custom rule '{id}' disappeared after update"))
}

pub async fn delete_rule(db: &DbInstances, id: &str) -> Result<(), String> {
    let pool = pool(db).await?;
    let affected = sqlx::query("DELETE FROM custom_rules WHERE id = ?1")
        .bind(id)
        .execute(&pool)
        .await
        .map_err(|e| e.to_string())?
        .rows_affected();
    if affected == 0 {
        return Err(format!("custom rule '{id}' not found"));
    }
    Ok(())
}

pub async fn get_rule(db: &DbInstances, id: &str) -> Result<Option<CustomRule>, String> {
    let pool = pool(db).await?;
    let row: Option<(
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
    )> = sqlx::query_as(
        "SELECT id, title, description, severity, resource_type, field_path, expected_value, created_at, updated_at \
         FROM custom_rules WHERE id = ?1",
    )
    .bind(id)
    .fetch_optional(&pool)
    .await
    .map_err(|e| e.to_string())?;
    row.map(row_to_rule).transpose()
}

fn severity_str(severity: Severity) -> &'static str {
    match severity {
        Severity::Critical => "critical",
        Severity::High => "high",
        Severity::Medium => "medium",
        Severity::Low => "low",
    }
}

fn severity_from_str(s: &str) -> Result<Severity, String> {
    match s {
        "critical" => Ok(Severity::Critical),
        "high" => Ok(Severity::High),
        "medium" => Ok(Severity::Medium),
        "low" => Ok(Severity::Low),
        other => Err(format!("unknown severity '{other}' in db")),
    }
}

fn row_to_rule(
    row: (
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
        String,
    ),
) -> Result<CustomRule, String> {
    let (id, title, description, severity, resource_type, field_path, expected_value, created_at, updated_at) = row;
    Ok(CustomRule {
        id,
        title,
        description,
        severity: severity_from_str(&severity)?,
        resource_type,
        field_path,
        expected_value,
        created_at,
        updated_at,
    })
}

/// Walk a dotted path through a JSON value. `[*]` iterates over array
/// elements. Returns every leaf value the path matches (zero, one, or many).
///
/// Examples:
/// - `metadata.name` -> `["default"]` (or whatever the name is)
/// - `spec.containers[*].securityContext.privileged` -> one entry per container
/// - `spec.containers[*].securityContext` -> the full SecurityContext object per container
pub fn evaluate_field_path(value: &Value, path: &str) -> Vec<Value> {
    let mut current = vec![value.clone()];
    for part in path.split('.') {
        let mut next = Vec::new();
        for v in &current {
            if part == "[*]" {
                if let Some(arr) = v.as_array() {
                    next.extend(arr.iter().cloned());
                }
            } else if let Some(obj) = v.as_object() {
                if let Some(child) = obj.get(part) {
                    next.push(child.clone());
                }
            }
        }
        current = next;
        if current.is_empty() {
            break;
        }
    }
    current
}

/// Coerce a JSON value to a string and compare against `expected`. Booleans,
/// numbers, and null are compared via their canonical string form
/// (`"true"`, `"42"`, `"null"`), so the user's `expected_value` is the
/// one source of truth regardless of the JSON type.
pub fn value_matches(value: &Value, expected: &str) -> bool {
    match value {
        Value::String(s) => s == expected,
        Value::Bool(b) => b.to_string() == expected,
        Value::Number(n) => n.to_string() == expected,
        Value::Null => expected == "null",
        _ => false,
    }
}

/// Aggregate user-authored and built-in rules. Built-ins come from
/// `builtin_rules`; user-authored from the DB. Both are returned as
/// `CustomRule` so the engine and the UI treat them uniformly.
pub async fn all_rules(db: &DbInstances) -> Result<Vec<CustomRule>, String> {
    let mut out = builtin_rules();
    out.extend(list_rules(db).await?);
    Ok(out)
}

/// Public so `lib.rs` can list built-ins for the UI.
pub fn builtin() -> Vec<CustomRule> {
    builtin_rules()
}
