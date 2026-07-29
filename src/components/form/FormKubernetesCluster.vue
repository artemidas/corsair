<script setup lang="ts">
import { FormField, FormItem, FormLabel, FormControl, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
</script>

<template>
  <FormField v-slot="{ componentField }" name="name">
    <FormItem>
      <FormLabel>Name</FormLabel>
      <FormControl>
        <Input
          v-bind="componentField"
          placeholder="Production cluster review"
          autocomplete="off"
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  </FormField>

  <FormField v-slot="{ componentField }" name="context">
    <FormItem>
      <FormLabel>
        Kubeconfig context
        <span class="text-xs font-normal text-muted-foreground">
          (optional)
        </span>
      </FormLabel>
      <FormControl>
        <Input
          v-bind="componentField"
          placeholder="leave empty to use the active context"
          autocomplete="off"
        />
      </FormControl>
      <FormMessage />
    </FormItem>
  </FormField>
</template>
