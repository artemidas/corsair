<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { invoke } from "@tauri-apps/api/core";
import { ArrowLeft, Trash2, TriangleAlert } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { useCustomRules, isBuiltIn, type CustomRule } from "@/composables/useCustomRules";
import { useProjects, type Project } from "@/composables/useProjects";

type Severity = "critical" | "high" | "medium" | "low";

interface Finding {
  id: string;
  ruleId: string;
  severity: Severity;
  resourceKind: string;
  resourceName: string;
  namespace: string | null;
  message: string;
}

interface Connection {
  context: string | null;
}

const props = defineProps<{
  rule: CustomRule;
  selectedProject: Project | null;
}>();

const emit = defineEmits<{
  back: [];
  edit: [rule: CustomRule];
}>();

const { deleteRule } = useCustomRules();
const { setConnection } = useProjects();

const builtIn = computed(() => isBuiltIn(props.rule));

const connecting = ref(false);
const scanning = ref(false);
const scanError = ref("");
const findings = ref<Finding[]>([]);
const activeContext = ref<Connection | null>(null);

async function refreshConnection() {
  try {
    activeContext.value = await invoke<Connection | null>("active_context");
  } catch {
    activeContext.value = null;
  }
}

async function onConnect() {
  const ctx = props.selectedProject?.config.context ?? null;
  connecting.value = true;
  try {
    await invoke("connect_cluster", { context: ctx });
    setConnection(ctx);
    activeContext.value = { context: ctx };
  } catch (err) {
    scanError.value = String(err);
  } finally {
    connecting.value = false;
  }
}

async function runScan() {
  scanning.value = true;
  scanError.value = "";
  try {
    findings.value = await invoke<Finding[]>("run_scan");
  } catch (err) {
    scanError.value = String(err);
  } finally {
    scanning.value = false;
  }
}

const isConnected = computed(() => activeContext.value !== null);

const contextMatches = computed(() => {
  if (!activeContext.value) return null;
  const want =
    props.rule.resourceType === "Pod" ||
    props.rule.resourceType === "ServiceAccount" ||
    props.rule.resourceType === "Role" ||
    props.rule.resourceType === "RoleBinding";
  if (!want) return true; // cluster-scoped resources always match
  return activeContext.value.context === (props.selectedProject?.config.context ?? null);
});

const matchingFindings = computed(() =>
  findings.value.filter((f) => f.ruleId === props.rule.id),
);

const severityBadgeClass: Record<string, string> = {
  critical: "bg-red-600 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-400 text-black",
  low: "bg-gray-400 text-black",
};

async function onDelete() {
  if (
    !confirm(
      `Delete the rule "${props.rule.title}"? This will not affect built-in rules.`,
    )
  )
    return;
  try {
    await deleteRule(props.rule.id);
    emit("back");
  } catch (err) {
    alert(String(err));
  }
}

onMounted(() => {
  refreshConnection();
});

watch(
  () => props.rule.id,
  () => {
    findings.value = [];
    scanError.value = "";
  },
);
</script>

<template>
  <div class="flex flex-col gap-4">
    <Button
      type="button"
      variant="ghost"
      size="sm"
      class="self-start -ml-3"
      @click="emit('back')"
    >
      <ArrowLeft class="size-4" />
      Back to rules
    </Button>

    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="flex items-start justify-between gap-4">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2 text-xs uppercase tracking-wide text-base-content/50">
              <span v-if="builtIn" class="badge badge-sm badge-outline">Built-in</span>
              <span v-else class="badge badge-sm badge-outline badge-primary">User</span>
              <span class="font-mono">{{ rule.id }}</span>
            </div>
            <h2 class="card-title">{{ rule.title }}</h2>
            <p v-if="rule.description" class="mt-1 text-sm text-base-content/60">
              {{ rule.description }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <span class="badge" :class="severityBadgeClass[rule.severity]">
              {{ rule.severity }}
            </span>
            <Button
              v-if="!builtIn"
              type="button"
              variant="ghost"
              size="sm"
              @click="emit('edit', rule)"
            >
              Edit
            </Button>
            <Button
              v-if="!builtIn"
              type="button"
              variant="ghost"
              size="icon-sm"
              class="text-error"
              title="Delete rule"
              @click="onDelete"
            >
              <Trash2 class="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <h3 class="card-title text-base">Check</h3>
        <div class="grid grid-cols-1 gap-3 text-sm md:grid-cols-3">
          <div>
            <div class="text-xs uppercase tracking-wide text-base-content/50">Resource</div>
            <div class="mt-1 font-mono">{{ rule.resourceType }}</div>
          </div>
          <div class="md:col-span-2">
            <div class="text-xs uppercase tracking-wide text-base-content/50">Field path</div>
            <div class="mt-1 font-mono">{{ rule.fieldPath }}</div>
          </div>
          <div>
            <div class="text-xs uppercase tracking-wide text-base-content/50">Expected</div>
            <div class="mt-1 font-mono">{{ rule.expectedValue }}</div>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow">
      <div class="card-body gap-3">
        <div class="flex items-center justify-between">
          <h3 class="card-title text-base">Findings for this rule</h3>
          <Button
            v-if="!isConnected"
            type="button"
            size="sm"
            :disabled="connecting"
            @click="onConnect"
          >
            <span v-if="connecting" class="loading loading-spinner loading-sm"></span>
            Connect to cluster
          </Button>
          <Button
            v-else
            type="button"
            size="sm"
            :disabled="scanning"
            @click="runScan"
          >
            <span v-if="scanning" class="loading loading-spinner loading-sm"></span>
            Run scan
          </Button>
        </div>

        <div v-if="!isConnected" class="flex items-center gap-2 text-sm text-base-content/60">
          <TriangleAlert class="size-4" />
          Connect to a cluster to scan for findings.
        </div>

        <div
          v-else-if="contextMatches === false"
          class="flex items-center gap-2 text-sm text-warning"
        >
          <TriangleAlert class="size-4" />
          This rule applies to namespaced resources, but the active connection uses a different context than the selected project.
        </div>

        <div v-else-if="scanError" class="alert alert-error text-sm">
          Scan failed: {{ scanError }}
        </div>

        <div v-else-if="matchingFindings.length === 0" class="text-sm text-base-content/50">
          No findings for this rule. Run a scan to check.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>Severity</th>
                <th>Resource</th>
                <th>Namespace</th>
                <th>Message</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="f in matchingFindings" :key="f.id">
                <td>
                  <span class="badge badge-sm" :class="severityBadgeClass[f.severity]">
                    {{ f.severity }}
                  </span>
                </td>
                <td>{{ f.resourceKind }}/{{ f.resourceName }}</td>
                <td>{{ f.namespace ?? "-" }}</td>
                <td class="max-w-xl text-sm">{{ f.message }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
