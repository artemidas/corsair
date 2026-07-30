<script setup lang="ts">
import { useRouter } from "vue-router";
import { BookCheck, FolderOpen, Plus } from "@lucide/vue";
import { Button } from "@/components/ui/button";

const router = useRouter();
</script>

<template>
  <div class="grid gap-4 md:grid-cols-3">
    <div
      class="card cursor-pointer bg-base-100 shadow transition-colors hover:bg-base-200"
      role="button"
      tabindex="0"
      @click="router.push('/rules/built-in')"
      @keydown.enter="router.push('/rules/built-in')"
    >
      <div class="card-body items-center gap-2 text-center">
        <BookCheck class="size-8 text-primary" />
        <h2 class="card-title text-base">Built-in rules</h2>
        <p class="text-sm text-base-content/50">
          Curated CIS-aligned checks that ship with Corsair. Read-only.
        </p>
      </div>
    </div>
    <div
      class="card cursor-pointer bg-base-100 shadow transition-colors hover:bg-base-200"
      role="button"
      tabindex="0"
      @click="router.push('/rules/custom')"
      @keydown.enter="router.push('/rules/custom')"
    >
      <div class="card-body items-center gap-2 text-center">
        <FolderOpen class="size-8 text-secondary" />
        <h2 class="card-title text-base">Custom rules</h2>
        <p class="text-sm text-base-content/50">
          Author your own simple matchers against any K8s resource.
        </p>
        <Button
          size="sm"
          variant="secondary"
          class="mt-2"
          @click.stop="router.push('/rules/custom')"
        >
          <Plus class="size-4" />
          New rule
        </Button>
      </div>
    </div>
  </div>
</template>
