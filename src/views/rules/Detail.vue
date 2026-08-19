<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { BookCheck } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { useCustomRules, type CustomRule } from "@/composables/useCustomRules";
import { RuleDetail, RuleEditor } from "@/components/rule";

const props = defineProps<{
  id: string;
}>();

const router = useRouter();
const { getRuleById, loading } = useCustomRules();

const rule = computed<CustomRule | null>(() => getRuleById(props.id));

const editorOpen = ref(false);

function openEdit() {
  editorOpen.value = true;
}

function onBack() {
  router.push({ name: "rules" });
}

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
      @edit="openEdit"
      @back="onBack"
    />
    <div v-else-if="loading" class="text-sm text-muted-foreground">Loading…</div>
    <Empty v-else>
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <BookCheck />
        </EmptyMedia>
        <EmptyTitle>Rule not found</EmptyTitle>
        <EmptyDescription>
          The rule
          <span class="font-mono">{{ id }}</span>
          doesn't exist anymore.
        </EmptyDescription>
      </EmptyHeader>
      <EmptyContent>
        <Button variant="ghost" size="sm" @click="onBack">Back to rules</Button>
      </EmptyContent>
    </Empty>

    <RuleEditor
      v-if="editorOpen && editorTarget"
      :rule="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
