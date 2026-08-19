<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import { Unplug, CircleCheck } from "@lucide/vue";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useCluster } from "@/composables/useCluster";

const { status, isConnected, contextLabel } = useCluster();

const label = computed(() => {
  if (isConnected.value) return contextLabel.value;
  if (status.value.connected) return `${contextLabel.value} · unreachable`;
  return "No cluster";
});

const tooltip = computed(() => {
  if (isConnected.value) {
    return `Connected to ${contextLabel.value}`;
  }
  if (status.value.connected && status.value.error) {
    return status.value.error;
  }
  return "Not connected — open Settings to connect";
});
</script>

<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <RouterLink
        :to="{ name: 'settings' }"
        class="inline-flex max-w-56"
      >
        <Badge variant="outline" class="max-w-full">
          <CircleCheck v-if="isConnected" color="green" />
          <Unplug v-else />
          <span class="truncate font-mono" aria-live="polite">{{ label }}</span>
        </Badge>
      </RouterLink>
    </TooltipTrigger>
    <TooltipContent>{{ tooltip }}</TooltipContent>
  </Tooltip>
</template>
