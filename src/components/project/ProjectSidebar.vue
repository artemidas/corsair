<script setup lang="ts">
import { useProjects, type Project } from "@/composables/useProjects";

const emit = defineEmits<{
  new: [];
  edit: [project: Project];
}>();

const { projects, selectedProjectId, selectProject, deleteProject, loading, loadError } =
  useProjects();

function onDelete(p: Project, ev: Event) {
  ev.stopPropagation();
  if (!confirm(`Delete project "${p.name}"?`)) return;
  deleteProject(p.id).catch((err) => alert(String(err)));
}

function kindLabel(kind: Project["kind"]): string {
  return kind === "kubernetesClusterReview" ? "K8s" : "Image";
}

function kindBadgeClass(kind: Project["kind"]): string {
  return kind === "kubernetesClusterReview" ? "badge-primary" : "badge-secondary";
}

function subtitle(p: Project): string {
  if (p.kind === "kubernetesClusterReview") {
    return p.config.context ? `context: ${p.config.context}` : "active context";
  }
  return `image: ${p.config.image ?? ""}`;
}
</script>

<template>
  <aside class="flex h-full w-72 flex-col border-r border-base-300 bg-base-100">
    <div class="flex items-center justify-between border-b border-base-300 px-4 py-3">
      <div>
        <div class="text-sm font-semibold">Projects</div>
        <div class="text-xs text-base-content/50">{{ projects.length }} total</div>
      </div>
      <button class="btn btn-sm btn-primary" @click="emit('new')">
        <span class="text-lg leading-none">+</span>
        New
      </button>
    </div>

    <div v-if="loadError" class="m-3 alert alert-error text-xs">
      {{ loadError }}
    </div>

    <div
      v-if="loading && projects.length === 0"
      class="flex items-center gap-2 px-4 py-6 text-sm text-base-content/50"
    >
      <span class="loading loading-spinner loading-sm"></span> Loading…
    </div>

    <div
      v-else-if="projects.length === 0"
      class="px-4 py-6 text-sm text-base-content/50"
    >
      No projects yet. Click <span class="font-medium">+ New</span> to create one.
    </div>

    <ul v-else class="menu menu-sm w-full flex-1 overflow-y-auto p-2">
      <li v-for="p in projects" :key="p.id">
        <a
          class="group flex items-start gap-2 py-2"
          :class="{ active: selectedProjectId === p.id }"
          @click="selectProject(p.id)"
        >
          <span class="badge badge-sm shrink-0" :class="kindBadgeClass(p.kind)">
            {{ kindLabel(p.kind) }}
          </span>
          <div class="min-w-0 flex-1">
            <div class="truncate font-medium">{{ p.name }}</div>
            <div class="truncate text-xs text-base-content/50">
              {{ subtitle(p) }}
            </div>
          </div>
          <div class="flex shrink-0 items-center gap-1 opacity-0 group-hover:opacity-100">
            <button
              class="btn btn-ghost btn-xs"
              title="Edit"
              @click.stop="emit('edit', p)"
            >
              Edit
            </button>
            <button
              class="btn btn-ghost btn-xs text-error"
              title="Delete"
              @click="onDelete(p, $event)"
            >
              ✕
            </button>
          </div>
        </a>
      </li>
    </ul>
  </aside>
</template>
