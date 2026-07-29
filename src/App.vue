<script setup lang="ts">
import { onMounted } from "vue";
import {
  ProjectSidebar,
  ProjectDetail,
  ProjectEditor,
} from "@/components/project/index";
import { ref } from "vue";
import { useProjects, type Project } from "@/composables/useProjects";

const {
  loadProjects,
  refreshConnection,
  selectedProject,
} = useProjects();

const editorOpen = ref(false);
const editorTarget = ref<Project | null>(null);

function openNew() {
  editorTarget.value = null;
  editorOpen.value = true;
}

function openEdit(project: Project) {
  editorTarget.value = project;
  editorOpen.value = true;
}

onMounted(async () => {
  await Promise.all([loadProjects(), refreshConnection()]);
});
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden bg-base-200">
    <ProjectSidebar @new="openNew" @edit="openEdit" />

    <main class="flex-1 overflow-y-auto p-6">
      <div class="mx-auto flex max-w-5xl flex-col gap-6">
        <header>
          <h1 class="text-2xl font-bold">Corsair</h1>
          <p class="text-sm text-base-content/60">Kubernetes security scan</p>
        </header>

        <ProjectDetail
          v-if="selectedProject"
          :project="selectedProject"
          @edit="openEdit"
        />

        <div v-else class="card bg-base-100 shadow">
          <div class="card-body items-center text-center text-base-content/50">
            <h2 class="card-title text-base-content">No project selected</h2>
            <p>Create a project from the sidebar to get started.</p>
          </div>
        </div>
      </div>
    </main>

    <ProjectEditor
      v-if="editorOpen"
      :project="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
