<script setup lang="ts">
import { RouterLink } from "vue-router";
import { BookCheck, Box } from "@lucide/vue";
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
            Connect a cluster or point at a container image, then run a scan.
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
            Your own matchers plus the built-in CIS-aligned checks that ship with
            Corsair.
          </CardDescription>
        </CardHeader>
      </Card>
    </RouterLink>
  </div>
</template>
