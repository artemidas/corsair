//! Versioned YAML rule-pack format (v1) for import/export.
//!
//! `notes` and `duplicatesBuiltin` are parsed for human-authored files
//! and discarded — they are never stored.

use serde::{de, Deserialize, Deserializer, Serialize};
use tauri_plugin_sql::DbInstances;

use crate::custom_rule::{
    create_imported_rule, delete_imported_rules, get_rule_by_import_id, list_rules, update_rule,
    Operator, Rule, RuleInput,
};
use crate::rules::Severity;

#[cfg(test)]
use sqlx::SqlitePool;

const PACK_VERSION: u32 = 1;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum MatcherKind {
    Declarative,
    Native,
}

fn default_matcher() -> MatcherKind {
    MatcherKind::Declarative
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RulePack {
    pub version: u32,
    pub rules: Vec<RulePackEntry>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct RulePackEntry {
    pub id: Option<String>,
    pub title: String,
    #[serde(default)]
    pub description: String,
    pub severity: Severity,
    pub resource_type: String,
    #[serde(default = "default_matcher")]
    pub matcher: MatcherKind,
    pub field_path: Option<String>,
    pub operator: Option<Operator>,
    #[serde(default, deserialize_with = "deserialize_scalar_string")]
    pub expected_value: String,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub notes: Option<String>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub duplicates_builtin: Option<String>,
}

#[derive(Debug, Clone, Copy, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum ImportMode {
    Merge,
    Replace,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct SkippedRule {
    pub id: String,
    pub title: String,
    pub reason: String,
}

#[derive(Debug, Default, Serialize, PartialEq, Eq)]
#[serde(rename_all = "camelCase")]
pub struct ImportSummary {
    pub created: u32,
    pub updated: u32,
    pub skipped: Vec<SkippedRule>,
}

pub fn parse_rule_pack(text: &str) -> Result<RulePack, String> {
    let trimmed = text.trim();
    if trimmed.is_empty() {
        return Err("YAML file is empty".into());
    }
    let pack: RulePack =
        serde_yaml::from_str(trimmed).map_err(|e| format!("invalid rule pack YAML: {e}"))?;
    if pack.version != PACK_VERSION {
        return Err(format!("unsupported rule pack version: {}", pack.version));
    }
    Ok(pack)
}

pub fn serialize_rule_pack(rules: &[Rule]) -> Result<String, String> {
    let pack = RulePack {
        version: PACK_VERSION,
        rules: rules.iter().map(to_pack_entry).collect(),
    };
    serde_yaml::to_string(&pack).map_err(|e| e.to_string())
}

fn to_pack_entry(rule: &Rule) -> RulePackEntry {
    RulePackEntry {
        id: Some(
            rule.import_id
                .clone()
                .unwrap_or_else(|| rule.id.clone()),
        ),
        title: rule.title.clone(),
        description: rule.description.clone(),
        severity: rule.severity,
        resource_type: rule.resource_type.clone(),
        matcher: MatcherKind::Declarative,
        field_path: Some(rule.field_path.clone()),
        operator: Some(rule.operator),
        expected_value: rule.expected_value.clone(),
        notes: None,
        duplicates_builtin: None,
    }
}

pub(crate) enum PreparedEntry {
    Skip(SkippedRule),
    Ready {
        import_id: Option<String>,
        input: RuleInput,
    },
}

pub(crate) fn prepare_entry(entry: RulePackEntry) -> PreparedEntry {
    let id = entry
        .id
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .unwrap_or("")
        .to_string();
    let title = entry.title.clone();

    if entry.matcher == MatcherKind::Native {
        return PreparedEntry::Skip(SkippedRule {
            id,
            title,
            reason: "native matcher is not supported".into(),
        });
    }

    let Some(field_path) = entry
        .field_path
        .as_deref()
        .map(str::trim)
        .filter(|s| !s.is_empty())
        .map(str::to_string)
    else {
        return PreparedEntry::Skip(SkippedRule {
            id,
            title,
            reason: "declarative rule is missing fieldPath".into(),
        });
    };

    let Some(operator) = entry.operator else {
        return PreparedEntry::Skip(SkippedRule {
            id,
            title,
            reason: "declarative rule is missing operator".into(),
        });
    };

    if operator.needs_expected_value() && entry.expected_value.trim().is_empty() {
        return PreparedEntry::Skip(SkippedRule {
            id,
            title,
            reason: format!("operator {operator:?} requires a non-empty expectedValue"),
        });
    }

    let import_id = if id.is_empty() { None } else { Some(id) };
    PreparedEntry::Ready {
        import_id,
        input: RuleInput {
            title: entry.title,
            description: entry.description,
            severity: entry.severity,
            resource_type: entry.resource_type,
            field_path,
            operator,
            expected_value: entry.expected_value,
        },
    }
}

fn ensure_yaml_extension(path: String) -> String {
    let lower = path.to_lowercase();
    if lower.ends_with(".yaml") || lower.ends_with(".yml") {
        path
    } else {
        format!("{path}.yaml")
    }
}

pub async fn export_to_path(db: &DbInstances, path: String) -> Result<usize, String> {
    let selected = list_rules(db).await?;
    if selected.is_empty() {
        return Err("no rules to export".into());
    }
    write_pack(&path, &selected)
}

fn write_pack(path: &str, rules: &[Rule]) -> Result<usize, String> {
    let path = ensure_yaml_extension(path.to_string());
    let yaml = serialize_rule_pack(rules)?;
    std::fs::write(&path, yaml).map_err(|e| format!("failed to write '{path}': {e}"))?;
    Ok(rules.len())
}

pub async fn import_from_path(
    db: &DbInstances,
    path: String,
    mode: ImportMode,
) -> Result<ImportSummary, String> {
    let text =
        std::fs::read_to_string(&path).map_err(|e| format!("failed to read '{path}': {e}"))?;
    let pack = parse_rule_pack(&text)?;
    if mode == ImportMode::Replace {
        delete_imported_rules(db).await?;
    }
    apply_pack(db, pack).await
}

async fn apply_pack(db: &DbInstances, pack: RulePack) -> Result<ImportSummary, String> {
    let mut summary = ImportSummary::default();
    for entry in pack.rules {
        match prepare_entry(entry) {
            PreparedEntry::Skip(skipped) => summary.skipped.push(skipped),
            PreparedEntry::Ready { import_id, input } => {
                if let Some(ref key) = import_id {
                    if let Some(existing) = get_rule_by_import_id(db, key).await? {
                        update_rule(db, existing.id, input).await?;
                        summary.updated += 1;
                        continue;
                    }
                }
                create_imported_rule(db, input, import_id).await?;
                summary.created += 1;
            }
        }
    }
    Ok(summary)
}

/// Same as `apply_pack` but against a pool — used by unit tests.
#[cfg(test)]
pub(crate) async fn apply_pack_pool(
    pool: &SqlitePool,
    pack: RulePack,
    mode: ImportMode,
) -> Result<ImportSummary, String> {
    if mode == ImportMode::Replace {
        crate::custom_rule::delete_imported_rules_pool(pool).await?;
    }
    let mut summary = ImportSummary::default();
    for entry in pack.rules {
        match prepare_entry(entry) {
            PreparedEntry::Skip(skipped) => summary.skipped.push(skipped),
            PreparedEntry::Ready { import_id, input } => {
                if let Some(ref key) = import_id {
                    if let Some(existing) =
                        crate::custom_rule::get_rule_by_import_id_pool(pool, key).await?
                    {
                        crate::custom_rule::update_rule_pool(pool, &existing.id, input).await?;
                        summary.updated += 1;
                        continue;
                    }
                }
                crate::custom_rule::insert_rule(pool, input, import_id).await?;
                summary.created += 1;
            }
        }
    }
    Ok(summary)
}

fn deserialize_scalar_string<'de, D>(deserializer: D) -> Result<String, D::Error>
where
    D: Deserializer<'de>,
{
    struct ScalarString;

    impl<'de> de::Visitor<'de> for ScalarString {
        type Value = String;

        fn expecting(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
            f.write_str("a string, bool, or number")
        }

        fn visit_str<E: de::Error>(self, v: &str) -> Result<String, E> {
            Ok(v.to_string())
        }

        fn visit_string<E: de::Error>(self, v: String) -> Result<String, E> {
            Ok(v)
        }

        fn visit_bool<E: de::Error>(self, v: bool) -> Result<String, E> {
            Ok(v.to_string())
        }

        fn visit_i64<E: de::Error>(self, v: i64) -> Result<String, E> {
            Ok(v.to_string())
        }

        fn visit_u64<E: de::Error>(self, v: u64) -> Result<String, E> {
            Ok(v.to_string())
        }

        fn visit_f64<E: de::Error>(self, v: f64) -> Result<String, E> {
            Ok(v.to_string())
        }

        fn visit_unit<E: de::Error>(self) -> Result<String, E> {
            Ok(String::new())
        }

        fn visit_none<E: de::Error>(self) -> Result<String, E> {
            Ok(String::new())
        }
    }

    deserializer.deserialize_any(ScalarString)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::custom_rule::{insert_rule, list_rules_pool};
    use sqlx::SqlitePool;

    const SAMPLE_PACK: &str = r#"
version: 1
rules:
  - id: POD010
    title: CPU limit not set
    description: A container without a CPU limit can consume unbounded CPU.
    severity: medium
    resourceType: Pod
    matcher: declarative
    fieldPath: spec.containers[*].resources.limits.cpu
    operator: absent
    expectedValue: ""
  - id: POD001
    title: Privileged container
    severity: critical
    resourceType: Pod
    matcher: native
    notes: implemented in rules.rs
  - id: POD002
    title: Host network
    severity: high
    resourceType: Pod
    fieldPath: spec.hostNetwork
    operator: equals
    expectedValue: true
  - title: No id always inserts
    severity: low
    resourceType: Pod
    fieldPath: spec.hostIPC
    operator: equals
    expectedValue: "true"
"#;

    fn sample_pack() -> RulePack {
        parse_rule_pack(SAMPLE_PACK).unwrap()
    }

    #[test]
    fn rejects_unsupported_version() {
        let err = parse_rule_pack("version: 2\nrules: []").unwrap_err();
        assert!(err.contains("unsupported rule pack version: 2"));
    }

    #[test]
    fn prepare_skips_native_and_incomplete() {
        let pack = sample_pack();
        let prepared: Vec<_> = pack.rules.into_iter().map(prepare_entry).collect();
        let skips: Vec<_> = prepared
            .iter()
            .filter_map(|p| match p {
                PreparedEntry::Skip(s) => Some(s.reason.as_str()),
                _ => None,
            })
            .collect();
        assert_eq!(skips.len(), 1);
        assert!(skips[0].contains("native matcher"));

        let ready = prepared
            .iter()
            .filter(|p| matches!(p, PreparedEntry::Ready { .. }))
            .count();
        assert_eq!(ready, 3);
    }

    #[test]
    fn expected_value_accepts_bool() {
        let pack = sample_pack();
        let pod002 = pack.rules.iter().find(|r| r.id.as_deref() == Some("POD002")).unwrap();
        assert_eq!(pod002.expected_value, "true");
    }

    #[test]
    fn export_round_trip_preserves_import_id() {
        let rule = Rule {
            id: "uuid-1".into(),
            rule_id: "POD10".into(),
            title: "CPU limit not set".into(),
            description: String::new(),
            severity: Severity::Medium,
            resource_type: "Pod".into(),
            field_path: "spec.containers[*].resources.limits.cpu".into(),
            operator: Operator::Absent,
            expected_value: String::new(),
            import_id: Some("POD010".into()),
            created_at: String::new(),
            updated_at: String::new(),
        };
        let yaml = serialize_rule_pack(&[rule]).unwrap();
        let pack = parse_rule_pack(&yaml).unwrap();
        assert_eq!(pack.version, 1);
        assert_eq!(pack.rules[0].id.as_deref(), Some("POD010"));
        assert_eq!(pack.rules[0].operator, Some(Operator::Absent));
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
        sqlx::query(
            "CREATE UNIQUE INDEX idx_custom_rules_import_id \
             ON custom_rules(import_id) WHERE import_id IS NOT NULL",
        )
        .execute(&pool)
        .await
        .unwrap();
        sqlx::query(
            "CREATE UNIQUE INDEX idx_custom_rules_rule_id \
             ON custom_rules(rule_id) WHERE rule_id IS NOT NULL AND rule_id != ''",
        )
        .execute(&pool)
        .await
        .unwrap();
        pool
    }

    #[tokio::test]
    async fn merge_reimport_updates_instead_of_duplicating() {
        let pool = setup_pool().await;
        let pack = sample_pack();
        let first = apply_pack_pool(&pool, pack, ImportMode::Merge)
            .await
            .unwrap();
        assert_eq!(first.created, 3);
        assert_eq!(first.updated, 0);
        assert_eq!(first.skipped.len(), 1);

        let second = apply_pack_pool(&pool, sample_pack(), ImportMode::Merge)
            .await
            .unwrap();
        assert_eq!(second.created, 1); // the no-id entry always inserts
        assert_eq!(second.updated, 2);
        assert_eq!(second.skipped.len(), 1);

        let rows = list_rules_pool(&pool).await.unwrap();
        let imported: Vec<_> = rows.iter().filter(|r| r.import_id.is_some()).collect();
        assert_eq!(imported.len(), 2);
    }

    #[tokio::test]
    async fn replace_leaves_hand_authored_rows() {
        let pool = setup_pool().await;
        insert_rule(
            &pool,
            RuleInput {
                title: "hand".into(),
                description: String::new(),
                severity: Severity::Low,
                resource_type: "Pod".into(),
                field_path: "spec.hostNetwork".into(),
                operator: Operator::Equals,
                expected_value: "true".into(),
            },
            None,
        )
        .await
        .unwrap();

        apply_pack_pool(&pool, sample_pack(), ImportMode::Merge)
            .await
            .unwrap();
        apply_pack_pool(&pool, sample_pack(), ImportMode::Replace)
            .await
            .unwrap();

        let rows = list_rules_pool(&pool).await.unwrap();
        assert!(rows.iter().any(|r| r.title == "hand" && r.import_id.is_none()));
        assert!(rows.iter().any(|r| r.import_id.as_deref() == Some("POD010")));
    }
}
