<script setup lang="ts">
import { RouterLink } from "vue-router";
import { BookCheck, Box, Settings } from "@lucide/vue";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2">
    <RouterLink :to="{ name: 'projects' }" class="block">
      <Card class="h-full transition-colors hover:bg-accent/50">
        <CardHeader class="items-center justify-items-center text-center">
          <Box class="size-8 text-muted-foreground" />
          <CardTitle>Projects</CardTitle>
          <CardDescription>
            Point at a cluster or container image, then run a scan.
          </CardDescription>
        </CardHeader>
      </Card>
    </RouterLink>
    <RouterLink :to="{ name: 'rules' }" class="block">
      <Card class="h-full transition-colors hover:bg-accent/50">
        <CardHeader class="items-center justify-items-center text-center">
          <BookCheck class="size-8 text-muted-foreground" />
          <CardTitle>Rules</CardTitle>
          <CardDescription>
            Matchers that flag Kubernetes resources during a scan.
          </CardDescription>
        </CardHeader>
      </Card>
    </RouterLink>
    <RouterLink :to="{ name: 'settings' }" class="block md:col-span-2">
      <Card class="h-full transition-colors hover:bg-accent/50">
        <CardHeader class="items-center justify-items-center text-center">
          <Settings class="size-8 text-muted-foreground" />
          <CardTitle>Settings</CardTitle>
          <CardDescription>
            Connect to a Kubernetes cluster through your local kubeconfig.
          </CardDescription>
        </CardHeader>
      </Card>
    </RouterLink>
  </div>
</template>
