<script setup lang="ts">
import { computed } from "vue";
import { DataTable } from "@/components/ui/data-table";
import type { Rule } from "@/composables/useRules";
import { createRulesColumns } from "./columns";

defineProps<{
  data: Rule[];
}>();

const emit = defineEmits<{
  rowClick: [rule: Rule];
  delete: [rule: Rule];
}>();

const columns = computed(() =>
  createRulesColumns({
    onDelete,
  }),
);

function getRowId(row: Rule) {
  return row.id;
}

function onRowClick(rule: Rule) {
  emit("rowClick", rule);
}

function onDelete(rule: Rule) {
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
