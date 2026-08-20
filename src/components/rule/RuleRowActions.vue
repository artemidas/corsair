<script setup lang="ts">
import { Trash2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { isBuiltIn, type CustomRule } from "@/composables/useCustomRules";

const props = defineProps<{
  rule: CustomRule;
}>();

const emit = defineEmits<{
  delete: [rule: CustomRule];
}>();
</script>

<template>
  <Button
    v-if="!isBuiltIn(rule)"
    variant="ghost"
    size="icon-sm"
    class="text-destructive hover:text-destructive"
    title="Delete rule"
    @click.stop="emit('delete', props.rule)"
  >
    <Trash2 />
  </Button>
</template>
