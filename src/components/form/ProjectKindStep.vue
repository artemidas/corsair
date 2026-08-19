<script setup lang="ts">
import { Box, Hexagon } from "@lucide/vue";
import type { ProjectKind } from "@/composables/useProjects";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const emit = defineEmits<{
  select: [kind: ProjectKind];
}>();
</script>

<template>
  <div class="flex flex-col gap-3">
    <p class="text-sm text-muted-foreground">
      What kind of project do you want to create?
    </p>
    <Card
      class="cursor-pointer py-4 transition-colors hover:bg-accent/50"
      role="button"
      tabindex="0"
      @click="emit('select', 'kubernetesClusterReview')"
      @keydown.enter="emit('select', 'kubernetesClusterReview')"
    >
      <CardHeader class="flex flex-row items-start gap-3">
        <div class="rounded-md bg-muted p-2 text-muted-foreground">
          <Hexagon class="size-5" />
        </div>
        <div>
          <CardTitle>Kubernetes cluster review</CardTitle>
          <CardDescription>
            Connect to a live cluster and scan it for misconfigurations.
          </CardDescription>
        </div>
      </CardHeader>
    </Card>
    <Card
      class="cursor-pointer py-4 transition-colors hover:bg-accent/50"
      role="button"
      tabindex="0"
      @click="emit('select', 'containerImageReview')"
      @keydown.enter="emit('select', 'containerImageReview')"
    >
      <CardHeader class="flex flex-row items-start gap-3">
        <div class="rounded-md bg-muted p-2 text-muted-foreground">
          <Box class="size-5" />
        </div>
        <div>
          <CardTitle>Container image review</CardTitle>
          <CardDescription>
            Inspect a container image for risks.
          </CardDescription>
        </div>
      </CardHeader>
    </Card>
  </div>
</template>
