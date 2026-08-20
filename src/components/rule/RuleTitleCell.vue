<script setup lang="ts">
import type { Rule } from "@/composables/useRules";

defineProps<{
  rule: Rule;
}>();
</script>

<template>
  <div class="flex min-w-0 flex-col gap-0.5 px-3 py-2">
    <span class="truncate font-medium">{{ rule.title }}</span>
  </div>
</template>
