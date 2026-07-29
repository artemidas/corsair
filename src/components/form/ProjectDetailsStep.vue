<script setup lang="ts">
import { computed } from "vue";
import { useFormContext } from "vee-validate";
import { ArrowLeft, Box, Hexagon } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import FormKubernetesCluster from "./FormKubernetesCluster.vue";
import FormContainerImage from "./FormContainerImage.vue";
import type { ProjectFormValues } from "./types";

defineProps<{
  isEdit: boolean;
  isSubmitting: boolean;
}>();

const emit = defineEmits<{
  submit: [values: ProjectFormValues];
  cancel: [];
  back: [];
}>();

const { values, setFieldValue, validate } = useFormContext<ProjectFormValues>();

const isK8s = computed(() => values.kind === "kubernetesClusterReview");

// Image projects don't have a user-typed name; derive one from the image
// reference (e.g. `nginx:1.27` -> `nginx`, `ghcr.io/org/app:latest` -> `app`).
function imageToName(image: string): string {
  const lastSlash = image.lastIndexOf("/");
  const lastPart = lastSlash >= 0 ? image.slice(lastSlash + 1) : image;
  const tagIndex = lastPart.search("[:@]");
  return tagIndex >= 0 ? lastPart.slice(0, tagIndex) : lastPart;
}

async function onSubmit() {
  if (
    isK8s &&
    !values.name?.trim() &&
    values.image?.trim()
  ) {
    setFieldValue("name", imageToName(values.image));
  }
  const result = await validate();
  if (result.valid) {
    emit("submit", result.values as ProjectFormValues);
  }
}
</script>

<template>
  <div class="mt-4 flex flex-col gap-4">

    <form @submit.prevent="onSubmit">
      <div  
        v-if="!isEdit"
        class="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <component :is="isK8s ? Hexagon : Box" class="size-4" />
        <span>
          {{ isK8s ? "Kubernetes cluster review" : "Container image review" }}
        </span>
      </div>

      <FormKubernetesCluster v-if="isK8s" />
      <FormContainerImage v-else />

      <div class="modal-action mt-2">
        <Button
          v-if="!isEdit"
          type="button"
          class="self-start -ml-3"
          @click="emit('back')"
        >
          <ArrowLeft class="size-4" />
          Back
        </Button>
        <Button type="submit" :disabled="isSubmitting">
          <span
            v-if="isSubmitting"
            class="loading loading-spinner loading-sm"
          ></span>
          {{ isEdit ? "Save" : "Create" }}
        </Button>
      </div>
    </form>
  </div>
</template>
