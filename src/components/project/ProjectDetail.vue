<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { invoke } from "@tauri-apps/api/core";
import { Box, CircleAlert, CircleCheck, ScanSearch } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useProjects, type Project } from "@/composables/useProjects";
import { severityBadgeVariant, type Severity } from "@/lib/severity";

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

const { isConnectedTo, setConnection, refreshConnection } = useProjects();

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
    <Card>
      <CardHeader>
        <CardDescription class="uppercase tracking-wide">
          {{ isK8s ? "Kubernetes cluster review" : "Container image review" }}
        </CardDescription>
        <CardTitle>{{ project.name }}</CardTitle>
        <CardDescription>
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
        </CardDescription>
        <CardAction>
          <Button variant="ghost" size="sm" @click="emit('edit', project)">
            Edit
          </Button>
        </CardAction>
      </CardHeader>
    </Card>

    <template v-if="isK8s">
      <Card>
        <CardHeader>
          <CardTitle>Cluster</CardTitle>
          <CardAction>
            <Button variant="ghost" size="sm" @click="onRefreshConnection">
              Refresh
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent class="flex flex-col gap-3">
          <Alert v-if="connected">
            <CircleCheck />
            <AlertTitle>Connected</AlertTitle>
          </Alert>
          <Alert v-else-if="connectError" variant="destructive">
            <CircleAlert />
            <AlertTitle>Failed to connect</AlertTitle>
            <AlertDescription>{{ connectError }}</AlertDescription>
          </Alert>
          <p v-else class="text-sm text-muted-foreground">
            Not connected to this project's context.
          </p>

          <div>
            <Button :disabled="connecting" @click="connect">
              <Spinner v-if="connecting" />
              {{ connected ? "Reconnect" : "Connect" }}
            </Button>
          </div>
        </CardContent>
      </Card>

      <Alert v-if="scanError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Scan failed</AlertTitle>
        <AlertDescription>{{ scanError }}</AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle>Findings</CardTitle>
          <CardAction>
            <Button
              variant="secondary"
              size="sm"
              :disabled="!connected || scanning"
              @click="runScan"
            >
              <Spinner v-if="scanning" />
              Run Scan
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div
            v-if="scanning"
            class="flex items-center gap-2 py-6 text-muted-foreground"
          >
            <Spinner /> Scanning cluster…
          </div>

          <Empty v-else-if="!hasScanned" class="border-0 py-6 md:py-8">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ScanSearch />
              </EmptyMedia>
              <EmptyTitle>No scan yet</EmptyTitle>
              <EmptyDescription>
                Connect to the cluster and run a scan to see findings.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>

          <Empty v-else-if="findings.length === 0" class="border-0 py-6 md:py-8">
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <CircleCheck />
              </EmptyMedia>
              <EmptyTitle>No findings</EmptyTitle>
              <EmptyDescription>
                The 4 fixed rules found nothing.
              </EmptyDescription>
            </EmptyHeader>
          </Empty>

          <Table v-else>
            <TableHeader>
              <TableRow>
                <TableHead>Severity</TableHead>
                <TableHead>Rule</TableHead>
                <TableHead>Resource</TableHead>
                <TableHead>Namespace</TableHead>
                <TableHead>Message</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="f in sortedFindings" :key="f.id">
                <TableCell>
                  <Badge :variant="severityBadgeVariant(f.severity)">
                    {{ f.severity }}
                  </Badge>
                </TableCell>
                <TableCell class="font-mono text-xs">{{ f.ruleId }}</TableCell>
                <TableCell>{{ f.resourceKind }}/{{ f.resourceName }}</TableCell>
                <TableCell>{{ f.namespace ?? "-" }}</TableCell>
                <TableCell class="max-w-xl">{{ f.message }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </template>

    <template v-else>
      <Empty>
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <Box />
          </EmptyMedia>
          <EmptyTitle>Container image scanning</EmptyTitle>
          <EmptyDescription>
            Image scanning isn't implemented yet. Once it lands, this is where
            findings for
            <span class="font-mono">{{ project.config.image }}</span>
            will show up.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
    </template>
  </div>
</template>
