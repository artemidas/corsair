<script setup lang="ts">
import { computed } from "vue";
import { DataTable } from "@/components/ui/data-table";
import type { CustomRule } from "@/composables/useCustomRules";
import { createRulesColumns } from "./columns";

defineProps<{
  data: CustomRule[];
}>();

const emit = defineEmits<{
  rowClick: [rule: CustomRule];
  delete: [rule: CustomRule];
}>();

const columns = computed(() =>
  createRulesColumns({
    onDelete,
  }),
);

function getRowId(row: CustomRule) {
  return row.id;
}

function onRowClick(rule: CustomRule) {
  emit("rowClick", rule);
}

function onDelete(rule: CustomRule) {
  emit("delete", rule);
}
</script>

<template>
  <DataTable
    :columns="columns"
    :data="data"
    :get-row-id="getRowId"
    :initial-sorting="[{ id: 'title', desc: false }]"
    :row-click="onRowClick"
  />
</template>
