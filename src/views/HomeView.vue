<script setup lang="ts">
import { RouterLink } from "vue-router";
import { BookCheck, Box, ShipWheel, Settings } from "@lucide/vue";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

const links = [
  {
    name: 'cluster-projects',
    icon: ShipWheel,
    title: 'Clusters',
    description: 'Point at a Kubernetes cluster, then run a scan.',
  },
  
  {
    name: 'image-projects',
    icon: Box,
    title: 'Images',
    description: 'Point at a container image, then run a scan.',
  },
  {
    name: 'rules',
    icon: BookCheck,
    title: 'Rules',
    description: 'Matchers that flag Kubernetes resources during a scan.',
  },
  
  {
    name: 'settings',
    icon: Settings,
    title: 'Settings',
    description: 'Connect to a Kubernetes cluster through your local kubeconfig.',
  },
];
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2">
    <RouterLink v-for="link in links" :to="{ name: link.name }" class="block">
      <Card class="h-full transition-colors hover:bg-accent/50">
        <CardHeader class="items-center justify-items-center text-center">
          <component :is="link.icon" class="size-8 text-muted-foreground" />
          <CardTitle>{{ link.title }}</CardTitle>
          <CardDescription>
            {{ link.description }}
          </CardDescription>
        </CardHeader>
      </Card>
    </RouterLink>
  </div>
</template>
