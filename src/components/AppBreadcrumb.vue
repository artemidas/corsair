<script setup lang="ts">
import { RouterLink } from "vue-router";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import type { BreadcrumbCrumb } from "@/composables/usePageHeader";

defineProps<{
  crumbs: BreadcrumbCrumb[];
}>();
</script>

<template>
  <Breadcrumb class="min-w-0">
    <BreadcrumbList class="flex-nowrap">
      <template v-for="(crumb, index) in crumbs" :key="`${index}-${crumb.label}`">
        <BreadcrumbItem class="min-w-0">
          <BreadcrumbLink v-if="crumb.to" as-child>
            <RouterLink :to="crumb.to">{{ crumb.label }}</RouterLink>
          </BreadcrumbLink>
          <BreadcrumbPage v-else class="truncate">
            {{ crumb.label }}
          </BreadcrumbPage>
        </BreadcrumbItem>
        <BreadcrumbSeparator v-if="index < crumbs.length - 1" />
      </template>
    </BreadcrumbList>
  </Breadcrumb>
</template>
