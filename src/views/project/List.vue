<script setup lang="ts">
import { RouterLink } from "vue-router";
import { Box, Hexagon, Plus } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { useProjects, type Project } from "@/composables/useProjects";

const { projects, loading, loadError } = useProjects();

function iconFor(kind: Project["kind"]) {
  return kind === "kubernetesClusterReview" ? Hexagon : Box;
}

function subtitleFor(p: Project) {
  return p.kind === "kubernetesClusterReview"
    ? p.config.context ?? "<active context>"
    : p.config.image ?? "";
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <h1 class="text-xl font-semibold">Projects</h1>
      <Button as-child size="sm">
        <RouterLink :to="{ name: 'new-project' }">
          <Plus class="size-4" />
          New project
        </RouterLink>
      </Button>
    </div>

    <div v-if="loadError" class="alert alert-error text-sm">
      {{ loadError }}
    </div>

    <div v-else-if="loading && projects.length === 0" class="text-sm text-muted-foreground">
      Loading…
    </div>

    <div v-else-if="projects.length === 0" class="card bg-base-100 shadow">
      <div class="card-body items-center gap-2 text-center text-base-content/50">
        <h2 class="card-title text-base text-base-content">No projects yet</h2>
        <p class="text-sm">Create one to connect a cluster and run a scan.</p>
      </div>
    </div>

    <div v-else class="grid gap-3 md:grid-cols-2">
      <RouterLink
        v-for="project in projects"
        :key="project.id"
        :to="{ name: 'project', params: { id: project.id } }"
        class="card bg-base-100 shadow transition-colors hover:bg-base-200"
      >
        <div class="card-body flex-row items-center gap-3 py-4">
          <component :is="iconFor(project.kind)" class="size-5 shrink-0 text-primary" />
          <div class="min-w-0">
            <div class="truncate font-medium">{{ project.name }}</div>
            <div class="truncate font-mono text-xs text-base-content/50">
              {{ subtitleFor(project) }}
            </div>
          </div>
        </div>
      </RouterLink>
    </div>
  </div>
</template>
