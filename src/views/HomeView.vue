<script setup lang="ts">
import { useRouter } from "vue-router";
import { BookCheck, Box } from "@lucide/vue";

const router = useRouter();
</script>

<template>
  <div class="grid gap-4 md:grid-cols-2">
    <div
      class="card cursor-pointer bg-base-100 shadow transition-colors hover:bg-base-200"
      role="button"
      tabindex="0"
      @click="router.push({ name: 'projects' })"
      @keydown.enter="router.push({ name: 'projects' })"
    >
      <div class="card-body items-center gap-2 text-center">
        <Box class="size-8 text-secondary" />
        <h2 class="card-title text-base">Projects</h2>
        <p class="text-sm text-base-content/50">
          Connect a cluster or point at a container image, then run a scan.
        </p>
      </div>
    </div>
    <div
      class="card cursor-pointer bg-base-100 shadow transition-colors hover:bg-base-200"
      role="button"
      tabindex="0"
      @click="router.push({ name: 'rules' })"
      @keydown.enter="router.push({ name: 'rules' })"
    >
      <div class="card-body items-center gap-2 text-center">
        <BookCheck class="size-8 text-primary" />
        <h2 class="card-title text-base">Rules</h2>
        <p class="text-sm text-base-content/50">
          Your own matchers plus the built-in CIS-aligned checks that ship with
          Corsair.
        </p>
      </div>
    </div>
  </div>
</template>
