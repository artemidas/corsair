<script setup lang="ts">
import { onMounted, onUnmounted } from "vue";
import { RouterView } from "vue-router";
import { SidebarProvider, SidebarInset } from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import AppHeader from "@/components/AppHeader.vue";
import { useProjects } from "@/composables/useProjects";
import { useRules } from "@/composables/useRules";
import { useCluster } from "@/composables/useCluster";

const { loadProjects } = useProjects();
const { loadRules } = useRules();
const { startPolling, stopPolling } = useCluster();

onMounted(async () => {
  startPolling();
  await Promise.all([loadProjects(), loadRules()]);
});

onUnmounted(() => {
  stopPolling();
});
</script>

<template>
  <SidebarProvider class="h-full min-h-0 overflow-hidden">
    <AppSidebar variant="inset" />
    <SidebarInset class="min-h-0 overflow-hidden">
      <AppHeader />
      <main class="min-h-0 flex-1 overflow-y-auto p-6">
        <div class="flex flex-col gap-4">
          <RouterView />
        </div>
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
