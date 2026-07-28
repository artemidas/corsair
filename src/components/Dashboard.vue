<script setup lang="ts">
import { computed, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";

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

type ConnectionState = "disconnected" | "connecting" | "connected" | "error";

const connectionState = ref<ConnectionState>("disconnected");
const connectError = ref("");
const scanning = ref(false);
const scanError = ref("");
const findings = ref<Finding[]>([]);
const hasScanned = ref(false);

const isConnected = computed(() => connectionState.value === "connected");

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
  [...findings.value].sort((a, b) => severityOrder[a.severity] - severityOrder[b.severity]),
);

async function connect() {
  connectionState.value = "connecting";
  connectError.value = "";
  try {
    await invoke("connect_cluster");
    connectionState.value = "connected";
  } catch (err) {
    connectionState.value = "error";
    connectError.value = String(err);
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
</script>

<template>
  <main class="min-h-screen bg-base-200 p-6">
    <div class="mx-auto flex max-w-5xl flex-col gap-6">
      <header class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold">Corsair</h1>
          <p class="text-sm text-base-content/60">Kubernetes security scan</p>
        </div>

        <div class="flex items-center gap-3">
          <button
            class="btn btn-primary"
            :class="{ 'btn-disabled': connectionState === 'connecting' }"
            @click="connect"
          >
            <span v-if="connectionState === 'connecting'" class="loading loading-spinner loading-sm"></span>
            {{ isConnected ? "Reconnect" : "Connect" }}
          </button>

          <button class="btn btn-secondary" :disabled="!isConnected || scanning" @click="runScan">
            <span v-if="scanning" class="loading loading-spinner loading-sm"></span>
            Run Scan
          </button>
        </div>
      </header>

      <div v-if="isConnected" class="alert alert-success py-2 text-sm">
        Connected to cluster (active kubeconfig context).
      </div>
      <div v-else-if="connectionState === 'error'" class="alert alert-error text-sm">
        Failed to connect: {{ connectError }}
      </div>

      <div v-if="scanError" class="alert alert-error text-sm">Scan failed: {{ scanError }}</div>

      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <div class="flex items-center justify-between">
            <h2 class="card-title">Findings</h2>
            <span class="text-sm text-base-content/60">{{ findings.length }} finding(s)</span>
          </div>

          <div v-if="scanning" class="flex items-center gap-2 py-6 text-base-content/60">
            <span class="loading loading-spinner"></span> Scanning cluster…
          </div>

          <div v-else-if="!hasScanned" class="py-6 text-base-content/50">
            Connect to a cluster and run a scan to see findings.
          </div>

          <div v-else-if="findings.length === 0" class="py-6 text-base-content/50">
            No findings — the 4 fixed rules found nothing.
          </div>

          <div v-else class="overflow-x-auto">
            <table class="table">
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
                <tr v-for="finding in sortedFindings" :key="finding.id">
                  <td>
                    <span class="badge" :class="severityBadgeClass[finding.severity]">
                      {{ finding.severity }}
                    </span>
                  </td>
                  <td class="font-mono text-xs">{{ finding.ruleId }}</td>
                  <td>{{ finding.resourceKind }}/{{ finding.resourceName }}</td>
                  <td>{{ finding.namespace ?? "-" }}</td>
                  <td class="max-w-xl text-sm">{{ finding.message }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </main>
</template>
