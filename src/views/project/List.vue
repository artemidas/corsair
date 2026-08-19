<script setup lang="ts">
import { RouterLink } from "vue-router";
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
import { useProjects, type Project } from "@/composables/useProjects";

const { projects, loading, loadError } = useProjects();

function iconFor(kind: Project["kind"]) {
  return kind === "kubernetesClusterReview" ? Hexagon : Box;
}

function subtitleFor(p: Project) {
  return p.kind === "kubernetesClusterReview"
    ? (p.config.context ?? "<active context>")
    : (p.config.image ?? "");
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">Projects</h1>
      <Button as-child size="sm">
        <RouterLink :to="{ name: 'new-project' }">
          <Plus />
          New project
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

    <Empty v-else-if="projects.length === 0">
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Box />
        </EmptyMedia>
        <EmptyTitle>No projects yet</EmptyTitle>
        <EmptyDescription>
          Create one to connect a cluster and run a scan.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button as-child size="sm">
          <RouterLink :to="{ name: 'new-project' }">
            <Plus />
            New project
          </RouterLink>
        </Button>
      </EmptyContent>
    </Empty>

    <div v-else class="grid gap-3 md:grid-cols-2">
      <RouterLink
        v-for="project in projects"
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
