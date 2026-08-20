<script setup lang="ts">
import { BookCheck, Box, HouseIcon, Moon, Settings, Sun } from "@lucide/vue";
import { RouterLink } from "vue-router";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import NavMain from "@/components/NavMain.vue";
import { useProjects } from "@/composables/useProjects";
import { useRules } from "@/composables/useRules";
import { useTheme } from "@/composables/useTheme";
import { usePageHeader } from "@/composables/usePageHeader";

const { projects } = useProjects();
const { rules } = useRules();
const { theme, toggleTheme } = useTheme();
const { nav } = usePageHeader();

const data = {
  navMain: [
    {
      title: "Home",
      icon: HouseIcon,
      to: "/",
    },
    {
      title: "Projects",
      icon: Box,
      to: "/projects",
    },
    {
      title: "Rules",
      icon: BookCheck,
      to: "/rules",
    },
  ],
}

</script>

<template>
  <Sidebar class="h-auto">
    <SidebarHeader>
      <div class="px-2 py-1">
        <div class="text-base font-semibold">Ladon</div>
        <div class="text-xs text-muted-foreground">Kubernetes security scan</div>
      </div>
    </SidebarHeader>

    <SidebarContent>
      <NavMain :items="data.navMain" /> 
    </SidebarContent>
    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton as-child :is-active="nav === 'settings'">
            <RouterLink to="/settings">
              <Settings />
              <span class="truncate">Settings</span>
            </RouterLink>
          </SidebarMenuButton>
        </SidebarMenuItem>
      </SidebarMenu>
      <div class="flex items-center justify-between gap-2 px-2">
        <span class="truncate text-xs text-muted-foreground">
          {{ projects.length }} project{{ projects.length === 1 ? "" : "s" }} ·
          {{ rules.length }} rule{{ rules.length === 1 ? "" : "s" }}
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
