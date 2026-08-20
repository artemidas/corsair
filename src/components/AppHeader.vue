<script setup lang="ts">
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import { ClusterStatusBadge } from "@/components/cluster";
import AppBreadcrumb from "@/components/AppBreadcrumb.vue";
import { usePageHeader } from "@/composables/usePageHeader";

const { title, crumbs } = usePageHeader();
</script>

<template>
  <header class="flex h-12 shrink-0 items-center gap-2 border-b border-border px-4">
    <SidebarTrigger />
    <Separator orientation="vertical" class="mx-2 data-[orientation=vertical]:h-4" />
    <h1 class="sr-only">{{ title }}</h1>
    <AppBreadcrumb class="min-w-0 flex-1" :crumbs="crumbs" />
    <div class="ml-auto">
      <ClusterStatusBadge />
    </div>
  </header>
</template>
