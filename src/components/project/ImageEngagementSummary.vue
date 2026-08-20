<script setup lang="ts">
import { computed, onMounted } from "vue";
import { Box } from "@lucide/vue";
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from "@/components/ui/item";
import { Skeleton } from "@/components/ui/skeleton";
import { imagesFor, type Project } from "@/composables/useProjects";
import {
  formatBytes,
  imageRowStats,
  useLocalImages,
} from "@/composables/useLocalImages";

const props = defineProps<{
  project: Project;
}>();

const { images, loading, loadImages } = useLocalImages();

onMounted(() => {
  void loadImages();
});

const references = computed(() => imagesFor(props.project.config));

const rows = computed(() =>
  references.value.map((reference) =>
    imageRowStats(reference, images.value),
  ),
);

const countLabel = computed(() => {
  const n = rows.value.length;
  return n === 1 ? "1 image" : `${n} images`;
});

const totalLabel = computed(() => {
  if (rows.value.length === 0) return "";
  if (rows.value.some((row) => row.sizeBytes == null)) return "";
  const total = rows.value.reduce((sum, row) => sum + (row.sizeBytes ?? 0), 0);
  return formatBytes(total);
});

const summary = computed(() =>
  [countLabel.value, totalLabel.value].filter(Boolean).join(" · "),
);
</script>

<template>
  <div class="flex flex-col gap-4">
    <Item>
      <ItemContent>
        <ItemTitle>
          <h3 class="scroll-m-20 text-2xl font-semibold tracking-tight">
            {{ project.name }}
          </h3>
        </ItemTitle>
        <ItemDescription class="line-clamp-none">
          {{ summary }}
        </ItemDescription>
      </ItemContent>
    </Item>

    <div v-if="loading && images.length === 0" class="flex flex-col gap-2">
      <Skeleton class="h-16 w-full" />
      <Skeleton class="h-16 w-full" />
    </div>

    <ItemGroup v-else class="rounded-md border">
      <Item
        v-for="row in rows"
        :key="row.reference"
        size="sm"
        class="border-b border-border last:border-b-0"
      >
        <ItemMedia variant="icon">
          <Box />
        </ItemMedia>
        <ItemContent>
          <ItemTitle class="font-mono">{{ row.reference }}</ItemTitle>
          <ItemDescription>
            <span v-if="row.size">{{ row.size }}</span>
            <span v-if="row.size && row.id"> · </span>
            <span v-if="row.id" class="font-mono">{{ row.id }}</span>
            <span v-if="!row.size && !row.id">Not found locally</span>
          </ItemDescription>
        </ItemContent>
      </Item>
    </ItemGroup>
  </div>
</template>
