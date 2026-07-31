<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { ProjectDetail } from "@/components/project";
import { useProjects, type Project } from "@/composables/useProjects";

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const { projects, loading } = useProjects();

const project = computed<Project | null>(
  () => projects.value.find((p) => p.id === props.id) ?? null,
);

function onEdit(p: Project) {
  router.push({ name: "edit-project", params: { id: p.id } });
}
</script>

<template>
  <ProjectDetail v-if="project" :project="project" @edit="onEdit" />

  <div v-else-if="loading" class="text-sm text-muted-foreground">Loading…</div>

  <div v-else class="card bg-base-100 shadow">
    <div class="card-body items-center text-center text-base-content/50">
      <h2 class="card-title text-base-content">Project not found</h2>
      <p>
        The project
        <span class="font-mono">{{ id }}</span>
        doesn't exist anymore.
      </p>
      <button class="btn btn-sm btn-ghost mt-2" @click="router.push({ name: 'projects' })">
        Back to projects
      </button>
    </div>
  </div>
</template>
