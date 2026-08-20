<script setup lang="ts">
import { Badge } from "@/components/ui/badge";
import { isBuiltIn, type CustomRule } from "@/composables/useCustomRules";

defineProps<{
  rule: CustomRule;
}>();
</script>

<template>
  <div class="flex min-w-0 flex-col gap-0.5 px-3 py-2">
    <div class="flex items-center gap-2">
      <span class="truncate font-medium">{{ rule.title }}</span>
      <Badge v-if="isBuiltIn(rule)" variant="outline">Built-in</Badge>
    </div>
  </div>
</template>
