<script setup lang="ts">
import { shallowRef, watch } from "vue";
import { CircleAlert } from "@lucide/vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  imagesFor,
  useProjects,
  type Project,
  type ProjectInput,
} from "@/composables/useProjects";
import { ProjectDetailsStep } from "@/components/form";
import type { ProjectFormValues } from "@/components/form";

const props = defineProps<{
  project: Project;
}>();

const emit = defineEmits<{
  close: [];
}>();

const { updateProject } = useProjects();

const isSubmitting = shallowRef(false);
const submitError = shallowRef<string | null>(null);

const form = useForm<ProjectFormValues>({
  validationSchema: toTypedSchema(
    z
      .object({
        name: z.string().trim().min(1, "Name is required."),
        kind: z.enum(["kubernetesClusterReview", "containerImageReview"]),
        context: z.string(),
        images: z.array(z.string()),
      })
      .refine(
        (data) =>
          data.kind !== "containerImageReview" ||
          data.images.some((image) => image.trim().length > 0),
        { path: ["images"], message: "Select at least one image." },
      ),
  ),
  initialValues: {
    name: "",
    kind: "kubernetesClusterReview",
    context: "",
    images: [] as string[],
  },
});

async function submitProject(values: ProjectFormValues) {
  isSubmitting.value = true;
  submitError.value = null;
  try {
    const input: ProjectInput = {
      name: values.name.trim(),
      kind: values.kind,
      config:
        values.kind === "kubernetesClusterReview"
          ? { context: values.context.trim() || null, images: [] }
          : {
              context: null,
              images: values.images.map((image) => image.trim()).filter(Boolean),
            },
    };
    await updateProject(props.project.id, input);
    emit("close");
  } catch (err) {
    console.error("Failed to save project:", err);
    submitError.value = err instanceof Error ? err.message : String(err);
  } finally {
    isSubmitting.value = false;
  }
}

function reset() {
  form.resetForm({
    values: {
      name: props.project.name,
      kind: props.project.kind,
      context: props.project.config.context ?? "",
      images: imagesFor(props.project.config),
    },
  });
}

watch(
  () => props.project,
  () => reset(),
  { immediate: true },
);
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Edit project</CardTitle>
    </CardHeader>
    <CardContent class="flex flex-col gap-4">
      <Alert v-if="submitError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Could not save project</AlertTitle>
        <AlertDescription>{{ submitError }}</AlertDescription>
      </Alert>

      <ProjectDetailsStep
        :is-submitting="isSubmitting"
        @submit="submitProject"
      />
    </CardContent>
  </Card>
</template>
