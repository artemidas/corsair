<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ProjectDetail, ProjectEditor } from "@/components/project";
import { useProjects, type Project } from "@/composables/useProjects";
import { ref } from "vue";

const route = useRoute();
const router = useRouter();
const { projects } = useProjects();

const project = computed<Project | null>(
  () => projects.value.find((p) => p.id === route.params.id) ?? null,
);

const editorOpen = ref(false);
const editorTarget = ref<Project | null>(null);

function openEdit(p: Project) {
  editorTarget.value = p;
  editorOpen.value = true;
}
</script>

<template>
  <div>
    <ProjectDetail
      v-if="project"
      :project="project"
      @edit="openEdit"
    />
    <div v-else class="card bg-base-100 shadow">
      <div class="card-body items-center text-center text-base-content/50">
        <h2 class="card-title text-base-content">Project not found</h2>
        <p>
          The project
          <span class="font-mono">{{ route.params.id }}</span>
          doesn't exist anymore.
        </p>
        <button class="btn btn-sm btn-ghost mt-2" @click="router.push('/')">
          Back home
        </button>
      </div>
    </div>

    <ProjectEditor
      v-if="editorOpen && project"
      :project="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
