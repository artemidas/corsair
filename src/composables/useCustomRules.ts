import { computed, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";

export type RuleSeverity = "critical" | "high" | "medium" | "low";

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
  expectedValue: string;
  createdAt: string;
  updatedAt: string;
}

export interface CustomRuleInput {
  title: string;
  description: string;
  severity: RuleSeverity;
  resourceType: string;
  fieldPath: string;
  expectedValue: string;
}

export const RESOURCE_TYPES = [
  "Pod",
  "ServiceAccount",
  "Role",
  "ClusterRole",
  "RoleBinding",
  "ClusterRoleBinding",
] as const;

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
    getRuleById,
  };
}
