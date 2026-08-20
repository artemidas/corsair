<script setup lang="ts">
import { shallowRef } from "vue";
import { useRouter } from "vue-router";
import { ArrowLeft, CircleAlert, Hexagon } from "@lucide/vue";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { FormKubernetesCluster } from "@/components/form";
import { Spinner } from "@/components/ui/spinner";
import { useProjects } from "@/composables/useProjects";

const router = useRouter();
const { createProject } = useProjects();

const isSubmitting = shallowRef(false);
const submitError = shallowRef<string | null>(null);

const form = useForm({
  validationSchema: toTypedSchema(
    z.object({
      name: z.string().trim().min(1, "Name is required."),
      context: z.string(),
    }),
  ),
  initialValues: {
    name: "",
    context: "",
  },
});

async function onSubmit() {
  const result = await form.validate();
  const values = result.values;
  if (!result.valid || !values?.name) return;

  isSubmitting.value = true;
  submitError.value = null;
  try {
    const created = await createProject({
      name: values.name.trim(),
      kind: "kubernetesClusterReview",
      config: { context: values.context?.trim() || null, image: null },
    });
    await router.replace({ name: "project", params: { id: created.id } });
  } catch (err) {
    console.error("Failed to save project:", err);
    submitError.value = err instanceof Error ? err.message : String(err);
  } finally {
    isSubmitting.value = false;
  }
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <Alert v-if="submitError" variant="destructive">
      <CircleAlert />
      <AlertTitle>Could not save project</AlertTitle>
      <AlertDescription>{{ submitError }}</AlertDescription>
    </Alert>

    <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
      <div class="flex items-center gap-2 text-sm text-muted-foreground">
        <Hexagon class="size-4" />
        <span>Kubernetes cluster review</span>
      </div>

      <FormKubernetesCluster />

      <div class="mt-2 flex items-center justify-between">
        <Button
          type="button"
          variant="ghost"
          @click="router.push({ name: 'new-project' })"
        >
          <ArrowLeft />
          Back
        </Button>
        <Button type="submit" :disabled="isSubmitting">
          <Spinner v-if="isSubmitting" />
          Create
        </Button>
      </div>
    </form>
  </div>
</template>
