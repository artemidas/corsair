<script setup lang="ts">
import { computed } from "vue";
import { useRouter } from "vue-router";
import { Box } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ProjectEditor } from "@/components/project";
import { useProjects, type Project } from "@/composables/useProjects";

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const { projects, loading } = useProjects();

const project = computed<Project | null>(
  () => projects.value.find((p) => p.id === props.id) ?? null,
);

function back() {
  router.push({ name: "project", params: { id: props.id } });
}
</script>

<template>
  <ProjectEditor v-if="project" :project="project" @close="back" />

  <div v-else-if="loading" class="text-sm text-muted-foreground">Loading…</div>

  <Empty v-else>
    <EmptyHeader>
      <EmptyMedia variant="icon">
        <Box />
      </EmptyMedia>
      <EmptyTitle>Project not found</EmptyTitle>
      <EmptyDescription>
        The project
        <span class="font-mono">{{ id }}</span>
        doesn't exist anymore.
      </EmptyDescription>
    </EmptyHeader>
    <EmptyContent>
      <Button variant="ghost" size="sm" @click="router.push({ name: 'projects' })">
        Back to projects
      </Button>
    </EmptyContent>
  </Empty>
</template>
