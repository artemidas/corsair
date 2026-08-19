<script setup lang="ts">
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { formatScanLabel, type Scan } from "@/composables/useScans";

defineProps<{
  scans: Scan[];
  modelValue?: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

function onSelect(value: unknown) {
  if (typeof value === "string") emit("update:modelValue", value);
}
</script>

<template>
  <Select
    :model-value="modelValue"
    @update:model-value="onSelect"
  >
    <SelectTrigger class="h-8 w-64 text-xs">
      <SelectValue placeholder="Scan history" />
    </SelectTrigger>
    <SelectContent>
      <SelectGroup>
        <SelectItem
          v-for="scan in scans"
          :key="scan.id"
          :value="scan.id"
        >
          {{ formatScanLabel(scan) }}
        </SelectItem>
      </SelectGroup>
    </SelectContent>
  </Select>
</template>
