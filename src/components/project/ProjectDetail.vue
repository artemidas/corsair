<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { RouterLink } from "vue-router";
import { CircleAlert, CircleCheck, ScanSearch, TriangleAlert } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Item, ItemContent, ItemTitle, ItemDescription, ItemActions } from "@/components/ui/item";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { FindingsDataTable } from "@/components/findings";
import ScanHistorySelect from "./ScanHistorySelect.vue";
import ImageProjectDetail from "./ImageProjectDetail.vue";
import type { Project } from "@/composables/useProjects";
import { useCluster } from "@/composables/useCluster";
import { useScans } from "@/composables/useScans";

const props = defineProps<{
  project: Project;
}>();

defineEmits<{
  edit: [project: Project];
}>();

const { isConnected, status, contextLabel } = useCluster();
const {
  scansFor,
  selectedScan,
  selectedScanId,
  findingsFor,
  loadScans,
  selectScan,
  runScan: persistScan,
} = useScans();

const scanning = ref(false);
const scanError = ref("");
const loadingHistory = ref(false);
const historyError = ref("");

const isK8s = computed(() => props.project.kind === "kubernetesClusterReview");
const scans = computed(() => scansFor(props.project.id));
const currentScan = computed(() => selectedScan(props.project.id));
const currentScanId = computed(() => selectedScanId(props.project.id));
const findings = computed(() =>
  currentScan.value ? findingsFor(currentScan.value.id) : [],
);
const contextMismatch = computed(() => {
  const pinned = props.project.config.context;
  if (!isK8s.value || !isConnected.value || !pinned) return false;
  return status.value.context !== pinned;
});

watch(
  () => props.project.id,
  async (id) => {
    scanning.value = false;
    scanError.value = "";
    historyError.value = "";
    loadingHistory.value = true;
    try {
      await loadScans(id);
    } catch (err) {
      historyError.value = String(err);
    } finally {
      loadingHistory.value = false;
    }
  },
  { immediate: true },
);

async function onSelectScan(scanId: string) {
  await selectScan(props.project.id, scanId);
}

async function runScan() {
  scanning.value = true;
  scanError.value = "";
  try {
    await persistScan(props.project.id);
  } catch (err) {
    scanError.value = String(err);
  } finally {
    scanning.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <template v-if="isK8s">
      <Alert v-if="!isConnected">
        <TriangleAlert />
        <AlertTitle>Not connected</AlertTitle>
        <AlertDescription>
          Connect to a cluster in
          <RouterLink :to="{ name: 'settings' }" class="underline">
            Settings
          </RouterLink>
          before running a scan.
        </AlertDescription>
      </Alert>
      <Alert v-else-if="contextMismatch">
        <TriangleAlert />
        <AlertTitle>Context mismatch</AlertTitle>
        <AlertDescription>
          This project is pinned to
          <span class="font-mono">{{ project.config.context }}</span>,
          but the active connection is
          <span class="font-mono">{{ contextLabel }}</span>.
          Change it in
          <RouterLink :to="{ name: 'settings' }" class="underline">
            Settings
          </RouterLink>
          if that isn't what you meant.
        </AlertDescription>
      </Alert>

      <Alert v-if="historyError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Could not load scan history</AlertTitle>
        <AlertDescription>{{ historyError }}</AlertDescription>
      </Alert>

      <Alert v-if="scanError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Scan failed</AlertTitle>
        <AlertDescription>{{ scanError }}</AlertDescription>
      </Alert>

      <Item>
        <ItemContent>
          <ItemTitle>
            <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
              {{ project.name }}
            </h3>
          </ItemTitle>
          <ItemDescription class="line-clamp-none" v-if="currentScan">
            <template v-if="currentScan.status === 'failed'">
              This scan failed
              <template v-if="currentScan.context">
                on
                <span class="font-mono">{{ currentScan.context }}</span>
              </template>
            </template>
            <template v-else>
              {{ currentScan.findingCount }}
              {{ currentScan.findingCount === 1 ? "finding" : "findings" }}
            </template>
          </ItemDescription>
        </ItemContent>
      </Item>
          
      <Item>
        <ItemContent>
          <ItemTitle>
            <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
              Scan history
            </h3>
          </ItemTitle>
          <ItemDescription>
            <template v-if="isK8s">
              Context:
              <span class="font-mono">
                {{ project.config.context ?? "<active context>" }}
              </span>
            </template>
          </ItemDescription>
        </ItemContent>
        <ItemActions>
          <ScanHistorySelect
            v-if="scans.length && currentScanId"
            :scans="scans"
            :model-value="currentScanId"
            @update:model-value="onSelectScan"
          />
          <Button
            variant="secondary"
            size="sm"
            :disabled="!isConnected || scanning"
            @click="runScan"
          >
            <Spinner v-if="scanning" />
            Run Scan
          </Button>
        </ItemActions>
      </Item>
      <div
        v-if="scanning || loadingHistory"
        class="flex items-center gap-2 py-6 text-muted-foreground"
      >
        <Spinner />
        {{ scanning ? "Scanning cluster…" : "Loading scan history…" }}
      </div>

      <Empty v-else-if="!currentScan" class="border-0 py-6 md:py-8">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <ScanSearch />
          </EmptyMedia>
          <EmptyTitle>No scan yet</EmptyTitle>
          <EmptyDescription>
            Connect in Settings, then run a scan to see findings.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>

      <Alert v-else-if="currentScan.status === 'failed'" variant="destructive">
        <CircleAlert />
        <AlertTitle>This scan failed</AlertTitle>
        <AlertDescription>
          {{ currentScan.error ?? "The cluster could not be evaluated." }}
        </AlertDescription>
      </Alert>

      <Empty v-else-if="findings.length === 0" class="border-0 py-6 md:py-8">
        <EmptyHeader>
          <EmptyMedia variant="icon">
            <CircleCheck />
          </EmptyMedia>
          <EmptyTitle>No findings</EmptyTitle>
          <EmptyDescription>
            This scan did not report any rule hits.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>

      <FindingsDataTable
        v-else
        :data="findings"
        :project-id="project.id"
        :scan-id="currentScan.id"
      />
    </template>

    <template v-else>
      <ImageProjectDetail :project="project" />
    </template>
  </div>
</template>
