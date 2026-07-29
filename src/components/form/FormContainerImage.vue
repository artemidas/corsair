<script setup lang="ts">
import { FormField, FormItem, FormLabel, FormControl, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
</script>

<template>
  <FormField v-slot="{ componentField }" name="image">
    <FormItem>
      <FormLabel>Image reference</FormLabel>
      <FormControl>
        <Input
          v-bind="componentField"
          placeholder="nginx:1.27"
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  </FormField>
</template>
