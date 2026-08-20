<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import { useFormContext } from "vee-validate";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import LocalImagePicker from "@/components/project/LocalImagePicker.vue";
import { useLocalImages } from "@/composables/useLocalImages";

const { values, setFieldValue } = useFormContext<{
  name: string;
  images: string[];
}>();
const { images, runtime, loading, loadError, loadImages } = useLocalImages();
const customImage = shallowRef("");

const selected = computed({
  get: () => values.images ?? [],
  set: (next: string[]) => {
    void setFieldValue("images", next, false);
  },
});

onMounted(() => {
  void loadImages();
});

function addCustom() {
  const reference = customImage.value.trim();
  if (!reference) return;
  if (!selected.value.includes(reference)) {
    selected.value = [...selected.value, reference];
  }
  customImage.value = "";
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <FormField v-slot="{ componentField }" name="name">
      <FormItem>
        <FormLabel>Name</FormLabel>
        <FormControl>
          <Input
            v-bind="componentField"
            placeholder="Q3 image review"
            autocomplete="off"
          />
        </FormControl>
        <FormMessage />
      </FormItem>
    </FormField>

    <FormField name="images">
      <FormItem>
        <div class="flex items-center justify-between gap-2">
          <FormLabel>Images</FormLabel>
          <Badge v-if="runtime" variant="secondary">via {{ runtime }}</Badge>
        </div>
        <p v-if="loadError" class="text-sm text-destructive">{{ loadError }}</p>
        <FormControl>
          <LocalImagePicker
            v-model="selected"
            :images="images"
            :loading="loading"
            @refresh="loadImages"
          />
        </FormControl>
        <p v-if="selected.length" class="text-xs text-muted-foreground">
          {{ selected.length }} selected
        </p>
        <FormMessage />
      </FormItem>
    </FormField>

    <div class="flex flex-col gap-2">
      <Label for="custom-image">
        Image not listed?
        <span class="text-xs font-normal text-muted-foreground">(optional)</span>
      </Label>
      <div class="flex items-center gap-2">
        <Input
          id="custom-image"
          v-model="customImage"
          placeholder="nginx:1.27"
          autocomplete="off"
          @keydown.enter.prevent="addCustom"
        />
        <Button type="button" variant="outline" :disabled="!customImage.trim()" @click="addCustom">
          Add
        </Button>
      </div>
    </div>
  </div>
</template>
