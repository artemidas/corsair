<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
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

interface FormValues {
  name: string;
  kind: ProjectKind;
  context: string;
  image: string;
}

const dialogRef = ref<HTMLDialogElement | null>(null);
const { createProject, updateProject, selectProject } = useProjects();

const isEdit = computed(() => props.project !== null);
const title = computed(() => (isEdit.value ? "Edit project" : "New project"));
const isSubmitting = ref(false);

const form = useForm({
  validationSchema: toTypedSchema(
    z
      .object({
        name: z.string().trim().min(1, "Name is required."),
        kind: z.enum(["kubernetesClusterReview", "containerImageReview"]),
        context: z.string(),
        image: z.string(),
      })
      .refine(
        (data) =>
          data.kind !== "containerImageReview" || data.image.trim().length > 0,
        { path: ["image"], message: "Image is required." },
      ),
  ),
  initialValues: {
    name: "",
    kind: "kubernetesClusterReview",
    context: "",
    image: "",
  },
});

async function onSubmit(values: FormValues) {
  isSubmitting.value = true;
  try {
    const input: ProjectInput = {
      name: values.name.trim(),
      kind: values.kind,
      config:
        values.kind === "kubernetesClusterReview"
          ? { context: values.context.trim() || null, image: null }
          : { context: null, image: values.image.trim() },
    };
    if (isEdit.value && props.project) {
      await updateProject(props.project.id, input);
    } else {
      const created = await createProject(input);
      selectProject(created.id);
    }
    close();
  } catch (err) {
    form.setFieldError("name", String(err));
  } finally {
    isSubmitting.value = false;
  }
}

async function handleSubmit() {
  const result = await form.validate();
  if (result.valid) {
    await onSubmit(result.values as FormValues);
  }
}

function reset() {
  if (props.project) {
    form.resetForm({
      values: {
        name: props.project.name,
        kind: props.project.kind,
        context: props.project.config.context ?? "",
        image: props.project.config.image ?? "",
      },
    });
  } else {
    form.resetForm({
      values: {
        name: "",
        kind: "kubernetesClusterReview",
        context: "",
        image: "",
      },
    });
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
</script>

<template>
  <dialog ref="dialogRef" class="modal" @close="emit('close')">
    <div class="modal-box max-w-lg">
      <h3 class="text-lg font-bold">{{ title }}</h3>

      <form class="mt-4 flex flex-col gap-4" @submit.prevent="handleSubmit">
        <FormField v-slot="{ field }" name="name">
          <FormItem>
            <FormLabel>Name</FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="Production cluster review"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ field }" name="kind">
          <FormItem>
            <FormLabel>Kind</FormLabel>
            <FormControl>
              <div class="flex gap-2">
                <label
                  class="flex flex-1 cursor-pointer items-start gap-2 rounded-md border p-3"
                  :class="field.value === 'kubernetesClusterReview' ? 'border-primary bg-primary/10' : 'border-input'"
                >
                  <input
                    type="radio"
                    value="kubernetesClusterReview"
                    :checked="field.value === 'kubernetesClusterReview'"
                    class="mt-1 h-4 w-4 accent-primary"
                    @change="(e) => field.onChange((e.target as HTMLInputElement).value)"
                  />
                  <div>
                    <div class="text-sm font-medium">Kubernetes cluster review</div>
                    <div class="text-xs text-muted-foreground">
                      Scan a live cluster for misconfigurations.
                    </div>
                  </div>
                </label>
                <label
                  class="flex flex-1 cursor-pointer items-start gap-2 rounded-md border p-3"
                  :class="field.value === 'containerImageReview' ? 'border-secondary bg-secondary/10' : 'border-input'"
                >
                  <input
                    type="radio"
                    value="containerImageReview"
                    :checked="field.value === 'containerImageReview'"
                    class="mt-1 h-4 w-4 accent-secondary"
                    @change="(e) => field.onChange((e.target as HTMLInputElement).value)"
                  />
                  <div>
                    <div class="text-sm font-medium">Container image review</div>
                    <div class="text-xs text-muted-foreground">
                      Inspect a container image for risks.
                    </div>
                  </div>
                </label>
              </div>
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField
          v-if="form.values.kind === 'kubernetesClusterReview'"
          v-slot="{ field }"
          name="context"
        >
          <FormItem>
            <FormLabel>
              Kubeconfig context
              <span class="text-xs font-normal text-muted-foreground">(optional)</span>
            </FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="leave empty to use the active context"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-else v-slot="{ field }" name="image">
          <FormItem>
            <FormLabel>Image reference</FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="nginx:1.27"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <div class="modal-action mt-2">
          <Button
            type="button"
            variant="ghost"
            :disabled="isSubmitting"
            @click="close"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="loading loading-spinner loading-sm"></span>
            {{ isEdit ? "Save" : "Create" }}
          </Button>
        </div>
      </form>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button>close</button>
    </form>
  </dialog>
</template>
