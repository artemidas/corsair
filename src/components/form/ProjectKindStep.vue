<script setup lang="ts">
import { Box, Hexagon } from "@lucide/vue";
import type { ProjectKind } from "@/composables/useProjects";

const emit = defineEmits<{
  select: [kind: ProjectKind];
}>();
</script>

<template>
  <div class="mt-4 flex flex-col gap-3">
    <p class="text-sm text-muted-foreground">
      What kind of project do you want to create?
    </p>
    <button
      type="button"
      class="group flex items-start gap-3 rounded-lg border border-input p-4 text-left transition-colors hover:border-primary hover:bg-primary/5"
      @click="emit('select', 'kubernetesClusterReview')"
    >
      <div
        class="rounded-md bg-primary/10 p-2 text-primary transition-colors group-hover:bg-primary/20"
      >
        <Hexagon class="size-5" />
      </div>
      <div>
        <div class="text-sm font-medium">Kubernetes cluster review</div>
        <div class="text-xs text-muted-foreground">
          Connect to a live cluster and scan it for misconfigurations.
        </div>
      </div>
    </button>
    <button
      type="button"
      class="group flex items-start gap-3 rounded-lg border border-input p-4 text-left transition-colors hover:border-secondary hover:bg-secondary/5"
      @click="emit('select', 'containerImageReview')"
    >
      <div
        class="rounded-md bg-secondary/10 p-2 text-secondary transition-colors group-hover:bg-secondary/20"
      >
        <Box class="size-5" />
      </div>
      <div>
        <div class="text-sm font-medium">Container image review</div>
        <div class="text-xs text-muted-foreground">
          Inspect a container image for risks.
        </div>
      </div>
    </button>
  </div>
</template>
