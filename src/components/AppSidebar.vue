<script setup lang="ts">
import { BookCheck, Box, Hexagon, Moon, Plus, Sun, Trash2 } from "@lucide/vue";
import { RouterLink, useRoute } from "vue-router";
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
} from "@/composables/useCustomRules";
import { useTheme } from "@/composables/useTheme";
import { confirm } from "@tauri-apps/plugin-dialog";

const emit = defineEmits<{
  new: [];
  edit: [project: Project];
  newRule: [];
}>();

const {
  projects,
  loading,
  loadError,
  deleteProject,
} = useProjects();

const {
  userRules,
} = useCustomRules();

const { theme, toggleTheme } = useTheme();

const route = useRoute();

function iconFor(kind: Project["kind"]) {
  return kind === "kubernetesClusterReview" ? Hexagon : Box;
}

async function onDelete(p: Project, ev: Event) {
  ev.stopPropagation();
  ev.preventDefault();
  const confirmed = await confirm(`Delete project "${p.name}"?`, {
    title: "Tauri",
    kind: "warning",
  });
  if (!confirmed) return;
  try {
    await deleteProject(p.id);
  } catch (err) {
    alert(String(err));
    console.error("Failed to delete project:", err);
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
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton as-child :is-active="route.name === 'rules'">
                <RouterLink to="/rules/built-in">
                  <BookCheck />
                  <span class="truncate">Built-in rules</span>
                </RouterLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
            <SidebarMenuItem>
              <SidebarMenuButton as-child :is-active="route.name === 'custom-rule'">
                <RouterLink to="/rules/custom">
                  <BookCheck />
                  <span class="truncate">Custom rules</span>
                </RouterLink>
              </SidebarMenuButton>
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
            class="px-2 py-2 text-xs text-base-content/50"
          >
            No projects yet. Click + to add one.
          </div>
          <SidebarMenu v-else>
            <SidebarMenuItem v-for="p in projects" :key="p.id">
              <SidebarMenuButton
                as-child
                :is-active="route.name === 'project' && route.params.id === p.id"
                :tooltip="p.name"
              >
                <RouterLink :to="{ name: 'project', params: { id: p.id } }">
                  <component :is="iconFor(p.kind)" />
                  <span class="truncate">{{ p.name }}</span>
                </RouterLink>
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
