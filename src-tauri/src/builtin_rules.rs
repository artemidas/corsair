//! Curated library of CIS-aligned rules baked into the binary.
//!
//! These are read-only from the user's perspective: they show up in the
//! sidebar with a "Built-in" badge and have no Edit/Delete affordance.
//! The engine evaluates them alongside user-authored rules. The `id`
//! is prefixed with `BUILTIN-` so it's never confused with user IDs
//! (which are UUIDs).

use crate::custom_rule::{CustomRule, Operator};
use crate::rules::Severity;

const FIXED_CREATED_AT: &str = "2024-01-01T00:00:00+00:00";

pub fn builtin_rules() -> Vec<CustomRule> {
    vec![
        CustomRule {
            id: "BUILTIN-001".into(),
            title: "Privileged container".into(),
            description: "Containers running with securityContext.privileged=true can access all host devices and capabilities.".into(),
            severity: Severity::Critical,
            resource_type: "Pod".into(),
            field_path: "spec.containers[*].securityContext.privileged".into(),
            operator: Operator::Equals,
            expected_value: "true".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-002".into(),
            title: "Host network".into(),
            description: "Pods with spec.hostNetwork=true share the host's network namespace and can listen on any interface.".into(),
            severity: Severity::High,
            resource_type: "Pod".into(),
            field_path: "spec.hostNetwork".into(),
            operator: Operator::Equals,
            expected_value: "true".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-003".into(),
            title: "Host PID namespace".into(),
            description: "Pods with spec.hostPID=true can see and signal all host processes.".into(),
            severity: Severity::High,
            resource_type: "Pod".into(),
            field_path: "spec.hostPID".into(),
            operator: Operator::Equals,
            expected_value: "true".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-004".into(),
            title: "Host IPC namespace".into(),
            description: "Pods with spec.hostIPC=true share the host's IPC namespace.".into(),
            severity: Severity::Medium,
            resource_type: "Pod".into(),
            field_path: "spec.hostIPC".into(),
            operator: Operator::Equals,
            expected_value: "true".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-005".into(),
            title: "Default ServiceAccount in use".into(),
            description: "Pods or workload bindings that rely on the 'default' ServiceAccount inherit its permissive token by default.".into(),
            severity: Severity::Medium,
            resource_type: "ServiceAccount".into(),
            field_path: "metadata.name".into(),
            operator: Operator::Equals,
            expected_value: "default".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-006".into(),
            title: "Role grants wildcard verb".into(),
            description: "A Role granting the '*' verb allows every action on the listed resources.".into(),
            severity: Severity::High,
            resource_type: "Role".into(),
            field_path: "rules[*].verbs[*]".into(),
            operator: Operator::Equals,
            expected_value: "*".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-007".into(),
            title: "ClusterRole grants wildcard verb".into(),
            description: "A ClusterRole granting the '*' verb allows every action cluster-wide.".into(),
            severity: Severity::High,
            resource_type: "ClusterRole".into(),
            field_path: "rules[*].verbs[*]".into(),
            operator: Operator::Equals,
            expected_value: "*".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-008".into(),
            title: "Role grants wildcard API group".into(),
            description: "A Role granting the '*' apiGroup effectively grants every API.".into(),
            severity: Severity::High,
            resource_type: "Role".into(),
            field_path: "rules[*].apiGroups[*]".into(),
            operator: Operator::Equals,
            expected_value: "*".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
        CustomRule {
            id: "BUILTIN-009".into(),
            title: "ClusterRole grants wildcard API group".into(),
            description: "A ClusterRole granting the '*' apiGroup effectively grants every API cluster-wide.".into(),
            severity: Severity::High,
            resource_type: "ClusterRole".into(),
            field_path: "rules[*].apiGroups[*]".into(),
            operator: Operator::Equals,
            expected_value: "*".into(),
            import_id: None,
            created_at: FIXED_CREATED_AT.into(),
            updated_at: FIXED_CREATED_AT.into(),
        },
    ]
}
