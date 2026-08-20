<script setup lang="ts">
import { RouterView } from "vue-router";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>New project</CardTitle>
    </CardHeader>
    <CardContent>
      <RouterView />
    </CardContent>
  </Card>
</template>
