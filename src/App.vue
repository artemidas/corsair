<script setup lang="ts">
import { onMounted, computed } from "vue";
import { RouterView, useRoute } from "vue-router";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/AppSidebar.vue";
import { useProjects } from "@/composables/useProjects";
import { useCustomRules } from "@/composables/useCustomRules";

const { loadProjects, refreshConnection } = useProjects();
const { loadRules } = useCustomRules();

onMounted(async () => {
  await Promise.all([loadProjects(), loadRules(), refreshConnection()]);
});

const route = useRoute();
const title = computed(() => route.meta?.title);
</script>

<template>
  <SidebarProvider>
    <AppSidebar />
    <SidebarInset>
      <header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
        <SidebarTrigger />
        <h1 class="text-lg font-semibold">{{ title }}</h1>
      </header>
      <main class="flex-1 overflow-y-auto p-6">
        <div class="flex max-w-5xl flex-col gap-4">
          <RouterView />
        </div>
      </main>
    </SidebarInset>
  </SidebarProvider>
</template>
