<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { ProjectDetail, ProjectEditor } from "@/components/project";
import { RuleDetail, RuleEditor } from "@/components/rule";
import { useProjects, type Project } from "@/composables/useProjects";
import { useCustomRules, type CustomRule } from "@/composables/useCustomRules";

const {
  loadProjects,
  refreshConnection,
  selectedProject,
} = useProjects();

const { loadRules, rules } = useCustomRules();

const projectEditorOpen = ref(false);
const projectEditorTarget = ref<Project | null>(null);

const ruleEditorOpen = ref(false);
const ruleEditorTarget = ref<CustomRule | null>(null);

const selectedRuleId = ref<string | null>(null);
const selectedRule = computed<CustomRule | null>(
  () => rules.value.find((r) => r.id === selectedRuleId.value) ?? null,
);

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

function onSelectRule(rule: CustomRule) {
  selectedRuleId.value = rule.id;
}

onMounted(async () => {
  await Promise.all([loadProjects(), loadRules(), refreshConnection()]);
});
</script>

<template>
  <SidebarProvider>
    <AppSidebar
      v-model:selectedRuleId="selectedRuleId"
      @new="openNewProject"
      @edit="openEditProject"
      @newRule="openNewRule"
      @editRule="openEditRule"
      @selectRule="onSelectRule"
    />
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
            @edit="openEditProject"
          />
          <RuleDetail
            v-else-if="selectedRule"
            :rule="selectedRule"
            :selected-project="selectedProject"
            @edit="openEditRule"
            @back="selectedRuleId = null"
          />
          <div v-else class="card bg-base-100 shadow">
            <div class="card-body items-center text-center text-base-content/50">
              <h2 class="card-title text-base-content">Nothing selected</h2>
              <p>Pick a project or a rule from the sidebar to get started.</p>
            </div>
          </div>
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
