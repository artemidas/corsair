<script setup lang="ts">
import { onMounted, onUnmounted, computed } from "vue";
import { RouterView, useRoute } from "vue-router";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { ClusterStatusBadge } from "@/components/cluster";
import { useProjects } from "@/composables/useProjects";
import { useCustomRules } from "@/composables/useCustomRules";
import { useCluster } from "@/composables/useCluster";

const { loadProjects } = useProjects();
const { loadRules } = useCustomRules();
const { startPolling, stopPolling } = useCluster();

onMounted(async () => {
  startPolling();
  await Promise.all([loadProjects(), loadRules()]);
});

onUnmounted(() => {
  stopPolling();
});

const route = useRoute();
const title = computed(() => route.meta?.title);
</script>

<template>
  <SidebarProvider class="h-full min-h-0 overflow-hidden">
    <AppSidebar />
    <SidebarInset class="min-h-0 overflow-hidden">
      <header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
        <SidebarTrigger />
        <h1 class="min-w-0 truncate text-lg font-semibold">{{ title }}</h1>
        <div class="ml-auto">
          <ClusterStatusBadge />
        </div>
      </header>
      <main class="min-h-0 flex-1 overflow-y-auto p-6">
        <div class="flex max-w-5xl flex-col gap-4">
          <RouterView />
        </div>
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
