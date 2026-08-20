<script setup lang="ts">
import { Trash2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import type { Rule } from "@/composables/useRules";

const props = defineProps<{
  rule: Rule;
}>();

const emit = defineEmits<{
  delete: [rule: Rule];
}>();
</script>

<template>
  <Button
    variant="ghost"
    size="icon-sm"
    class="text-destructive hover:text-destructive"
    title="Delete rule"
    @click.stop="emit('delete', props.rule)"
  >
    <Trash2 />
  </Button>
</template>
