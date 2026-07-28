//! Fixed set of security rules run against a connected cluster.
//!
//! MVP scope: four hardcoded rules, no custom rule management, no
//! persistence. Resources are fetched once into `ClusterData` and every
//! rule inspects that snapshot.

use k8s_openapi::api::core::v1::Pod;
use k8s_openapi::api::rbac::v1::{ClusterRole, ClusterRoleBinding, Role};
use kube::api::ListParams;
use kube::{Api, Client};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Critical,
    High,
    Medium,
    Low,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Finding {
    pub id: String,
    pub rule_id: String,
    pub severity: Severity,
    pub resource_kind: String,
    pub resource_name: String,
    pub namespace: Option<String>,
    pub message: String,
}

/// Snapshot of the cluster resources the rules below need. Fetched once per
/// scan so every rule reads the same consistent view.
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
    fn check(&self, data: &ClusterData) -> Vec<Finding>;
}

/// RBAC001: the "default" ServiceAccount in any namespace is bound to
/// cluster-admin via a ClusterRoleBinding.
pub struct Rbac001;

impl Rule for Rbac001 {
    fn id(&self) -> &'static str {
        "RBAC001"
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

/// Run every fixed rule against the given snapshot and flatten the results.
pub fn run_rules(data: &ClusterData) -> Vec<Finding> {
    all_rules().iter().flat_map(|rule| rule.check(data)).collect()
}
