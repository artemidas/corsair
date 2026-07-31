<script setup lang="ts">
import { ProjectEditor } from "@/components/project";
import { useRouter } from "vue-router";

const router = useRouter();

function onCreated(id: string) {
  router.replace({ name: "project", params: { id } });
}
</script>

<template>
  <ProjectEditor
    :project="null"
    @created="onCreated"
    @close="router.push({ name: 'projects' })"
  />
</template>
