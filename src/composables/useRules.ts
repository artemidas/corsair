import { ref } from "vue";
import {
  CreateRule as createRuleBound,
  DeleteRule as deleteRuleBound,
  ExportRules as exportRulesBound,
  ImportRules as importRulesBound,
  ListRules,
  UpdateRule as updateRuleBound,
} from "@/bindings/ladon/backend/rule/service";
import type {
  ImportMode as BoundImportMode,
  ImportSummary as BoundImportSummary,
  Operator as BoundOperator,
  Rule as BoundRule,
  RuleInput as BoundRuleInput,
  Severity as BoundSeverity,
} from "@/bindings/ladon/backend/rule/models";

export type RuleSeverity = "critical" | "high" | "medium" | "low";
export type Operator =
  | "equals"
  | "notEquals"
  | "present"
  | "absent"
  | "arrayExcludes";
export type ImportMode = "merge" | "replace";

export interface Rule {
  id: string;
  ruleId: string;
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

export interface RuleInput {
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

export function describeCheck(r: Rule): string {
  const needsValue = NEEDS_EXPECTED_VALUE.includes(r.operator);
  return `${r.resourceType} · ${r.fieldPath} ${OPERATOR_LABEL[r.operator]}${needsValue ? " " + r.expectedValue : ""}`;
}

function toBoundInput(input: RuleInput): BoundRuleInput {
  return {
    title: input.title,
    description: input.description,
    severity: input.severity as BoundSeverity,
    resourceType: input.resourceType,
    fieldPath: input.fieldPath,
    operator: input.operator as BoundOperator,
    expectedValue: input.expectedValue,
  };
}

function fromBound(rule: BoundRule): Rule {
  return {
    id: rule.id,
    ruleId: rule.ruleId,
    title: rule.title,
    description: rule.description,
    severity: rule.severity as RuleSeverity,
    resourceType: rule.resourceType,
    fieldPath: rule.fieldPath,
    operator: rule.operator as Operator,
    expectedValue: rule.expectedValue,
    importId: rule.importId,
    createdAt: rule.createdAt,
    updatedAt: rule.updatedAt,
  };
}

function fromBoundSummary(summary: BoundImportSummary): ImportSummary {
  return {
    created: summary.created,
    updated: summary.updated,
    skipped: (summary.skipped ?? []).map((s) => ({
      id: s.id,
      title: s.title,
      reason: s.reason,
    })),
  };
}

const rules = ref<Rule[]>([]);
const loading = ref(false);
const loadError = ref("");

export function useRules() {
  async function loadRules() {
    loading.value = true;
    loadError.value = "";
    try {
      rules.value = ((await ListRules()) ?? []).map(fromBound);
    } catch (err) {
      loadError.value = String(err);
    } finally {
      loading.value = false;
    }
  }

  async function createRule(input: RuleInput): Promise<Rule> {
    const created = fromBound(await createRuleBound(toBoundInput(input)));
    rules.value = [...rules.value, created];
    return created;
  }

  async function updateRule(id: string, input: RuleInput): Promise<Rule> {
    const updated = fromBound(await updateRuleBound(id, toBoundInput(input)));
    rules.value = rules.value.map((r) => (r.id === id ? updated : r));
    return updated;
  }

  async function deleteRule(id: string) {
    await deleteRuleBound(id);
    rules.value = rules.value.filter((r) => r.id !== id);
  }

  async function exportRules(path: string): Promise<number> {
    return exportRulesBound(path);
  }

  async function importRules(
    path: string,
    mode: ImportMode = "merge",
  ): Promise<ImportSummary> {
    const summary = await importRulesBound(path, mode as BoundImportMode);
    await loadRules();
    return fromBoundSummary(summary);
  }

  function getRuleById(id: string): Rule | null {
    return rules.value.find((r) => r.id === id) ?? null;
  }

  return {
    rules,
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
