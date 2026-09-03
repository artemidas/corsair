<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { CircleAlert, CircleCheck, ScanSearch } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { FindingsDataTable } from "@/components/findings";
import { defaultTrivyScanOptions } from "@/lib/trivy";
import { TrashIcon } from "@lucide/vue";
import ImageEngagementSummary from "./ImageEngagementSummary.vue";
import ScanHistorySelect from "./ScanHistorySelect.vue";
import TrivyScanOptions from "./TrivyScanOptions.vue";
import type { Project } from "@/composables/useProjects";
import { useScans } from "@/composables/useScans";

const props = defineProps<{
  project: Project;
}>();

const {
  scansFor,
  selectedScan,
  selectedScanId,
  findingsFor,
  loadScans,
  selectScan,
  runImageScan,
  deleteScan,
} = useScans();

const scanning = ref(false);
const scanError = ref("");
const loadingHistory = ref(false);
const historyError = ref("");
const trivyOptions = ref(defaultTrivyScanOptions());

const scans = computed(() => scansFor(props.project.id));
const currentScan = computed(() => selectedScan(props.project.id));
const currentScanId = computed(() => selectedScanId(props.project.id));
const findings = computed(() =>
  currentScan.value ? findingsFor(currentScan.value.id) : [],
);

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

async function onDeleteScan(scanId: string) {
  await deleteScan(props.project.id, scanId);
}

async function runScan() {
  scanning.value = true;
  scanError.value = "";
  try {
    await runImageScan(props.project.id, trivyOptions.value);
  } catch (err) {
    scanError.value = String(err);
  } finally {
    scanning.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <ImageEngagementSummary :project="project" />

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

    <TrivyScanOptions v-model="trivyOptions" />

    <div class="flex items-center justify-between gap-2">
      <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
        Scan history
      </h3>
      <div class="flex items-center gap-2">
        <Button v-if="currentScanId" variant="secondary" size="sm" @click="onDeleteScan(currentScanId)">
          <TrashIcon class="size-4" />
          Delete Scan
        </Button>
        <ScanHistorySelect
          v-if="scans.length && currentScanId"
          :scans="scans"
          :model-value="currentScanId"
          @update:model-value="onSelectScan"
        />
        <Button variant="secondary" size="sm" :disabled="scanning" @click="runScan">
          <Spinner v-if="scanning" />
          Run Trivy Scan
        </Button>
      </div>
    </div>

    <div
      v-if="scanning || loadingHistory"
      class="flex items-center gap-2 py-6 text-muted-foreground"
    >
      <Spinner />
      {{ scanning ? "Scanning images with Trivy…" : "Loading scan history…" }}
    </div>

    <Empty v-else-if="!currentScan" class="border-0 py-6 md:py-8">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <ScanSearch />
        </EmptyMedia>
        <EmptyTitle>No scan yet</EmptyTitle>
        <EmptyDescription>
          Configure Trivy options above, then run a scan to see findings.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>

    <Alert v-else-if="currentScan.status === 'failed'" variant="destructive">
      <CircleAlert />
      <AlertTitle>This scan failed</AlertTitle>
      <AlertDescription>
        {{ currentScan.error ?? "Trivy could not scan the configured images." }}
      </AlertDescription>
    </Alert>

    <Empty v-else-if="findings.length === 0" class="border-0 py-6 md:py-8">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <CircleCheck />
        </EmptyMedia>
        <EmptyTitle>No findings</EmptyTitle>
        <EmptyDescription>
          Trivy did not report any issues for the selected options.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>

    <FindingsDataTable
      v-else
      :data="findings"
      :project-id="project.id"
      :scan-id="currentScan.id"
    />
  </div>
</template>
