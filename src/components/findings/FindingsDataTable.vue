<script setup lang="ts">
import { computed } from "vue";
import { DataTable } from "@/components/ui/data-table";
import { summarizeFindings, type Finding } from "@/lib/findings";
import { createFindingsSummaryColumns } from "./columns";

const props = defineProps<{
  data: Finding[];
  projectId: string;
  scanId: string;
}>();

const columns = computed(() =>
  createFindingsSummaryColumns(props.projectId, props.scanId),
);
const rows = computed(() => summarizeFindings(props.data));

function getRowId(row: { ruleId: string }) {
  return row.ruleId;
}
</script>

<template>
  <DataTable
    :columns="columns"
    :data="rows"
    :get-row-id="getRowId"
    :initial-sorting="[{ id: 'affectedResources', desc: true }]"
  />
</template>
