<script setup lang="ts">
import { computed, shallowRef } from "vue";
import { RefreshCw } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item";
import { cn } from "@/lib/utils";
import type { LocalImage } from "@/composables/useLocalImages";

const props = defineProps<{
  images: LocalImage[];
  loading: boolean;
}>();

const emit = defineEmits<{
  refresh: [];
}>();

const selected = defineModel<string[]>({ default: () => [] });
const query = shallowRef("");

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  if (!q) return props.images;
  return props.images.filter((img) => img.reference.toLowerCase().includes(q));
});

const allFilteredSelected = computed(
  () =>
    filtered.value.length > 0 &&
    filtered.value.every((img) => selected.value.includes(img.reference)),
);

function shortId(id: string): string {
  const hex = id.replace(/^sha256:/, "");
  return hex.slice(0, 12);
}

function toggle(reference: string) {
  selected.value = selected.value.includes(reference)
    ? selected.value.filter((ref) => ref !== reference)
    : [...selected.value, reference];
}

function toggleFiltered() {
  if (allFilteredSelected.value) {
    const drop = new Set(filtered.value.map((img) => img.reference));
    selected.value = selected.value.filter((ref) => !drop.has(ref));
    return;
  }
  selected.value = [
    ...new Set([
      ...selected.value,
      ...filtered.value.map((img) => img.reference),
    ]),
  ];
}
</script>

<template>
  <div class="flex flex-col gap-3">
    <div class="flex items-center gap-2">
      <Input
        v-model="query"
        placeholder="Filter images"
        autocomplete="off"
        class="flex-1"
      />
      <Button
        type="button"
        variant="outline"
        size="sm"
        :disabled="loading || filtered.length === 0"
        @click="toggleFiltered"
      >
        {{ allFilteredSelected ? "Clear" : "Select all" }}
      </Button>
      <Button
        type="button"
        variant="outline"
        size="icon-sm"
        :disabled="loading"
        title="Refresh images"
        @click="emit('refresh')"
      >
        <RefreshCw :class="cn(loading && 'animate-spin')" />
        <span class="sr-only">Refresh images</span>
      </Button>
    </div>

    <div v-if="loading && images.length === 0" class="flex flex-col gap-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <p
      v-else-if="filtered.length === 0"
      class="rounded-md border border-dashed px-3 py-6 text-center text-sm text-muted-foreground"
    >
      {{ images.length === 0 ? "No local images found." : "No images match that filter." }}
    </p>

    <ItemGroup v-else class="max-h-96 overflow-y-auto rounded-md border">
      <Item
        v-for="img in filtered"
        :key="img.reference"
        as="label"
        size="sm"
        class="cursor-pointer border-b border-border last:border-b-0"
        :class="selected.includes(img.reference) && 'bg-accent/50'"
      >
        <ItemMedia>
          <input
            type="checkbox"
            class="size-4 accent-primary"
            :checked="selected.includes(img.reference)"
            @change="toggle(img.reference)"
          />
        </ItemMedia>
        <ItemContent>
          <ItemTitle class="font-mono">{{ img.reference }}</ItemTitle>
          <ItemDescription>
            <span v-if="img.size">{{ img.size }}</span>
            <span v-if="img.size && img.id"> · </span>
            <span v-if="img.id" class="font-mono">{{ shortId(img.id) }}</span>
          </ItemDescription>
        </ItemContent>
      </Item>
    </ItemGroup>
  </div>
</template>
