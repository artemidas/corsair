<script setup lang="ts">
import { onMounted, ref } from "vue";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { ProjectDetail, ProjectEditor } from "@/components/project";
import { useProjects, type Project } from "@/composables/useProjects";

const { loadProjects, refreshConnection, selectedProject } = useProjects();

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
  <SidebarProvider>
    <AppSidebar @new="openNew" @edit="openEdit" />
    <SidebarInset>
      <header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
        <SidebarTrigger />
        <h1 class="text-lg font-semibold">Corsair</h1>
      </header>
      <main class="flex-1 overflow-y-auto bg-base-200 p-6">
        <div class="mx-auto flex max-w-5xl flex-col gap-4">
          <ProjectDetail
            v-if="selectedProject"
            :project="selectedProject"
            @edit="openEdit"
          />
          <div v-else class="card bg-base-100 shadow">
            <div class="card-body items-center text-center text-base-content/50">
              <h2 class="card-title text-base-content">No project selected</h2>
              <p>Pick a project from the sidebar or create a new one.</p>
            </div>
          </div>
        </div>
      </main>
    </SidebarInset>
    <ProjectEditor
      v-if="editorOpen"
      :project="editorTarget"
      @close="editorOpen = false"
    />
  </SidebarProvider>
</template>
