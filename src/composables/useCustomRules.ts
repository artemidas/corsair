import { computed, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";

export type RuleSeverity = "critical" | "high" | "medium" | "low";
export type Operator =
  | "equals"
  | "notEquals"
  | "present"
  | "absent"
  | "arrayExcludes";
export type ExportScope = "user" | "all";
export type ImportMode = "merge" | "replace";

export type BuiltInKind = "builtin";
export type UserKind = "user";
export type RuleSource = BuiltInKind | UserKind;

export interface CustomRule {
  id: string;
  title: string;
  description: string;
  severity: RuleSeverity;
  resourceType: string;
  fieldPath: string;
  operator: Operator;
  expectedValue: string;
  importId: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface CustomRuleInput {
  title: string;
  description: string;
  severity: RuleSeverity;
  resourceType: string;
  fieldPath: string;
  operator: Operator;
  expectedValue: string;
}

export interface SkippedRule {
  id: string;
  title: string;
  reason: string;
}

export interface ImportSummary {
  created: number;
  updated: number;
  skipped: SkippedRule[];
}

export const RESOURCE_TYPES = [
  "Pod",
  "ServiceAccount",
  "Role",
  "ClusterRole",
  "RoleBinding",
  "ClusterRoleBinding",
] as const;

export const OPERATORS: { value: Operator; label: string }[] = [
  { value: "equals", label: "equals" },
  { value: "notEquals", label: "not equals" },
  { value: "present", label: "is present" },
  { value: "absent", label: "is absent" },
  { value: "arrayExcludes", label: "array excludes" },
];

export const NEEDS_EXPECTED_VALUE: Operator[] = [
  "equals",
  "notEquals",
  "arrayExcludes",
];

export const OPERATOR_LABEL: Record<Operator, string> = {
  equals: "==",
  notEquals: "!=",
  present: "is set",
  absent: "is not set",
  arrayExcludes: "excludes",
};

export function describeCheck(r: CustomRule): string {
  const needsValue = NEEDS_EXPECTED_VALUE.includes(r.operator);
  return `${r.resourceType} · ${r.fieldPath} ${OPERATOR_LABEL[r.operator]}${needsValue ? " " + r.expectedValue : ""}`;
}

export const YAML_FILTERS = [{ name: "YAML", extensions: ["yaml", "yml"] }];

const BUILTIN_PREFIX = "BUILTIN-";

const rules = ref<CustomRule[]>([]);
const loading = ref(false);
const loadError = ref("");

export function isBuiltIn(rule: CustomRule): boolean {
  return rule.id.startsWith(BUILTIN_PREFIX);
}

export function useCustomRules() {
  const userRules = computed(() => rules.value.filter((r) => !isBuiltIn(r)));
  const builtInRules = computed(() => rules.value.filter(isBuiltIn));

  async function loadRules() {
    loading.value = true;
    loadError.value = "";
    try {
      rules.value = await invoke<CustomRule[]>("list_custom_rules");
    } catch (err) {
      loadError.value = String(err);
    } finally {
      loading.value = false;
    }
  }

  async function createRule(input: CustomRuleInput): Promise<CustomRule> {
    const created = await invoke<CustomRule>("create_custom_rule", { input });
    rules.value = [...userRules.value, created, ...builtInRules.value];
    return created;
  }

  async function updateRule(id: string, input: CustomRuleInput): Promise<CustomRule> {
    const updated = await invoke<CustomRule>("update_custom_rule", { id, input });
    rules.value = rules.value.map((r) => (r.id === id ? updated : r));
    return updated;
  }

  async function deleteRule(id: string) {
    await invoke("delete_custom_rule", { id });
    rules.value = rules.value.filter((r) => r.id !== id);
  }

  async function exportRules(
    path: string,
    scope: ExportScope = "user",
  ): Promise<number> {
    return invoke<number>("export_rules", { path, scope });
  }

  async function importRules(
    path: string,
    mode: ImportMode = "merge",
  ): Promise<ImportSummary> {
    const summary = await invoke<ImportSummary>("import_rules", { path, mode });
    await loadRules();
    return summary;
  }

  function getRuleById(id: string): CustomRule | null {
    return rules.value.find((r) => r.id === id) ?? null;
  }

  return {
    rules,
    userRules,
    builtInRules,
    loading,
    loadError,
    loadRules,
    createRule,
    updateRule,
    deleteRule,
    importRules,
    exportRules,
    getRuleById,
  };
}
