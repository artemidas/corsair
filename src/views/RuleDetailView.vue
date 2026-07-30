<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useCustomRules, type CustomRule } from "@/composables/useCustomRules";
import { RuleDetail, RuleEditor } from "@/components/rule";

const route = useRoute();
const router = useRouter();
const { getRuleById } = useCustomRules();

const source = computed<"built-in" | "custom">(
  () => (route.meta.source as "built-in" | "custom") ?? "custom",
);

const rule = computed<CustomRule | null>(
  () => getRuleById(route.params.id as string),
);

const editorOpen = ref(false);

function openEdit() {
  editorOpen.value = true;
}

function onBack() {
  router.push({ name: source.value === "built-in" ? "built-in-rules" : "custom-rules" });
}

// Re-reset the editor target when the route changes so the modal opens
// for the right rule.
const editorTarget = ref<CustomRule | null>(null);
watch(
  () => rule.value?.id,
  () => {
    editorTarget.value = rule.value;
  },
  { immediate: true },
);
</script>

<template>
  <div>
    <RuleDetail
      v-if="rule"
      :rule="rule"
      :selected-project="null"
      @edit="openEdit"
      @back="onBack"
    />
    <div v-else class="card bg-base-100 shadow">
      <div class="card-body items-center text-center text-base-content/50">
        <h2 class="card-title text-base-content">Rule not found</h2>
        <p>
          The rule
          <span class="font-mono">{{ route.params.id }}</span>
          doesn't exist anymore.
        </p>
        <button class="btn btn-sm btn-ghost mt-2" @click="onBack">Back to rules</button>
      </div>
    </div>

    <RuleEditor
      v-if="editorOpen && editorTarget"
      :rule="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
