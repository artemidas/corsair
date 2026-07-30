<script setup lang="ts">
import { onMounted, ref } from "vue";
import { RouterView } from "vue-router";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { ProjectEditor } from "@/components/project";
import { RuleEditor } from "@/components/rule";
import { useProjects, type Project } from "@/composables/useProjects";
import { useCustomRules, type CustomRule } from "@/composables/useCustomRules";

const { loadProjects, refreshConnection } = useProjects();
const { loadRules } = useCustomRules();

const projectEditorOpen = ref(false);
const projectEditorTarget = ref<Project | null>(null);

const ruleEditorOpen = ref(false);
const ruleEditorTarget = ref<CustomRule | null>(null);

function openNewProject() {
  projectEditorTarget.value = null;
  projectEditorOpen.value = true;
}

function openEditProject(project: Project) {
  projectEditorTarget.value = project;
  projectEditorOpen.value = true;
}

function openNewRule() {
  ruleEditorTarget.value = null;
  ruleEditorOpen.value = true;
}

function openEditRule(rule: CustomRule) {
  ruleEditorTarget.value = rule;
  ruleEditorOpen.value = true;
}

onMounted(async () => {
  await Promise.all([loadProjects(), loadRules(), refreshConnection()]);
});
</script>

<template>
  <SidebarProvider>
    <AppSidebar
      @new="openNewProject"
      @edit="openEditProject"
      @newRule="openNewRule"
      @editRule="openEditRule"
    />
    <SidebarInset>
      <header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
        <SidebarTrigger />
        <h1 class="text-lg font-semibold">Corsair</h1>
      </header>
      <main class="flex-1 overflow-y-auto bg-base-200 p-6">
        <div class="mx-auto flex max-w-5xl flex-col gap-4">
          <RouterView />
        </div>
      </main>
    </SidebarInset>
    <ProjectEditor
      v-if="projectEditorOpen"
      :project="projectEditorTarget"
      @close="projectEditorOpen = false"
    />
    <RuleEditor
      v-if="ruleEditorOpen"
      :rule="ruleEditorTarget"
      @close="ruleEditorOpen = false"
    />
  </SidebarProvider>
</template>
