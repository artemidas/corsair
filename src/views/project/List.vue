<script setup lang="ts">
import { computed } from "vue";
import { RouterLink, useRoute } from "vue-router";
import { Box, CircleAlert, Hexagon, Plus } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { imagesFor, useProjects, type Project } from "@/composables/useProjects";

const route = useRoute();
const { projects, loading, loadError } = useProjects();

const kind = computed(() => route.meta.projectKind ?? "kubernetesClusterReview");
const listed = computed(() =>
  projects.value.filter((p) => p.kind === kind.value),
);
const isImages = computed(() => kind.value === "containerImageReview");

const copy = computed(() =>
  isImages.value
    ? {
        newLabel: "New engagement",
        newTo: { name: "new-container-image-project" as const },
        emptyTitle: "No engagements yet",
        emptyDescription: "Create one, give it a name, and pick the images to review.",
        icon: Box,
      }
    : {
        newLabel: "New cluster",
        newTo: { name: "new-kubernetes-project" as const },
        emptyTitle: "No clusters yet",
        emptyDescription: "Create one to scan a connected Kubernetes cluster.",
        icon: Hexagon,
      },
);

function iconFor(projectKind: Project["kind"]) {
  return projectKind === "kubernetesClusterReview" ? Hexagon : Box;
}

function subtitleFor(p: Project) {
  if (p.kind === "kubernetesClusterReview") {
    return p.config.context ?? "<active context>";
  }
  const imgs = imagesFor(p.config);
  if (imgs.length === 0) return "";
  if (imgs.length === 1) return imgs[0];
  return `${imgs[0]} +${imgs.length - 1}`;
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex items-center justify-end">
      <Button as-child size="sm">
        <RouterLink :to="copy.newTo">
          <Plus />
          {{ copy.newLabel }}
        </RouterLink>
      </Button>
    </div>

    <Alert v-if="loadError" variant="destructive">
      <CircleAlert />
      <AlertTitle>Could not load projects</AlertTitle>
      <AlertDescription>{{ loadError }}</AlertDescription>
    </Alert>

    <div v-else-if="loading && projects.length === 0" class="grid gap-3 md:grid-cols-2">
      <Skeleton class="h-24 w-full" />
      <Skeleton class="h-24 w-full" />
    </div>

    <Empty v-else-if="listed.length === 0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <component :is="copy.icon" />
        </EmptyMedia>
        <EmptyTitle>{{ copy.emptyTitle }}</EmptyTitle>
        <EmptyDescription>
          {{ copy.emptyDescription }}
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button as-child size="sm">
          <RouterLink :to="copy.newTo">
            <Plus />
            {{ copy.newLabel }}
          </RouterLink>
        </Button>
      </EmptyContent>
    </Empty>

    <div v-else class="grid gap-3 md:grid-cols-2">
      <RouterLink
        v-for="project in listed"
        :key="project.id"
        :to="{ name: 'project', params: { id: project.id } }"
        class="block"
      >
        <Card class="h-full py-4 transition-colors hover:bg-accent/50">
          <CardHeader class="flex flex-row items-center gap-3">
            <component
              :is="iconFor(project.kind)"
              class="size-5 shrink-0 text-muted-foreground"
            />
            <div class="min-w-0">
              <CardTitle class="truncate">{{ project.name }}</CardTitle>
              <CardDescription class="truncate font-mono">
                {{ subtitleFor(project) }}
              </CardDescription>
            </div>
          </CardHeader>
        </Card>
      </RouterLink>
    </div>
  </div>
</template>
