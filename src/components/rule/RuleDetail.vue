<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { invoke } from "@tauri-apps/api/core";
import { ArrowLeft, CircleAlert, ScanSearch, Trash2, TriangleAlert } from "@lucide/vue";
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
import { useCustomRules, isBuiltIn, type CustomRule } from "@/composables/useCustomRules";
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
  if (!want) return true;
  return activeContext.value.context === (props.selectedProject?.config.context ?? null);
});

const matchingFindings = computed(() =>
  findings.value.filter((f) => f.ruleId === props.rule.id),
);

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
      class="-ml-3 self-start"
      @click="emit('back')"
    >
      <ArrowLeft />
      Back to rules
    </Button>

    <Card>
      <CardHeader>
        <div class="flex items-center gap-2">
          <Badge v-if="builtIn" variant="outline">Built-in</Badge>
          <Badge v-else variant="secondary">User</Badge>
          <CardDescription class="font-mono">{{ rule.id }}</CardDescription>
        </div>
        <CardTitle>{{ rule.title }}</CardTitle>
        <CardDescription v-if="rule.description">
          {{ rule.description }}
        </CardDescription>
        <CardAction class="flex items-center gap-2">
          <Badge :variant="severityBadgeVariant(rule.severity)">
            {{ rule.severity }}
          </Badge>
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
            class="text-destructive"
            title="Delete rule"
            @click="onDelete"
          >
            <Trash2 />
          </Button>
        </CardAction>
      </CardHeader>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Check</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="grid grid-cols-1 gap-3 text-sm md:grid-cols-3">
          <div>
            <div class="text-xs uppercase tracking-wide text-muted-foreground">
              Resource
            </div>
            <div class="mt-1 font-mono">{{ rule.resourceType }}</div>
          </div>
          <div class="md:col-span-2">
            <div class="text-xs uppercase tracking-wide text-muted-foreground">
              Field path
            </div>
            <div class="mt-1 font-mono">{{ rule.fieldPath }}</div>
          </div>
          <div>
            <div class="text-xs uppercase tracking-wide text-muted-foreground">
              Expected
            </div>
            <div class="mt-1 font-mono">{{ rule.expectedValue }}</div>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Findings for this rule</CardTitle>
        <CardAction>
          <Button
            v-if="!isConnected"
            type="button"
            size="sm"
            :disabled="connecting"
            @click="onConnect"
          >
            <Spinner v-if="connecting" />
            Connect to cluster
          </Button>
          <Button
            v-else
            type="button"
            size="sm"
            :disabled="scanning"
            @click="runScan"
          >
            <Spinner v-if="scanning" />
            Run scan
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Alert v-if="!isConnected">
          <TriangleAlert />
          <AlertTitle>Not connected</AlertTitle>
          <AlertDescription>
            Connect to a cluster to scan for findings.
          </AlertDescription>
        </Alert>

        <Alert v-else-if="contextMatches === false">
          <TriangleAlert />
          <AlertTitle>Context mismatch</AlertTitle>
          <AlertDescription>
            This rule applies to namespaced resources, but the active connection
            uses a different context than the selected project.
          </AlertDescription>
        </Alert>

        <Alert v-else-if="scanError" variant="destructive">
          <CircleAlert />
          <AlertTitle>Scan failed</AlertTitle>
          <AlertDescription>{{ scanError }}</AlertDescription>
        </Alert>

        <Empty
          v-else-if="matchingFindings.length === 0"
          class="border-0 py-6 md:py-8"
        >
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <ScanSearch />
            </EmptyMedia>
            <EmptyTitle>No findings for this rule</EmptyTitle>
            <EmptyDescription>Run a scan to check.</EmptyDescription>
          </EmptyHeader>
        </Empty>

        <Table v-else>
          <TableHeader>
            <TableRow>
              <TableHead>Severity</TableHead>
              <TableHead>Resource</TableHead>
              <TableHead>Namespace</TableHead>
              <TableHead>Message</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="f in matchingFindings" :key="f.id">
              <TableCell>
                <Badge :variant="severityBadgeVariant(f.severity)">
                  {{ f.severity }}
                </Badge>
              </TableCell>
              <TableCell>{{ f.resourceKind }}/{{ f.resourceName }}</TableCell>
              <TableCell>{{ f.namespace ?? "-" }}</TableCell>
              <TableCell class="max-w-xl">{{ f.message }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
