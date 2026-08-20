<script setup lang="ts">
import { computed } from "vue";
import { RouterView, useRoute } from "vue-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const route = useRoute();
const title = computed(() => {
  switch (route.name) {
    case "new-kubernetes-project":
      return "New cluster";
    case "new-container-image-project":
      return "New engagement";
    default:
      return "New project";
  }
});
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>{{ title }}</CardTitle>
    </CardHeader>
    <CardContent>
      <RouterView />
    </CardContent>
  </Card>
</template>
