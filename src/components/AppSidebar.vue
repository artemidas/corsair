<script setup lang="ts">
import { BookCheck, Box, Hexagon, Moon, Plus, Sun, Trash2 } from "@lucide/vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupAction,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { useProjects, type Project } from "@/composables/useProjects";
import {
  useCustomRules,
  isBuiltIn,
  type CustomRule,
} from "@/composables/useCustomRules";
import { useTheme } from "@/composables/useTheme";
import { confirm } from "@tauri-apps/plugin-dialog";

const emit = defineEmits<{
  new: [];
  edit: [project: Project];
  newRule: [];
  editRule: [rule: CustomRule];
  selectRule: [rule: CustomRule];
}>();

const {
  projects,
  selectedProjectId,
  selectProject,
  deleteProject,
  loading,
  loadError,
} = useProjects();

const {
  rules,
  userRules,
  loading: rulesLoading,
  loadError: rulesLoadError,
  deleteRule,
} = useCustomRules();

const { theme, toggleTheme } = useTheme();

function iconFor(kind: Project["kind"]) {
  return kind === "kubernetesClusterReview" ? Hexagon : Box;
}

const selectedRuleId = defineModel<string | null>("selectedRuleId", { default: null });

function selectRule(rule: CustomRule) {
  selectedRuleId.value = rule.id;
  selectProject(null);
  emit("selectRule", rule);
}

function selectProjectAndClearRule(id: string | null) {
  selectedRuleId.value = null;
  selectProject(id);
}

async function onDelete(p: Project, ev: Event) {
  ev.stopPropagation();
  ev.preventDefault();
  const confirmed = await confirm(`Delete project "${p.name}"?`, { title: "Tauri", kind: "warning" });
  if (!confirmed) return;
  try {
    await deleteProject(p.id);
  } catch (err) {
    alert(String(err));
    console.error("Failed to delete project:", err);
  }
}

async function onDeleteRule(rule: CustomRule, ev: Event) {
  ev.stopPropagation();
  ev.preventDefault();
  const confirmed = await confirm(`Delete rule "${rule.title}"?`, {
    title: "Tauri",
    kind: "warning",
  });
  if (!confirmed) return;
  try {
    await deleteRule(rule.id);
    if (selectedRuleId.value === rule.id) {
      selectedRuleId.value = null;
    }
  } catch (err) {
    alert(String(err));
    console.error("Failed to delete rule:", err);
  }
}
</script>

<template>
  <Sidebar>
    <SidebarHeader>
      <div class="px-2 py-1">
        <div class="text-base font-semibold">Corsair</div>
        <div class="text-xs text-muted-foreground">Kubernetes security scan</div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupLabel>Rules</SidebarGroupLabel>
        <SidebarGroupAction title="New rule" @click="emit('newRule')">
          <Plus />
          <span class="sr-only">New rule</span>
        </SidebarGroupAction>
        <SidebarGroupContent>
          <div
            v-if="rulesLoadError"
            class="rounded-md p-2 text-xs text-destructive"
          >
            {{ rulesLoadError }}
          </div>
          <div
            v-else-if="rulesLoading && rules.length === 0"
            class="px-2 py-2 text-xs text-muted-foreground"
          >
            Loading…
          </div>
          <div
            v-else-if="rules.length === 0"
            class="px-2 py-2 text-xs text-muted-foreground"
          >
            No rules yet. Click + to add one.
          </div>
          <SidebarMenu v-else>
            <SidebarMenuItem v-for="r in rules" :key="r.id">
              <SidebarMenuButton
                :is-active="selectedRuleId === r.id"
                :tooltip="`${r.title} (${r.severity})`"
                @click="selectRule(r)"
              >
                <BookCheck />
                <span class="truncate">{{ r.title }}</span>
                <span
                  v-if="isBuiltIn(r)"
                  class="badge badge-xs badge-outline ml-auto shrink-0"
                >
                  built-in
                </span>
              </SidebarMenuButton>
              <SidebarMenuAction
                v-if="!isBuiltIn(r)"
                show-on-hover
                title="Delete rule"
                @click="onDeleteRule(r, $event)"
              >
                <Trash2 />
                <span class="sr-only">Delete</span>
              </SidebarMenuAction>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>

      <SidebarGroup>
        <SidebarGroupLabel>Projects</SidebarGroupLabel>
        <SidebarGroupAction title="New project" @click="emit('new')">
          <Plus />
          <span class="sr-only">New project</span>
        </SidebarGroupAction>
        <SidebarGroupContent>
          <div
            v-if="loadError"
            class="rounded-md p-2 text-xs text-destructive"
          >
            {{ loadError }}
          </div>
          <div
            v-else-if="loading && projects.length === 0"
            class="px-2 py-2 text-xs text-muted-foreground"
          >
            Loading…
          </div>
          <div
            v-else-if="projects.length === 0"
            class="px-2 py-2 text-xs text-muted-foreground"
          >
            No projects yet. Click + to add one.
          </div>
          <SidebarMenu v-else>
            <SidebarMenuItem v-for="p in projects" :key="p.id">
              <SidebarMenuButton
                :is-active="selectedProjectId === p.id && selectedRuleId === null"
                :tooltip="p.name"
                @click="selectProjectAndClearRule(p.id)"
              >
                <component :is="iconFor(p.kind)" />
                <span class="truncate">{{ p.name }}</span>
              </SidebarMenuButton>
              <SidebarMenuAction
                show-on-hover
                title="Delete project"
                @click="onDelete(p, $event)"
              >
                <Trash2 />
                <span class="sr-only">Delete</span>
              </SidebarMenuAction>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <div class="flex items-center justify-between gap-2 px-2">
        <span class="truncate text-xs text-muted-foreground">
          {{ projects.length }} project{{ projects.length === 1 ? "" : "s" }} ·
          {{ userRules.length }} rule{{ userRules.length === 1 ? "" : "s" }}
        </span>
        <Button
          variant="ghost"
          size="icon-sm"
          class="shrink-0"
          :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="toggleTheme"
        >
          <Sun v-if="theme === 'dark'" />
          <Moon v-else />
          <span class="sr-only">Toggle theme</span>
        </Button>
      </div>
    </SidebarFooter>
  </Sidebar>
</template>
