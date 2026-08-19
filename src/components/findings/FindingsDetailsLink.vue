<script setup lang="ts">
import { RouterLink } from "vue-router";
import { Button } from "@/components/ui/button";

defineProps<{
  projectId: string;
  scanId: string;
  ruleId: string;
}>();
</script>

<template>
  <Button variant="link" as-child>
    <RouterLink
      :to="{
        name: 'finding',
        params: { id: projectId, scanId, ruleId },
      }"
    >
      Details
    </RouterLink>
  </Button>
</template>
