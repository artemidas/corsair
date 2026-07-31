<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import {
  Stepper,
  StepperDescription,
  StepperIndicator,
  StepperItem,
  StepperSeparator,
  StepperTitle,
  StepperTrigger,
} from "@/components/ui/stepper";
import {
  useProjects,
  type Project,
  type ProjectInput,
} from "@/composables/useProjects";
import { ProjectKindStep, ProjectDetailsStep } from "@/components/form";
import type { ProjectFormValues } from "@/components/form";

const props = defineProps<{
  project: Project | null;
}>();

const emit = defineEmits<{
  close: [];
  created: [id: string];
}>();

const { createProject, updateProject } = useProjects();

const isEdit = computed(() => props.project !== null);
const title = computed(() => (isEdit.value ? "Edit project" : "New project"));
const isSubmitting = ref(false);
const submitError = ref<string | null>(null);
const currentStep = ref(1);

const form = useForm<ProjectFormValues>({
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

async function submitProject(values: ProjectFormValues) {
  isSubmitting.value = true;
  submitError.value = null;
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
      emit("close");
    } else {
      const created = await createProject(input);
      emit("created", created.id);
    }
  } catch (err) {
    console.error("Failed to save project:", err);
    submitError.value = err instanceof Error ? err.message : String(err);
  } finally {
    isSubmitting.value = false;
  }
}

function selectKind(kind: ProjectFormValues["kind"]) {
  form.setFieldValue("kind", kind);
  currentStep.value = 2;
}

function goBack() {
  if (isEdit.value) return;
  currentStep.value = 1;
}

function reset() {
  if (props.project) {
    currentStep.value = 2;
    form.resetForm({
      values: {
        name: props.project.name,
        kind: props.project.kind,
        context: props.project.config.context ?? "",
        image: props.project.config.image ?? "",
      },
    });
  } else {
    currentStep.value = 1;
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

</script>

<template>
  <div>
    <h3 class="text-lg font-bold">{{ title }}</h3>
    <div class="flex flex-col items-center">  
      <Stepper v-if="!isEdit" v-model="currentStep" class="flex w-10/12 items-start gap-2">
        <StepperItem :step="1" class="relative flex w-full flex-col items-center justify-center">
          <StepperTrigger>
            <StepperIndicator class="bg-muted">1</StepperIndicator>
            <div class="flex flex-col items-center">
              <StepperTitle>Kind</StepperTitle>
              <StepperDescription>Pick a project type</StepperDescription>
            </div>
          </StepperTrigger>
          <StepperSeparator/>
        </StepperItem>
        <StepperItem :step="2" class="relative flex w-full flex-col items-center justify-center">
          <StepperTrigger>
            <StepperIndicator class="bg-muted">2</StepperIndicator>
            <div class="flex flex-col items-center">
              <StepperTitle>Details</StepperTitle>
              <StepperDescription>Name and config</StepperDescription>
            </div>
          </StepperTrigger>
        </StepperItem>
      </Stepper>
    </div>
    <div
      v-if="submitError"
      class="alert alert-error mt-4 text-sm"
      role="alert"
    >
      <span>{{ submitError }}</span>
    </div>

    <ProjectKindStep v-if="currentStep === 1" @select="selectKind" />
    <ProjectDetailsStep
      v-else
      :is-edit="isEdit"
      :is-submitting="isSubmitting"
      @submit="submitProject"
      @cancel="emit('close')"
      @back="goBack"
    />
  </div>
</template>
