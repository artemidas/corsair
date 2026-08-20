//! Security rules run against a connected cluster.
//!
//! Two kinds of rules are evaluated together during a scan:
//! - **Hardcoded** rules (RBAC001/002, POD001/004) — fixed set in this file.
//! - **Stored rules** — matchers persisted in SQLite. Each is a simple
//!   matcher: given a resource type, a field path, and an expected value,
//!   flag any resource that has a value equal to the expected one at that
//!   path. See `custom_rule::evaluate_field_path` and `value_matches`.
//!
//! Hardcoded rules read from a pre-fetched `ClusterData` snapshot so they
//! all see the same view. Stored rules use the cluster client directly to
//! fetch only the resource type they need.

use k8s_openapi::api::core::v1::{Pod, ServiceAccount};
use k8s_openapi::api::rbac::v1::{
    ClusterRole, ClusterRoleBinding, Role, RoleBinding,
};
use kube::api::ListParams;
use kube::{Api, Client};
use serde::{Deserialize, Serialize};
use serde_json::Value;

use crate::custom_rule::{evaluate_field_path, evaluate_operator, Rule as StoredRule};

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Critical,
    High,
    Medium,
    Low,
}

impl Severity {
    pub fn as_str(self) -> &'static str {
        match self {
            Self::Critical => "critical",
            Self::High => "high",
            Self::Medium => "medium",
            Self::Low => "low",
        }
    }

    pub fn parse(value: &str) -> Result<Self, String> {
        match value {
            "critical" => Ok(Self::Critical),
            "high" => Ok(Self::High),
            "medium" => Ok(Self::Medium),
            "low" => Ok(Self::Low),
            other => Err(format!("unknown severity '{other}' in db")),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Finding {
    pub id: String,
    pub rule_id: String,
    pub rule_title: String,
    pub severity: Severity,
    pub resource_kind: String,
    pub resource_name: String,
    pub namespace: Option<String>,
    pub message: String,
}

/// Snapshot of the cluster resources the hardcoded rules need. Fetched
/// once per scan so every rule reads the same consistent view.
pub struct ClusterData {
    pub cluster_role_bindings: Vec<ClusterRoleBinding>,
    pub roles: Vec<Role>,
    pub cluster_roles: Vec<ClusterRole>,
    pub pods: Vec<Pod>,
}

pub async fn fetch_cluster_data(client: &Client) -> Result<ClusterData, kube::Error> {
    let crb_api: Api<ClusterRoleBinding> = Api::all(client.clone());
    let role_api: Api<Role> = Api::all(client.clone());
    let cluster_role_api: Api<ClusterRole> = Api::all(client.clone());
    let pod_api: Api<Pod> = Api::all(client.clone());

    let cluster_role_bindings = crb_api.list(&ListParams::default()).await?.items;
    let roles = role_api.list(&ListParams::default()).await?.items;
    let cluster_roles = cluster_role_api.list(&ListParams::default()).await?.items;
    let pods = pod_api.list(&ListParams::default()).await?.items;

    Ok(ClusterData {
        cluster_role_bindings,
        roles,
        cluster_roles,
        pods,
    })
}

pub trait Rule {
    fn id(&self) -> &'static str;
    fn title(&self) -> &'static str;
    fn check(&self, data: &ClusterData) -> Vec<Finding>;
}

/// RBAC001: the "default" ServiceAccount in any namespace is bound to
/// cluster-admin via a ClusterRoleBinding.
pub struct Rbac001;

impl Rule for Rbac001 {
    fn id(&self) -> &'static str {
        "RBAC001"
    }

    fn title(&self) -> &'static str {
        "Default ServiceAccount bound to cluster-admin"
    }

    fn check(&self, data: &ClusterData) -> Vec<Finding> {
        let mut findings = Vec::new();
        for crb in &data.cluster_role_bindings {
            if crb.role_ref.name != "cluster-admin" {
                continue;
            }
            let Some(subjects) = &crb.subjects else {
                continue;
            };
            for subject in subjects {
                if subject.kind == "ServiceAccount" && subject.name == "default" {
                    let ns = subject.namespace.clone().unwrap_or_else(|| "default".into());
                    let crb_name = crb.metadata.name.clone().unwrap_or_default();
                    findings.push(Finding {
                        id: format!("RBAC001-{ns}-{crb_name}"),
                        rule_id: self.id().to_string(),
                        rule_title: self.title().to_string(),
                        severity: Severity::Critical,
                        resource_kind: "ServiceAccount".into(),
                        resource_name: "default".into(),
                        namespace: Some(ns.clone()),
                        message: format!(
                            "ServiceAccount 'default' in namespace '{ns}' is bound to cluster-admin via ClusterRoleBinding '{crb_name}'"
                        ),
                    });
                }
            }
        }
        findings
    }
}

/// RBAC002: a Role or ClusterRole grants the wildcard verb "*".
pub struct Rbac002;

impl Rule for Rbac002 {
    fn id(&self) -> &'static str {
        "RBAC002"
    }

    fn title(&self) -> &'static str {
        "Role grants wildcard verb"
    }

    fn check(&self, data: &ClusterData) -> Vec<Finding> {
        let mut findings = Vec::new();

        for role in &data.roles {
            let name = role.metadata.name.clone().unwrap_or_default();
            let ns = role.metadata.namespace.clone();
            let has_wildcard = role
                .rules
                .as_ref()
                .map(|rules| rules.iter().any(|r| r.verbs.iter().any(|v| v == "*")))
                .unwrap_or(false);
            if has_wildcard {
                findings.push(Finding {
                    id: format!("RBAC002-Role-{}-{name}", ns.clone().unwrap_or_default()),
                    rule_id: self.id().to_string(),
                    rule_title: self.title().to_string(),
                    severity: Severity::High,
                    resource_kind: "Role".into(),
                    resource_name: name.clone(),
                    namespace: ns,
                    message: format!("Role '{name}' grants wildcard verb '*'"),
                });
            }
        }

        for cluster_role in &data.cluster_roles {
            let name = cluster_role.metadata.name.clone().unwrap_or_default();
            let has_wildcard = cluster_role
                .rules
                .as_ref()
                .map(|rules| rules.iter().any(|r| r.verbs.iter().any(|v| v == "*")))
                .unwrap_or(false);
            if has_wildcard {
                findings.push(Finding {
                    id: format!("RBAC002-ClusterRole-{name}"),
                    rule_id: self.id().to_string(),
                    rule_title: self.title().to_string(),
                    severity: Severity::High,
                    resource_kind: "ClusterRole".into(),
                    resource_name: name.clone(),
                    namespace: None,
                    message: format!("ClusterRole '{name}' grants wildcard verb '*'"),
                });
            }
        }

        findings
    }
}

/// POD001: a Pod has a container running with securityContext.privileged = true.
pub struct Pod001;

impl Rule for Pod001 {
    fn id(&self) -> &'static str {
        "POD001"
    }

    fn title(&self) -> &'static str {
        "Privileged container"
    }

    fn check(&self, data: &ClusterData) -> Vec<Finding> {
        let mut findings = Vec::new();
        for pod in &data.pods {
            let name = pod.metadata.name.clone().unwrap_or_default();
            let ns = pod.metadata.namespace.clone();
            let Some(spec) = &pod.spec else { continue };
            for c in &spec.containers {
                let privileged = c
                    .security_context
                    .as_ref()
                    .and_then(|sc| sc.privileged)
                    .unwrap_or(false);
                if privileged {
                    findings.push(Finding {
                        id: format!(
                            "POD001-{}-{name}-{}",
                            ns.clone().unwrap_or_default(),
                            c.name
                        ),
                        rule_id: self.id().to_string(),
                        rule_title: self.title().to_string(),
                        severity: Severity::Critical,
                        resource_kind: "Pod".into(),
                        resource_name: name.clone(),
                        namespace: ns.clone(),
                        message: format!(
                            "Container '{}' in Pod '{name}' runs with securityContext.privileged=true",
                            c.name
                        ),
                    });
                }
            }
        }
        findings
    }
}

/// POD004: a container does not set securityContext.runAsNonRoot = true
/// (checking the container's own setting, falling back to the Pod-level
/// setting when the container doesn't override it).
pub struct Pod004;

impl Rule for Pod004 {
    fn id(&self) -> &'static str {
        "POD004"
    }

    fn title(&self) -> &'static str {
        "Container not running as non-root"
    }

    fn check(&self, data: &ClusterData) -> Vec<Finding> {
        let mut findings = Vec::new();
        for pod in &data.pods {
            let name = pod.metadata.name.clone().unwrap_or_default();
            let ns = pod.metadata.namespace.clone();
            let Some(spec) = &pod.spec else { continue };
            let pod_level = spec
                .security_context
                .as_ref()
                .and_then(|sc| sc.run_as_non_root);
            for c in &spec.containers {
                let container_level = c
                    .security_context
                    .as_ref()
                    .and_then(|sc| sc.run_as_non_root);
                let effective = container_level.or(pod_level);
                if effective != Some(true) {
                    findings.push(Finding {
                        id: format!(
                            "POD004-{}-{name}-{}",
                            ns.clone().unwrap_or_default(),
                            c.name
                        ),
                        rule_id: self.id().to_string(),
                        rule_title: self.title().to_string(),
                        severity: Severity::Medium,
                        resource_kind: "Pod".into(),
                        resource_name: name.clone(),
                        namespace: ns.clone(),
                        message: format!(
                            "Container '{}' in Pod '{name}' does not set securityContext.runAsNonRoot=true",
                            c.name
                        ),
                    });
                }
            }
        }
        findings
    }
}

fn all_rules() -> Vec<Box<dyn Rule>> {
    vec![
        Box::new(Rbac001),
        Box::new(Rbac002),
        Box::new(Pod001),
        Box::new(Pod004),
    ]
}

/// Run every hardcoded rule against the given snapshot and flatten the results.
pub fn run_rules(data: &ClusterData) -> Vec<Finding> {
    all_rules().iter().flat_map(|rule| rule.check(data)).collect()
}

/// Evaluate a single stored rule against the cluster. Fetches the
/// resources of the rule's `resource_type`, walks the `field_path`, and
/// emits one finding per resource where the rule's operator holds over
/// the resolved leaves.
pub async fn evaluate_rule(
    client: &Client,
    rule: &StoredRule,
) -> Result<Vec<Finding>, kube::Error> {
    let mut findings = Vec::new();
    let items = fetch_resources_as_json(client, &rule.resource_type).await?;
    let kind = rule.resource_type.clone();

    for item in &items {
        let values = evaluate_field_path(item, &rule.field_path);
        if evaluate_operator(&values, rule.operator, &rule.expected_value) {
            let name = item
                .get("metadata")
                .and_then(|m| m.get("name"))
                .and_then(|n| n.as_str())
                .unwrap_or("")
                .to_string();
            let ns = item
                .get("metadata")
                .and_then(|m| m.get("namespace"))
                .and_then(|n| n.as_str())
                .map(String::from);
            let id = format!(
                "{}-{}-{}",
                rule.id,
                ns.clone().unwrap_or_default(),
                name
            );
            let msg = if rule.description.is_empty() {
                rule.title.clone()
            } else {
                format!("{}: {}", rule.title, rule.description)
            };
            findings.push(Finding {
                id,
                rule_id: rule.rule_id.clone(),
                rule_title: rule.title.clone(),
                severity: rule.severity,
                resource_kind: kind.clone(),
                resource_name: name,
                namespace: ns,
                message: msg,
            });
        }
    }

    Ok(findings)
}

/// Run every stored rule, collecting findings. Errors evaluating an
/// individual rule are logged and skipped (the scan keeps going).
pub async fn evaluate_rules(client: &Client, rules: &[StoredRule]) -> Vec<Finding> {
    let mut out = Vec::new();
    for rule in rules {
        match evaluate_rule(client, rule).await {
            Ok(f) => out.extend(f),
            Err(e) => eprintln!("error evaluating rule {}: {e}", rule.id),
        }
    }
    out
}

/// Fetch every resource of the given type as a `serde_json::Value`.
/// Returns an empty vec for unknown resource types.
async fn fetch_resources_as_json(
    client: &Client,
    resource_type: &str,
) -> Result<Vec<Value>, kube::Error> {
    match resource_type {
        "Pod" => {
            let api: Api<Pod> = Api::all(client.clone());
            let pods = api.list(&ListParams::default()).await?;
            Ok(pods
                .items
                .iter()
                .filter_map(|p| serde_json::to_value(p).ok())
                .collect())
        }
        "ServiceAccount" => {
            let api: Api<ServiceAccount> = Api::all(client.clone());
            let sas = api.list(&ListParams::default()).await?;
            Ok(sas
                .items
                .iter()
                .filter_map(|s| serde_json::to_value(s).ok())
                .collect())
        }
        "Role" => {
            let api: Api<Role> = Api::all(client.clone());
            let rs = api.list(&ListParams::default()).await?;
            Ok(rs.items
                .iter()
                .filter_map(|r| serde_json::to_value(r).ok())
                .collect())
        }
        "ClusterRole" => {
            let api: Api<ClusterRole> = Api::all(client.clone());
            let rs = api.list(&ListParams::default()).await?;
            Ok(rs.items
                .iter()
                .filter_map(|r| serde_json::to_value(r).ok())
                .collect())
        }
        "RoleBinding" => {
            let api: Api<RoleBinding> = Api::all(client.clone());
            let rs = api.list(&ListParams::default()).await?;
            Ok(rs.items
                .iter()
                .filter_map(|r| serde_json::to_value(r).ok())
                .collect())
        }
        "ClusterRoleBinding" => {
            let api: Api<ClusterRoleBinding> = Api::all(client.clone());
            let rs = api.list(&ListParams::default()).await?;
            Ok(rs.items
                .iter()
                .filter_map(|r| serde_json::to_value(r).ok())
                .collect())
        }
        _ => Ok(vec![]),
    }
}
