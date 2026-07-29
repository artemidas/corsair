<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { invoke } from "@tauri-apps/api/core";
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

const props = defineProps<{
  project: Project;
}>();

const emit = defineEmits<{
  edit: [project: Project];
}>();

const {
  isConnectedTo,
  setConnection,
  refreshConnection,
} = useProjects();

const connecting = ref(false);
const connectError = ref("");
const scanning = ref(false);
const scanError = ref("");
const findings = ref<Finding[]>([]);
const hasScanned = ref(false);

const isK8s = computed(() => props.project.kind === "kubernetesClusterReview");
const connected = computed(() => isConnectedTo(props.project));

const severityOrder: Record<Severity, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
};
const severityBadgeClass: Record<Severity, string> = {
  critical: "bg-red-600 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-400 text-black",
  low: "bg-gray-400 text-black",
};
const sortedFindings = computed(() =>
  [...findings.value].sort(
    (a, b) => severityOrder[a.severity] - severityOrder[b.severity],
  ),
);

watch(
  () => props.project.id,
  () => {
    connecting.value = false;
    connectError.value = "";
    scanning.value = false;
    scanError.value = "";
    findings.value = [];
    hasScanned.value = false;
  },
);

async function connect() {
  if (!isK8s.value) return;
  connecting.value = true;
  connectError.value = "";
  try {
    const ctx = props.project.config.context;
    await invoke("connect_cluster", { context: ctx });
    setConnection(ctx ?? null);
  } catch (err) {
    connectError.value = String(err);
  } finally {
    connecting.value = false;
  }
}

async function runScan() {
  scanning.value = true;
  scanError.value = "";
  try {
    findings.value = await invoke<Finding[]>("run_scan");
    hasScanned.value = true;
  } catch (err) {
    scanError.value = String(err);
  } finally {
    scanning.value = false;
  }
}

async function onRefreshConnection() {
  await refreshConnection();
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="card bg-base-100 shadow">
      <div class="card-body">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="text-xs uppercase tracking-wide text-base-content/50">
              {{ isK8s ? "Kubernetes cluster review" : "Container image review" }}
            </div>
            <h2 class="card-title">{{ project.name }}</h2>
            <div class="mt-1 text-sm text-base-content/60">
              <template v-if="isK8s">
                Context:
                <span class="font-mono">
                  {{ project.config.context ?? "<active context>" }}
                </span>
              </template>
              <template v-else>
                Image:
                <span class="font-mono">{{ project.config.image }}</span>
              </template>
            </div>
          </div>
          <button class="btn btn-sm btn-ghost" @click="emit('edit', project)">
            Edit
          </button>
        </div>
      </div>
    </div>

    <template v-if="isK8s">
      <div class="card bg-base-100 shadow">
        <div class="card-body gap-3">
          <div class="flex items-center justify-between">
            <h3 class="card-title text-base">Cluster</h3>
            <button class="btn btn-sm btn-ghost" @click="onRefreshConnection">
              Refresh
            </button>
          </div>

          <div v-if="connected" class="alert alert-success py-2 text-sm">
            Connected.
          </div>
          <div
            v-else-if="connectError"
            class="alert alert-error text-sm"
          >
            Failed to connect: {{ connectError }}
          </div>
          <div v-else class="text-sm text-base-content/50">
            Not connected to this project's context.
          </div>

          <div>
            <button
              class="btn btn-primary"
              :disabled="connecting"
              @click="connect"
            >
              <span v-if="connecting" class="loading loading-spinner loading-sm"></span>
              {{ connected ? "Reconnect" : "Connect" }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="scanError" class="alert alert-error text-sm">
        Scan failed: {{ scanError }}
      </div>

      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <div class="flex items-center justify-between">
            <h3 class="card-title text-base">Findings</h3>
            <button
              class="btn btn-secondary btn-sm"
              :disabled="!connected || scanning"
              @click="runScan"
            >
              <span v-if="scanning" class="loading loading-spinner loading-sm"></span>
              Run Scan
            </button>
          </div>

          <div v-if="scanning" class="flex items-center gap-2 py-6 text-base-content/60">
            <span class="loading loading-spinner"></span> Scanning cluster…
          </div>

          <div v-else-if="!hasScanned" class="py-6 text-sm text-base-content/50">
            Connect to the cluster and run a scan to see findings.
          </div>

          <div v-else-if="findings.length === 0" class="py-6 text-sm text-base-content/50">
            No findings — the 4 fixed rules found nothing.
          </div>

          <div v-else class="overflow-x-auto">
            <table class="table table-sm">
              <thead>
                <tr>
                  <th>Severity</th>
                  <th>Rule</th>
                  <th>Resource</th>
                  <th>Namespace</th>
                  <th>Message</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="f in sortedFindings" :key="f.id">
                  <td>
                    <span class="badge" :class="severityBadgeClass[f.severity]">
                      {{ f.severity }}
                    </span>
                  </td>
                  <td class="font-mono text-xs">{{ f.ruleId }}</td>
                  <td>{{ f.resourceKind }}/{{ f.resourceName }}</td>
                  <td>{{ f.namespace ?? "-" }}</td>
                  <td class="max-w-xl text-sm">{{ f.message }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <template v-else>
      <div class="card bg-base-100 shadow">
        <div class="card-body items-center gap-3 text-center">
          <h3 class="card-title text-base">Container image scanning</h3>
          <p class="text-sm text-base-content/50">
            Image scanning isn't implemented yet. Once it lands, this is where
            findings for <span class="font-mono">{{ project.config.image }}</span>
            will show up.
          </p>
        </div>
      </div>
    </template>
  </div>
</template>
