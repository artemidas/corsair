<script setup lang="ts">
import { BookCheck, Box, HouseIcon, Moon, Settings, Sun } from "@lucide/vue";
import { RouterLink, useRoute } from "vue-router";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { useProjects } from "@/composables/useProjects";
import { useCustomRules } from "@/composables/useCustomRules";
import { useTheme } from "@/composables/useTheme";

const { projects } = useProjects();
const { userRules } = useCustomRules();
const { theme, toggleTheme } = useTheme();

const route = useRoute();
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

        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as-child :is-active="route.name === 'home'">
              <RouterLink to="/">
                <HouseIcon />
                <span class="truncate">Home</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>

        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as-child :is-active="route.name === 'projects'">
              <RouterLink to="/projects">
                <Box />
                <span class="truncate">Projects</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>

        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton as-child :is-active="route.name === 'rules'">
              <RouterLink to="/rules">
                <BookCheck />
                <span class="truncate">Rules</span>
              </RouterLink>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter>
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton as-child :is-active="route.name === 'settings'">
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
