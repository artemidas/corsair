<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  useProjects,
  type Project,
  type ProjectInput,
  type ProjectKind,
} from "@/composables/useProjects";

const props = defineProps<{
  project: Project | null;
}>();

const emit = defineEmits<{
  close: [];
}>();

const dialogRef = ref<HTMLDialogElement | null>(null);

const { createProject, updateProject, selectProject } = useProjects();

const name = ref("");
const kind = ref<ProjectKind>("kubernetesClusterReview");
const context = ref("");
const image = ref("");

const isEdit = computed(() => props.project !== null);
const title = computed(() => (isEdit.value ? "Edit project" : "New project"));

const error = ref("");
const saving = ref(false);

function reset() {
  error.value = "";
  saving.value = false;
  if (props.project) {
    name.value = props.project.name;
    kind.value = props.project.kind;
    context.value = props.project.config.context ?? "";
    image.value = props.project.config.image ?? "";
  } else {
    name.value = "";
    kind.value = "kubernetesClusterReview";
    context.value = "";
    image.value = "";
  }
}

watch(
  () => props.project,
  () => reset(),
  { immediate: true },
);

onMounted(() => {
  dialogRef.value?.showModal();
});

function close() {
  dialogRef.value?.close();
  emit("close");
}

function buildInput(): ProjectInput {
  if (kind.value === "kubernetesClusterReview") {
    const trimmed = context.value.trim();
    return {
      name: name.value.trim(),
      kind: "kubernetesClusterReview",
      config: { context: trimmed || null, image: null },
    };
  }
  return {
    name: name.value.trim(),
    kind: "containerImageReview",
    config: { context: null, image: image.value.trim() },
  };
}

async function onSubmit(ev: Event) {
  ev.preventDefault();
  error.value = "";
  if (name.value.trim() === "") {
    error.value = "Name is required.";
    return;
  }
  if (kind.value === "containerImageReview" && image.value.trim() === "") {
    error.value = "Image is required.";
    return;
  }

  saving.value = true;
  try {
    if (isEdit.value && props.project) {
      await updateProject(props.project.id, buildInput());
    } else {
      const created = await createProject(buildInput());
      selectProject(created.id);
    }
    close();
  } catch (err) {
    error.value = String(err);
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <dialog ref="dialogRef" class="modal" @close="emit('close')">
    <div class="modal-box max-w-lg">
      <h3 class="text-lg font-bold">{{ title }}</h3>

      <form class="mt-4 flex flex-col gap-4" @submit="onSubmit">
        <label class="form-control">
          <div class="label">
            <span class="label-text">Name</span>
          </div>
          <input
            v-model="name"
            type="text"
            class="input input-bordered"
            placeholder="Production cluster review"
            required
          />
        </label>

        <div class="form-control">
          <div class="label">
            <span class="label-text">Kind</span>
          </div>
          <div class="flex gap-2">
            <label
              class="flex flex-1 cursor-pointer items-center gap-2 rounded-lg border p-3"
              :class="kind === 'kubernetesClusterReview' ? 'border-primary bg-primary/10' : 'border-base-300'"
            >
              <input
                v-model="kind"
                type="radio"
                value="kubernetesClusterReview"
                class="radio radio-sm radio-primary"
              />
              <div>
                <div class="text-sm font-medium">Kubernetes cluster review</div>
                <div class="text-xs text-base-content/50">
                  Scan a live cluster for misconfigurations.
                </div>
              </div>
            </label>
            <label
              class="flex flex-1 cursor-pointer items-center gap-2 rounded-lg border p-3"
              :class="kind === 'containerImageReview' ? 'border-secondary bg-secondary/10' : 'border-base-300'"
            >
              <input
                v-model="kind"
                type="radio"
                value="containerImageReview"
                class="radio radio-sm radio-secondary"
              />
              <div>
                <div class="text-sm font-medium">Container image review</div>
                <div class="text-xs text-base-content/50">
                  Inspect a container image for risks.
                </div>
              </div>
            </label>
          </div>
        </div>

        <label
          v-if="kind === 'kubernetesClusterReview'"
          class="form-control"
        >
          <div class="label">
            <span class="label-text">Kubeconfig context</span>
            <span class="label-text-alt text-base-content/50">optional</span>
          </div>
          <input
            v-model="context"
            type="text"
            class="input input-bordered"
            placeholder="leave empty to use the active context"
          />
        </label>

        <label v-else class="form-control">
          <div class="label">
            <span class="label-text">Image reference</span>
          </div>
          <input
            v-model="image"
            type="text"
            class="input input-bordered"
            placeholder="nginx:1.27"
            required
          />
        </label>

        <div v-if="error" class="alert alert-error text-sm">{{ error }}</div>

        <div class="modal-action mt-2">
          <button
            type="button"
            class="btn btn-ghost"
            :disabled="saving"
            @click="close"
          >
            Cancel
          </button>
          <button type="submit" class="btn btn-primary" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            {{ isEdit ? "Save" : "Create" }}
          </button>
        </div>
      </form>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button>close</button>
    </form>
  </dialog>
</template>
