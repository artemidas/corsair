//! Rules stored in SQLite.
//!
//! A rule is a matcher: given a Kubernetes resource type, a field path
//! (e.g. `spec.containers[*].securityContext.privileged`), an operator,
//! and (for some operators) an expected value, flag any resource whose
//! resolved leaves satisfy the operator. See `evaluate_field_path` and
//! `evaluate_operator`.

use chrono::Utc;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use sqlx::SqlitePool;
use tauri_plugin_sql::{DbInstances, DbPool};
use uuid::Uuid;

use crate::rules::Severity;

const DB_KEY: &str = "sqlite:corsair.db";

const SELECT_COLS: &str = "id, rule_id, title, description, severity, resource_type, \
     field_path, operator, expected_value, import_id, created_at, updated_at";

type RuleRow = (
    String,
    Option<String>,
    String,
    String,
    String,
    String,
    String,
    String,
    String,
    Option<String>,
    String,
    String,
);

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq, Default)]
#[serde(rename_all = "camelCase")]
pub enum Operator {
    #[default]
    Equals,
    NotEquals,
    Present,
    Absent,
    ArrayExcludes,
}

impl Operator {
    pub fn needs_expected_value(self) -> bool {
        matches!(
            self,
            Operator::Equals | Operator::NotEquals | Operator::ArrayExcludes
        )
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Rule {
    pub id: String,
    pub rule_id: String,
    pub title: String,
    pub description: String,
    pub severity: Severity,
    pub resource_type: String,
    pub field_path: String,
    pub operator: Operator,
    pub expected_value: String,
    pub import_id: Option<String>,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct RuleInput {
    pub title: String,
    pub description: String,
    pub severity: Severity,
    pub resource_type: String,
    pub field_path: String,
    #[serde(default)]
    pub operator: Operator,
    pub expected_value: String,
}

async fn pool(db: &DbInstances) -> Result<SqlitePool, String> {
    let map = db.0.read().await;
    match map.get(DB_KEY) {
        Some(DbPool::Sqlite(p)) => Ok(p.clone()),
        None => Err(format!("database '{DB_KEY}' is not loaded")),
    }
}

pub async fn list_rules(db: &DbInstances) -> Result<Vec<Rule>, String> {
    let pool = pool(db).await?;
    list_rules_pool(&pool).await
}

pub(crate) async fn list_rules_pool(pool: &SqlitePool) -> Result<Vec<Rule>, String> {
    backfill_rule_ids(pool).await?;
    let rows: Vec<RuleRow> = sqlx::query_as(&format!(
        "SELECT {SELECT_COLS} FROM custom_rules ORDER BY created_at ASC"
    ))
    .fetch_all(pool)
    .await
    .map_err(|e| e.to_string())?;
    rows.into_iter().map(row_to_rule).collect()
}

pub async fn create_rule(
    db: &DbInstances,
    input: RuleInput,
) -> Result<Rule, String> {
    let pool = pool(db).await?;
    insert_rule(&pool, input, None).await
}

pub async fn create_imported_rule(
    db: &DbInstances,
    input: RuleInput,
    import_id: Option<String>,
) -> Result<Rule, String> {
    let pool = pool(db).await?;
    insert_rule(&pool, input, import_id).await
}

pub(crate) async fn insert_rule(
    pool: &SqlitePool,
    input: RuleInput,
    import_id: Option<String>,
) -> Result<Rule, String> {
    let prepared = prepare_input(input)?;
    let id = Uuid::new_v4().to_string();
    let now = Utc::now().to_rfc3339();
    let import_id = normalize_import_id(import_id);
    let rule_id = allocate_rule_id(pool, &prepared.resource_type, import_id.as_deref()).await?;

    sqlx::query(
        "INSERT INTO custom_rules \
         (id, rule_id, title, description, severity, resource_type, field_path, operator, expected_value, import_id, created_at, updated_at) \
         VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12)",
    )
    .bind(&id)
    .bind(&rule_id)
    .bind(&prepared.title)
    .bind(&prepared.description)
    .bind(severity_str(prepared.severity))
    .bind(&prepared.resource_type)
    .bind(&prepared.field_path)
    .bind(operator_str(prepared.operator))
    .bind(&prepared.expected_value)
    .bind(&import_id)
    .bind(&now)
    .bind(&now)
    .execute(pool)
    .await
    .map_err(|e| e.to_string())?;

    Ok(Rule {
        id,
        rule_id,
        title: prepared.title,
        description: prepared.description,
        severity: prepared.severity,
        resource_type: prepared.resource_type,
        field_path: prepared.field_path,
        operator: prepared.operator,
        expected_value: prepared.expected_value,
        import_id,
        created_at: now.clone(),
        updated_at: now,
    })
}

pub async fn update_rule(
    db: &DbInstances,
    id: String,
    input: RuleInput,
) -> Result<Rule, String> {
    let pool = pool(db).await?;
    update_rule_pool(&pool, &id, input).await
}

pub(crate) async fn update_rule_pool(
    pool: &SqlitePool,
    id: &str,
    input: RuleInput,
) -> Result<Rule, String> {
    let prepared = prepare_input(input)?;
    let now = Utc::now().to_rfc3339();

    let affected = sqlx::query(
        "UPDATE custom_rules SET \
         title = ?1, description = ?2, severity = ?3, resource_type = ?4, \
         field_path = ?5, operator = ?6, expected_value = ?7, updated_at = ?8 \
         WHERE id = ?9",
    )
    .bind(&prepared.title)
    .bind(&prepared.description)
    .bind(severity_str(prepared.severity))
    .bind(&prepared.resource_type)
    .bind(&prepared.field_path)
    .bind(operator_str(prepared.operator))
    .bind(&prepared.expected_value)
    .bind(&now)
    .bind(id)
    .execute(pool)
    .await
    .map_err(|e| e.to_string())?
    .rows_affected();

    if affected == 0 {
        return Err(format!("rule '{id}' not found"));
    }

    get_rule_pool(pool, id)
        .await?
        .ok_or_else(|| format!("rule '{id}' disappeared after update"))
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
        return Err(format!("rule '{id}' not found"));
    }
    Ok(())
}

pub async fn delete_imported_rules(db: &DbInstances) -> Result<(), String> {
    let pool = pool(db).await?;
    delete_imported_rules_pool(&pool).await
}

pub(crate) async fn delete_imported_rules_pool(pool: &SqlitePool) -> Result<(), String> {
    sqlx::query("DELETE FROM custom_rules WHERE import_id IS NOT NULL")
        .execute(pool)
        .await
        .map_err(|e| e.to_string())?;
    Ok(())
}

#[allow(dead_code)]
pub async fn get_rule(db: &DbInstances, id: &str) -> Result<Option<Rule>, String> {
    let pool = pool(db).await?;
    get_rule_pool(&pool, id).await
}

pub(crate) async fn get_rule_pool(
    pool: &SqlitePool,
    id: &str,
) -> Result<Option<Rule>, String> {
    let row: Option<RuleRow> = sqlx::query_as(&format!(
        "SELECT {SELECT_COLS} FROM custom_rules WHERE id = ?1"
    ))
    .bind(id)
    .fetch_optional(pool)
    .await
    .map_err(|e| e.to_string())?;
    row.map(row_to_rule).transpose()
}

pub async fn get_rule_by_import_id(
    db: &DbInstances,
    import_id: &str,
) -> Result<Option<Rule>, String> {
    let pool = pool(db).await?;
    get_rule_by_import_id_pool(&pool, import_id).await
}

pub(crate) async fn get_rule_by_import_id_pool(
    pool: &SqlitePool,
    import_id: &str,
) -> Result<Option<Rule>, String> {
    let row: Option<RuleRow> = sqlx::query_as(&format!(
        "SELECT {SELECT_COLS} FROM custom_rules WHERE import_id = ?1"
    ))
    .bind(import_id)
    .fetch_optional(pool)
    .await
    .map_err(|e| e.to_string())?;
    row.map(row_to_rule).transpose()
}

struct PreparedInput {
    title: String,
    description: String,
    severity: Severity,
    resource_type: String,
    field_path: String,
    operator: Operator,
    expected_value: String,
}

fn prepare_input(input: RuleInput) -> Result<PreparedInput, String> {
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
    if input.operator.needs_expected_value() && expected_value.trim().is_empty() {
        return Err("expected_value is required for this operator".into());
    }

    Ok(PreparedInput {
        title: title.to_string(),
        description: description.to_string(),
        severity: input.severity,
        resource_type: resource_type.to_string(),
        field_path: field_path.to_string(),
        operator: input.operator,
        expected_value,
    })
}

fn normalize_import_id(import_id: Option<String>) -> Option<String> {
    import_id.and_then(|s| {
        let t = s.trim();
        if t.is_empty() {
            None
        } else {
            Some(t.to_string())
        }
    })
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

fn operator_str(operator: Operator) -> &'static str {
    match operator {
        Operator::Equals => "equals",
        Operator::NotEquals => "notEquals",
        Operator::Present => "present",
        Operator::Absent => "absent",
        Operator::ArrayExcludes => "arrayExcludes",
    }
}

fn operator_from_str(s: &str) -> Result<Operator, String> {
    match s {
        "equals" => Ok(Operator::Equals),
        "notEquals" => Ok(Operator::NotEquals),
        "present" => Ok(Operator::Present),
        "absent" => Ok(Operator::Absent),
        "arrayExcludes" => Ok(Operator::ArrayExcludes),
        other => Err(format!("unknown operator '{other}' in db")),
    }
}

fn rule_id_prefix(resource_type: &str) -> &'static str {
    match resource_type {
        "Pod" => "POD",
        "ServiceAccount" => "SA",
        "Role" => "ROLE",
        "ClusterRole" => "CR",
        "RoleBinding" => "RB",
        "ClusterRoleBinding" => "CRB",
        _ => "RULE",
    }
}

fn parse_rule_id_seq(rule_id: &str, prefix: &str) -> Option<u32> {
    let rest = rule_id.strip_prefix(prefix)?;
    if rest.is_empty() || !rest.chars().all(|c| c.is_ascii_digit()) {
        return None;
    }
    rest.parse().ok()
}

fn format_rule_id(prefix: &str, seq: u32) -> String {
    format!("{prefix}{seq:02}")
}

async fn existing_rule_ids(pool: &SqlitePool) -> Result<Vec<String>, String> {
    sqlx::query_scalar("SELECT rule_id FROM custom_rules WHERE rule_id IS NOT NULL AND rule_id != ''")
        .fetch_all(pool)
        .await
        .map_err(|e| e.to_string())
}

async fn allocate_rule_id(
    pool: &SqlitePool,
    resource_type: &str,
    preferred: Option<&str>,
) -> Result<String, String> {
    let existing = existing_rule_ids(pool).await?;
    if let Some(preferred) = preferred.map(str::trim).filter(|s| !s.is_empty()) {
        if !existing.iter().any(|id| id == preferred) {
            return Ok(preferred.to_string());
        }
    }

    let prefix = rule_id_prefix(resource_type);
    let next = existing
        .iter()
        .filter_map(|id| parse_rule_id_seq(id, prefix))
        .max()
        .unwrap_or(0)
        + 1;
    Ok(format_rule_id(prefix, next))
}

async fn backfill_rule_ids(pool: &SqlitePool) -> Result<(), String> {
    let missing: Vec<(String, String, Option<String>)> = sqlx::query_as(
        "SELECT id, resource_type, import_id FROM custom_rules \
         WHERE rule_id IS NULL OR rule_id = '' \
         ORDER BY created_at ASC, id ASC",
    )
    .fetch_all(pool)
    .await
    .map_err(|e| e.to_string())?;

    for (id, resource_type, import_id) in missing {
        let rule_id = allocate_rule_id(pool, &resource_type, import_id.as_deref()).await?;
        sqlx::query("UPDATE custom_rules SET rule_id = ?1 WHERE id = ?2")
            .bind(&rule_id)
            .bind(&id)
            .execute(pool)
            .await
            .map_err(|e| e.to_string())?;
    }
    Ok(())
}

fn row_to_rule(row: RuleRow) -> Result<Rule, String> {
    let (
        id,
        rule_id,
        title,
        description,
        severity,
        resource_type,
        field_path,
        operator,
        expected_value,
        import_id,
        created_at,
        updated_at,
    ) = row;
    Ok(Rule {
        id,
        rule_id: rule_id.unwrap_or_default(),
        title,
        description,
        severity: severity_from_str(&severity)?,
        resource_type,
        field_path,
        operator: operator_from_str(&operator)?,
        expected_value,
        import_id,
        created_at,
        updated_at,
    })
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub(crate) enum PathSegment {
    Field(String),
    Iterate,
}

pub(crate) fn parse_path(path: &str) -> Vec<PathSegment> {
    let mut segments = Vec::new();
    for raw in path.split('.') {
        if raw == "[*]" {
            segments.push(PathSegment::Iterate);
        } else if let Some(field) = raw.strip_suffix("[*]") {
            segments.push(PathSegment::Field(field.to_string()));
            segments.push(PathSegment::Iterate);
        } else {
            segments.push(PathSegment::Field(raw.to_string()));
        }
    }
    segments
}

/// Walk a dotted path through a JSON value. `[*]` iterates over array
/// elements, either as its own segment or attached to a field name
/// (`containers[*]`). Missing objects contribute `Null` so cardinality
/// is preserved across branches.
pub fn evaluate_field_path(value: &Value, path: &str) -> Vec<Value> {
    let mut current = vec![value.clone()];
    for seg in parse_path(path) {
        let mut next = Vec::new();
        for v in &current {
            match &seg {
                PathSegment::Iterate => {
                    if let Some(arr) = v.as_array() {
                        next.extend(arr.iter().cloned());
                    }
                }
                PathSegment::Field(name) => match v {
                    Value::Object(obj) => {
                        next.push(obj.get(name).cloned().unwrap_or(Value::Null))
                    }
                    _ => next.push(Value::Null),
                },
            }
        }
        current = next;
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

fn is_present(v: &Value) -> bool {
    match v {
        Value::Null => false,
        Value::Array(a) => !a.is_empty(),
        Value::String(s) => !s.is_empty(),
        _ => true,
    }
}

fn array_contains(v: &Value, expected: &str) -> bool {
    match v.as_array() {
        Some(arr) => arr.iter().any(|el| value_matches(el, expected)),
        None => false,
    }
}

/// Returns true when the operator's "raise a finding" condition holds
/// for the resolved leaves of one resource.
pub fn evaluate_operator(leaves: &[Value], operator: Operator, expected: &str) -> bool {
    match operator {
        Operator::Equals => leaves.iter().any(|v| value_matches(v, expected)),
        Operator::NotEquals => {
            !leaves.is_empty() && leaves.iter().any(|v| !value_matches(v, expected))
        }
        Operator::Present => leaves.iter().any(is_present),
        Operator::Absent => leaves.is_empty() || leaves.iter().any(|v| !is_present(v)),
        Operator::ArrayExcludes => {
            leaves.is_empty() || leaves.iter().any(|v| !array_contains(v, expected))
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::rules::Severity;
    use serde_json::json;
    use sqlx::SqlitePool;

    #[test]
    fn parse_attached_and_standalone_brackets() {
        assert_eq!(
            parse_path("spec.containers[*].foo"),
            vec![
                PathSegment::Field("spec".into()),
                PathSegment::Field("containers".into()),
                PathSegment::Iterate,
                PathSegment::Field("foo".into()),
            ]
        );
        assert_eq!(
            parse_path("spec.containers.[*].foo"),
            vec![
                PathSegment::Field("spec".into()),
                PathSegment::Field("containers".into()),
                PathSegment::Iterate,
                PathSegment::Field("foo".into()),
            ]
        );
    }

    fn two_container_pod() -> Value {
        json!({
            "spec": {
                "containers": [
                    {
                        "resources": { "limits": { "cpu": "500m" } },
                        "securityContext": { "capabilities": { "drop": ["ALL"] } }
                    },
                    {
                        "name": "bare"
                    }
                ]
            }
        })
    }

    #[test]
    fn missing_intermediate_preserves_null_leaf() {
        let leaves = evaluate_field_path(
            &two_container_pod(),
            "spec.containers[*].resources.limits.cpu",
        );
        assert_eq!(leaves, vec![json!("500m"), Value::Null]);
    }

    #[test]
    fn array_without_iterate_is_the_leaf() {
        let v = json!({ "spec": { "containers": [{ "name": "a" }, { "name": "b" }] } });
        let leaves = evaluate_field_path(&v, "spec.containers");
        assert_eq!(leaves.len(), 1);
        assert!(leaves[0].is_array());
        assert_eq!(leaves[0].as_array().unwrap().len(), 2);
    }

    #[test]
    fn operator_equals() {
        let leaves = vec![json!(true), json!(false)];
        assert!(evaluate_operator(&leaves, Operator::Equals, "true"));
        assert!(!evaluate_operator(&leaves, Operator::Equals, "maybe"));
    }

    #[test]
    fn operator_not_equals() {
        let leaves = vec![json!(true), Value::Null];
        assert!(evaluate_operator(&leaves, Operator::NotEquals, "true"));
        assert!(!evaluate_operator(&[json!(true)], Operator::NotEquals, "true"));
        assert!(!evaluate_operator(&[], Operator::NotEquals, "true"));
    }

    #[test]
    fn operator_present() {
        assert!(evaluate_operator(&[json!("x")], Operator::Present, ""));
        assert!(!evaluate_operator(&[Value::Null], Operator::Present, ""));
        assert!(!evaluate_operator(&[json!("")], Operator::Present, ""));
        assert!(!evaluate_operator(&[json!([])], Operator::Present, ""));
    }

    #[test]
    fn operator_absent_two_containers() {
        let leaves = evaluate_field_path(
            &two_container_pod(),
            "spec.containers[*].resources.limits.cpu",
        );
        assert!(evaluate_operator(&leaves, Operator::Absent, ""));
        assert!(!evaluate_operator(&[json!("500m")], Operator::Absent, ""));
    }

    #[test]
    fn operator_array_excludes_two_containers() {
        let leaves = evaluate_field_path(
            &two_container_pod(),
            "spec.containers[*].securityContext.capabilities.drop",
        );
        assert!(evaluate_operator(&leaves, Operator::ArrayExcludes, "ALL"));
        assert!(!evaluate_operator(
            &[json!(["ALL"])],
            Operator::ArrayExcludes,
            "ALL"
        ));
    }

    #[test]
    fn privileged_equals_matches_fixture_pod() {
        let pod = json!({
            "spec": {
                "hostNetwork": true,
                "hostPID": true,
                "hostIPC": true,
                "containers": [{
                    "securityContext": { "privileged": true }
                }]
            }
        });
        let leaves = evaluate_field_path(
            &pod,
            "spec.containers[*].securityContext.privileged",
        );
        assert!(evaluate_operator(&leaves, Operator::Equals, "true"));
    }

    #[test]
    fn rule_id_seq_parses_and_formats() {
        assert_eq!(parse_rule_id_seq("POD01", "POD"), Some(1));
        assert_eq!(parse_rule_id_seq("POD12", "POD"), Some(12));
        assert_eq!(parse_rule_id_seq("CRB01", "CR"), None);
        assert_eq!(parse_rule_id_seq("CR01", "CR"), Some(1));
        assert_eq!(format_rule_id("POD", 1), "POD01");
        assert_eq!(format_rule_id("SA", 12), "SA12");
    }

    async fn setup_pool() -> SqlitePool {
        let pool = SqlitePool::connect("sqlite::memory:").await.unwrap();
        sqlx::query(
            "CREATE TABLE custom_rules (
                id TEXT PRIMARY KEY,
                rule_id TEXT,
                title TEXT NOT NULL,
                description TEXT NOT NULL,
                severity TEXT NOT NULL,
                resource_type TEXT NOT NULL,
                field_path TEXT NOT NULL,
                operator TEXT NOT NULL DEFAULT 'equals',
                expected_value TEXT NOT NULL,
                import_id TEXT,
                created_at TEXT NOT NULL,
                updated_at TEXT NOT NULL
            )",
        )
        .execute(&pool)
        .await
        .unwrap();
        pool
    }

    fn pod_input(title: &str) -> RuleInput {
        RuleInput {
            title: title.into(),
            description: String::new(),
            severity: Severity::Low,
            resource_type: "Pod".into(),
            field_path: "spec.hostNetwork".into(),
            operator: Operator::Equals,
            expected_value: "true".into(),
        }
    }

    #[tokio::test]
    async fn assigns_sequential_rule_ids_per_resource_type() {
        let pool = setup_pool().await;
        let first = insert_rule(&pool, pod_input("a"), None).await.unwrap();
        let second = insert_rule(&pool, pod_input("b"), None).await.unwrap();
        let sa = insert_rule(
            &pool,
            RuleInput {
                title: "sa".into(),
                description: String::new(),
                severity: Severity::Low,
                resource_type: "ServiceAccount".into(),
                field_path: "metadata.name".into(),
                operator: Operator::Equals,
                expected_value: "default".into(),
            },
            None,
        )
        .await
        .unwrap();

        assert_eq!(first.rule_id, "POD01");
        assert_eq!(second.rule_id, "POD02");
        assert_eq!(sa.rule_id, "SA01");
    }

    #[tokio::test]
    async fn prefers_import_id_when_unused() {
        let pool = setup_pool().await;
        let imported = insert_rule(&pool, pod_input("imported"), Some("POD10".into()))
            .await
            .unwrap();
        let next = insert_rule(&pool, pod_input("next"), None).await.unwrap();
        assert_eq!(imported.rule_id, "POD10");
        assert_eq!(next.rule_id, "POD11");
    }
}
