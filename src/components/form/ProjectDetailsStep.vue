<script setup lang="ts">
import { computed } from "vue";
import { useFormContext } from "vee-validate";
import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import FormKubernetesCluster from "./FormKubernetesCluster.vue";
import FormImageEngagement from "./FormImageEngagement.vue";
import type { ProjectFormValues } from "./types";

defineProps<{
  isSubmitting: boolean;
}>();

const emit = defineEmits<{
  submit: [values: ProjectFormValues];
}>();

const { values, validate } = useFormContext<ProjectFormValues>();

const isK8s = computed(() => values.kind === "kubernetesClusterReview");

async function onSubmit() {
  const result = await validate();
  if (result.valid) {
    emit("submit", result.values as ProjectFormValues);
  }
}
</script>

<template>
  <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
    <FormKubernetesCluster v-if="isK8s" />
    <FormImageEngagement v-else />

    <div class="mt-2 flex items-center justify-end">
      <Button type="submit" :disabled="isSubmitting">
        <Spinner v-if="isSubmitting" />
        Save
      </Button>
    </div>
  </form>
</template>
