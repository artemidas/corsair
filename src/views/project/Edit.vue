<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { ProjectEditor } from "@/components/project";
import { useProjects, type Project } from "@/composables/useProjects";

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const { projects, loading } = useProjects();

const project = computed<Project | null>(
  () => projects.value.find((p) => p.id === props.id) ?? null,
);

function back() {
  router.push({ name: "project", params: { id: props.id } });
}
</script>

<template>
  <ProjectEditor v-if="project" :project="project" @close="back" />

  <div v-else-if="loading" class="text-sm text-muted-foreground">Loading…</div>

  <div v-else class="card bg-base-100 shadow">
    <div class="card-body items-center text-center text-base-content/50">
      <h2 class="card-title text-base-content">Project not found</h2>
      <button class="btn btn-sm btn-ghost mt-2" @click="router.push({ name: 'projects' })">
        Back to projects
      </button>
    </div>
  </div>
</template>
