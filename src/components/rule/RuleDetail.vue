<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import {
  ArrowLeft,
  CircleAlert,
  ScanSearch,
  Trash2,
  TriangleAlert,
} from "@lucide/vue";
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
import { DataTable } from "@/components/ui/data-table";
import { ruleFindingsColumns } from "@/components/findings";
import {
  useRules,
  describeCheck,
  NEEDS_EXPECTED_VALUE,
  OPERATOR_LABEL,
  type Rule,
} from "@/composables/useRules";
import { useCluster } from "@/composables/useCluster";
import { useScans } from "@/composables/useScans";
import { severityBadgeVariant } from "@/lib/severity";
import type { Finding } from "@/lib/findings";
import { confirm } from "@tauri-apps/plugin-dialog";

const props = defineProps<{
  rule: Rule;
}>();

const emit = defineEmits<{
  back: [];
  edit: [rule: Rule];
}>();

const { deleteRule } = useRules();
const { isConnected } = useCluster();
const { previewScan } = useScans();

const needsExpected = computed(() =>
  NEEDS_EXPECTED_VALUE.includes(props.rule.operator),
);

const scanning = ref(false);
const scanError = ref("");
const actionError = ref("");
const findings = ref<Finding[]>([]);

async function runScan() {
  scanning.value = true;
  scanError.value = "";
  try {
    findings.value = await previewScan();
  } catch (err) {
    scanError.value = String(err);
  } finally {
    scanning.value = false;
  }
}

const matchingFindings = computed(() =>
  findings.value.filter((f) => f.ruleId === props.rule.ruleId),
);

function getRowId(row: Finding) {
  return row.id;
}

async function onDelete() {
  const confirmed = await confirm(`Delete rule "${props.rule.title}"?`, {
    title: "Delete rule",
    kind: "warning",
  });
  if (!confirmed) return;
  actionError.value = "";
  try {
    await deleteRule(props.rule.id);
    emit("back");
  } catch (err) {
    actionError.value = String(err);
  }
}

watch(
  () => props.rule.id,
  () => {
    findings.value = [];
    scanError.value = "";
    actionError.value = "";
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
          <CardDescription class="font-mono">{{ rule.ruleId }}</CardDescription>
        </div>
        <CardTitle>{{ rule.title }}</CardTitle>
        <CardDescription v-if="rule.description">
          {{ rule.description }}
        </CardDescription>
        <CardAction class="flex flex-wrap items-center justify-end gap-2">
          <Badge :variant="severityBadgeVariant(rule.severity)">
            {{ rule.severity }}
          </Badge>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            @click="emit('edit', rule)"
          >
            Edit
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            class="text-destructive hover:text-destructive"
            @click="onDelete"
          >
            <Trash2 />
            Delete
          </Button>
        </CardAction>
      </CardHeader>
    </Card>

    <Alert v-if="actionError" variant="destructive">
      <CircleAlert />
      <AlertTitle>Could not update rule</AlertTitle>
      <AlertDescription>{{ actionError }}</AlertDescription>
    </Alert>

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
              Operator
            </div>
            <div class="mt-1 font-mono">{{ OPERATOR_LABEL[rule.operator] }}</div>
          </div>
          <div v-if="needsExpected">
            <div class="text-xs uppercase tracking-wide text-muted-foreground">
              Expected
            </div>
            <div class="mt-1 font-mono">{{ rule.expectedValue }}</div>
          </div>
          <div class="md:col-span-3 font-mono text-xs text-muted-foreground">
            {{ describeCheck(rule) }}
          </div>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Findings for this rule</CardTitle>
        <CardAction>
          <Button
            v-if="isConnected"
            type="button"
            size="sm"
            :disabled="scanning"
            @click="runScan"
          >
            <Spinner v-if="scanning" />
            Run scan
          </Button>
          <Button v-else as-child size="sm">
            <RouterLink :to="{ name: 'settings' }">Connect in Settings</RouterLink>
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Alert v-if="!isConnected">
          <TriangleAlert />
          <AlertTitle>Not connected</AlertTitle>
          <AlertDescription>
            Connect to a cluster in
            <RouterLink :to="{ name: 'settings' }" class="underline">
              Settings
            </RouterLink>
            to scan for findings.
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

        <DataTable
          v-else
          :columns="ruleFindingsColumns"
          :data="matchingFindings"
          :get-row-id="getRowId"
          :initial-sorting="[{ id: 'severity', desc: false }]"
        />
      </CardContent>
    </Card>
  </div>
</template>
