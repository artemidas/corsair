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

const schemaScans = `CREATE TABLE IF NOT EXISTS scans (
	id TEXT PRIMARY KEY,
	project_id TEXT NOT NULL,
	status TEXT NOT NULL,
	context TEXT,
	error TEXT,
	finding_count INTEGER NOT NULL DEFAULT 0,
	started_at TEXT NOT NULL,
	finished_at TEXT NOT NULL
)`

const schemaFindings = `CREATE TABLE IF NOT EXISTS findings (
	id TEXT PRIMARY KEY,
	scan_id TEXT NOT NULL,
	rule_id TEXT NOT NULL,
	rule_title TEXT NOT NULL,
	severity TEXT NOT NULL,
	resource_kind TEXT NOT NULL,
	resource_name TEXT NOT NULL,
	namespace TEXT,
	message TEXT NOT NULL
)`

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
		version = 2
	}
	if version < 3 {
		if _, err := db.Exec(schemaScans); err != nil {
			return fmt.Errorf("create scans table: %w", err)
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_scans_project_started
			ON scans(project_id, started_at)`); err != nil {
			return fmt.Errorf("index scans: %w", err)
		}
		if _, err := db.Exec(schemaFindings); err != nil {
			return fmt.Errorf("create findings table: %w", err)
		}
		if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_findings_scan_rule
			ON findings(scan_id, rule_id)`); err != nil {
			return fmt.Errorf("index findings: %w", err)
		}
		if _, err := db.Exec("PRAGMA user_version = 3"); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
		version = 3
	}
	if version < 4 {
		if err := migrateRulesToRego(db); err != nil {
			return err
		}
		if _, err := db.Exec("PRAGMA user_version = 4"); err != nil {
			return fmt.Errorf("set schema version: %w", err)
		}
	}
	return nil
}

func migrateRulesToRego(db *sql.DB) error {
	if _, err := db.Exec(`ALTER TABLE rules ADD COLUMN rego TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("add rules.rego: %w", err)
	}
	rows, err := db.Query(`SELECT id, field_path, operator, expected_value FROM rules`)
	if err != nil {
		return fmt.Errorf("select rules for rego migration: %w", err)
	}
	defer rows.Close()

	type legacy struct {
		id, fieldPath, operator, expected string
	}
	var todo []legacy
	for rows.Next() {
		var row legacy
		if err := rows.Scan(&row.id, &row.fieldPath, &row.operator, &row.expected); err != nil {
			return err
		}
		todo = append(todo, row)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, row := range todo {
		src, err := convertDeclarativeToRego(row.fieldPath, row.operator, row.expected)
		if err != nil {
			return fmt.Errorf("convert rule %s: %w", row.id, err)
		}
		if _, err := db.Exec(`UPDATE rules SET rego = ? WHERE id = ?`, src, row.id); err != nil {
			return fmt.Errorf("write rego for rule %s: %w", row.id, err)
		}
	}

	if _, err := db.Exec(`CREATE TABLE rules_v4 (
		id TEXT PRIMARY KEY,
		rule_id TEXT,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		severity TEXT NOT NULL,
		resource_type TEXT NOT NULL,
		rego TEXT NOT NULL,
		import_id TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`); err != nil {
		return fmt.Errorf("create rules_v4: %w", err)
	}
	if _, err := db.Exec(`INSERT INTO rules_v4
		(id, rule_id, title, description, severity, resource_type, rego, import_id, created_at, updated_at)
		SELECT id, rule_id, title, description, severity, resource_type, rego, import_id, created_at, updated_at
		FROM rules`); err != nil {
		return fmt.Errorf("copy rules: %w", err)
	}
	if _, err := db.Exec(`DROP TABLE rules`); err != nil {
		return fmt.Errorf("drop legacy rules: %w", err)
	}
	if _, err := db.Exec(`ALTER TABLE rules_v4 RENAME TO rules`); err != nil {
		return fmt.Errorf("rename rules_v4: %w", err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_import_id
		ON rules(import_id) WHERE import_id IS NOT NULL`); err != nil {
		return fmt.Errorf("index rules.import_id: %w", err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_rules_rule_id
		ON rules(rule_id) WHERE rule_id IS NOT NULL AND rule_id != ''`); err != nil {
		return fmt.Errorf("index rules.rule_id: %w", err)
	}
	return nil
}
