package appdb

import (
	"database/sql"
	"fmt"
)

const schemaProjects = `CREATE TABLE IF NOT EXISTS projects (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	config TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
)`

const schemaRules = `CREATE TABLE IF NOT EXISTS rules (
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
)`

const seedRules = `INSERT OR IGNORE INTO rules
	(id, title, description, severity, resource_type, field_path, operator, expected_value, import_id, created_at, updated_at)
	VALUES
	('BUILTIN-001', 'Privileged container', 'Containers running with securityContext.privileged=true can access all host devices and capabilities.', 'critical', 'Pod', 'spec.containers[*].securityContext.privileged', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-002', 'Host network', 'Pods with spec.hostNetwork=true share the host''s network namespace and can listen on any interface.', 'high', 'Pod', 'spec.hostNetwork', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-003', 'Host PID namespace', 'Pods with spec.hostPID=true can see and signal all host processes.', 'high', 'Pod', 'spec.hostPID', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-004', 'Host IPC namespace', 'Pods with spec.hostIPC=true share the host''s IPC namespace.', 'medium', 'Pod', 'spec.hostIPC', 'equals', 'true', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-005', 'Default ServiceAccount in use', 'Pods or workload bindings that rely on the ''default'' ServiceAccount inherit its permissive token by default.', 'medium', 'ServiceAccount', 'metadata.name', 'equals', 'default', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-006', 'Role grants wildcard verb', 'A Role granting the ''*'' verb allows every action on the listed resources.', 'high', 'Role', 'rules[*].verbs[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-007', 'ClusterRole grants wildcard verb', 'A ClusterRole granting the ''*'' verb allows every action cluster-wide.', 'high', 'ClusterRole', 'rules[*].verbs[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-008', 'Role grants wildcard API group', 'A Role granting the ''*'' apiGroup effectively grants every API.', 'high', 'Role', 'rules[*].apiGroups[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00'),
	('BUILTIN-009', 'ClusterRole grants wildcard API group', 'A ClusterRole granting the ''*'' apiGroup effectively grants every API cluster-wide.', 'high', 'ClusterRole', 'rules[*].apiGroups[*]', 'equals', '*', NULL, '2024-01-01T00:00:00+00:00', '2024-01-01T00:00:00+00:00')`

func migrate(db *sql.DB) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version < 1 {
		if _, err := db.Exec(schemaProjects); err != nil {
			return fmt.Errorf("create projects table: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 1"); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
		version = 1
	}
	if version < 2 {
		if _, err := db.Exec(schemaRules); err != nil {
			return fmt.Errorf("create rules table: %w", err)
		}
		if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_import_id
			ON rules(import_id) WHERE import_id IS NOT NULL`); err != nil {
			return fmt.Errorf("index rules.import_id: %w", err)
		}
		if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_rule_id
			ON rules(rule_id) WHERE rule_id IS NOT NULL AND rule_id != ''`); err != nil {
			return fmt.Errorf("index rules.rule_id: %w", err)
		}
		if _, err := db.Exec(seedRules); err != nil {
			return fmt.Errorf("seed rules: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 2"); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}
	return nil
}
