<script setup lang="ts">
import { useRouter } from "vue-router";
import { FindingDetail } from "@/components/findings";

const props = defineProps<{
  id: string;
  scanId: string;
  ruleId: string;
}>();

const router = useRouter();

function onBack() {
  router.push({ name: "project", params: { id: props.id } });
}
</script>

<template>
  <FindingDetail
    :project-id="id"
    :scan-id="scanId"
    :rule-id="ruleId"
    @back="onBack"
  />
</template>
