<script setup lang="ts">
import { computed } from "vue";
import { useFormContext } from "vee-validate";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import FormKubernetesCluster from "./FormKubernetesCluster.vue";
import FormContainerImage from "./FormContainerImage.vue";
import type { ProjectFormValues } from "./types";

defineProps<{
  isSubmitting: boolean;
}>();

const emit = defineEmits<{
  submit: [values: ProjectFormValues];
}>();

const { values, setFieldValue, validate } = useFormContext<ProjectFormValues>();

const isK8s = computed(() => values.kind === "kubernetesClusterReview");

function imageToName(image: string): string {
  const lastSlash = image.lastIndexOf("/");
  const lastPart = lastSlash >= 0 ? image.slice(lastSlash + 1) : image;
  const tagIndex = lastPart.search("[:@]");
  return tagIndex >= 0 ? lastPart.slice(0, tagIndex) : lastPart;
}

async function onSubmit() {
  if (!isK8s.value && !values.name?.trim() && values.image?.trim()) {
    setFieldValue("name", imageToName(values.image));
  }
  const result = await validate();
  if (result.valid) {
    emit("submit", result.values as ProjectFormValues);
  }
}
</script>

<template>
  <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
    <FormKubernetesCluster v-if="isK8s" />
    <FormContainerImage v-else />

    <div class="mt-2 flex items-center justify-end">
      <Button type="submit" :disabled="isSubmitting">
        <Spinner v-if="isSubmitting" />
        Save
      </Button>
    </div>
  </form>
</template>
