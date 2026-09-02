<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { ArrowLeft, CircleAlert, ScanSearch } from "@lucide/vue";
import { Item, ItemContent, ItemTitle, ItemDescription, ItemActions } from "@/components/ui/item";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { DataTable } from "@/components/ui/data-table";
import { useScans, type Scan } from "@/composables/useScans";
import { severityBadgeVariant } from "@/lib/severity";
import { summarizeFindings, type Finding } from "@/lib/findings";
import { createRuleFindingsColumns } from "./columns";

const props = defineProps<{
  projectId: string;
  scanId: string;
  ruleId: string;
}>();

const emit = defineEmits<{
  back: [];
}>();

const { getScan, loadFindings } = useScans();

const loading = ref(true);
const loadError = ref("");
const scan = ref<Scan | null>(null);
const findings = ref<Finding[]>([]);

const items = computed(() =>
  findings.value.filter((finding) => finding.ruleId === props.ruleId),
);
const summary = computed(() => summarizeFindings(items.value)[0] ?? null);
const scanBelongsToProject = computed(
  () => scan.value?.projectId === props.projectId,
);

const namespaces = computed(() => {
  const values = new Set(
    items.value
      .map((finding) => finding.namespace)
      .filter((ns): ns is string => Boolean(ns)),
  );
  return [...values].sort();
});

const message = computed(() => items.value[0]?.message ?? "");

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    const [nextScan, nextFindings] = await Promise.all([
      getScan(props.scanId),
      loadFindings(props.scanId),
    ]);
    scan.value = nextScan;
    findings.value = nextFindings;
  } catch (err) {
    loadError.value = String(err);
    scan.value = null;
    findings.value = [];
  } finally {
    loading.value = false;
  }
}

watch(
  () => [props.scanId, props.ruleId, props.projectId],
  () => void load(),
  { immediate: true },
);

function getRowId(row: Finding) {
  return row.id;
}
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
      Back to project
    </Button>

    <div v-if="loading" class="flex items-center gap-2 py-6 text-muted-foreground">
      <Spinner /> Loading finding…
    </div>

    <Alert v-else-if="loadError" variant="destructive">
      <CircleAlert />
      <AlertTitle>Could not load finding</AlertTitle>
      <AlertDescription>{{ loadError }}</AlertDescription>
    </Alert>

    <Empty v-else-if="!scan || !scanBelongsToProject || !summary">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <ScanSearch />
        </EmptyMedia>
        <EmptyTitle>Finding not found</EmptyTitle>
        <EmptyDescription>
          This scan may have been deleted, or this rule had no hits in it.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>

    <template v-else>
      <Item>
        <ItemContent>
          <ItemTitle>
            <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
              {{ summary.ruleTitle }}
            </h3>
          </ItemTitle>
          <ItemDescription class="line-clamp-none">
            <p v-if="message" class="leading-7 [&:not(:first-child)]:mt-6">
              {{ message }}
            </p>
            <ul class="my-3 ml-6 list-disc [&>li]:mt-2">
              <li v-if="summary.affectedResources">
                {{ summary.affectedResources }} affected {{ summary.affectedResources === 1 ? "resource" : "resources" }}
              </li>
              <li v-if="namespaces.length">
                <span class="font-medium">Namespaces:</span>
                <span class="font-mono">{{ namespaces.join(", ") }}</span>
              </li>
            </ul>
          </ItemDescription>
        </ItemContent>
        <ItemActions>
          <Badge :variant="severityBadgeVariant(summary.severity)">
            {{ summary.severity }}
          </Badge>
        </ItemActions>
      </Item>
      <Item>
        <ItemContent>
          <ItemTitle class="mb-2">
            <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
              Affected resources
            </h3>
          </ItemTitle>
          <ItemDescription class="line-clamp-none">
            <DataTable
              :columns="createRuleFindingsColumns()"
              :data="items"
              :get-row-id="getRowId"
            />
          </ItemDescription>
        </ItemContent>
      </Item>        
    </template>
  </div>
</template>
