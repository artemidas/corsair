<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { BookCheck, Plus, Trash2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { RuleEditor } from "@/components/rule";
import {
  useCustomRules,
  isBuiltIn,
  type CustomRule,
} from "@/composables/useCustomRules";
import { confirm } from "@tauri-apps/plugin-dialog";

const router = useRouter();
const { userRules, builtInRules, loading, loadError, deleteRule } = useCustomRules();

// User rules first — built-ins are read-only reference material.
const items = computed<CustomRule[]>(() => [
  ...userRules.value,
  ...builtInRules.value,
]);

const editorOpen = ref(false);
const editorTarget = ref<CustomRule | null>(null);

function open(rule: CustomRule) {
  router.push({ name: "rule", params: { id: rule.id } });
}

function openNew() {
  editorTarget.value = null;
  editorOpen.value = true;
}

async function onDelete(rule: CustomRule, ev: Event) {
  ev.stopPropagation();
  ev.preventDefault();
  const confirmed = await confirm(`Delete rule "${rule.title}"?`, {
    title: "Tauri",
    kind: "warning",
  });
  if (!confirmed) return;
  try {
    await deleteRule(rule.id);
  } catch (err) {
    console.error("Failed to delete rule:", err);
  }
}

const severityBadgeClass: Record<string, string> = {
  critical: "bg-red-600 text-white",
  high: "bg-orange-500 text-white",
  medium: "bg-yellow-400 text-black",
  low: "bg-gray-400 text-black",
};
</script>

<template>
  <div>
    <div class="card bg-base-100 shadow">
      <div class="card-body gap-3">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h2 class="card-title">Rules</h2>
            <p class="text-sm text-base-content/60">
              Your own matchers plus the built-in checks that ship with Corsair.
              Built-in rules are read-only.
            </p>
          </div>
          <Button size="sm" @click="openNew">
            <Plus class="size-4" />
            New rule
          </Button>
        </div>

        <div v-if="loadError" class="alert alert-error text-sm">
          {{ loadError }}
        </div>
        <div
          v-else-if="loading && items.length === 0"
          class="py-6 text-sm text-base-content/50"
        >
          Loading…
        </div>
        <div
          v-else-if="items.length === 0"
          class="py-6 text-sm text-base-content/50"
        >
          No rules yet. Click "New rule" to add one.
        </div>

        <ul v-else class="flex flex-col gap-2">
          <li
            v-for="r in items"
            :key="r.id"
            class="group flex cursor-pointer items-center gap-3 rounded-lg border border-base-300 p-3 transition-colors hover:bg-base-200"
            @click="open(r)"
          >
            <BookCheck class="size-5 shrink-0 text-primary" />
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="truncate font-medium">{{ r.title }}</span>
                <span
                  v-if="isBuiltIn(r)"
                  class="badge badge-xs badge-outline shrink-0"
                >
                  Built-in
                </span>
              </div>
              <div
                v-if="r.description"
                class="truncate text-xs text-base-content/50"
              >
                {{ r.description }}
              </div>
              <div class="mt-1 font-mono text-xs text-base-content/40">
                {{ r.resourceType }} · {{ r.fieldPath }} == {{ r.expectedValue }}
              </div>
            </div>
            <span class="badge shrink-0" :class="severityBadgeClass[r.severity]">
              {{ r.severity }}
            </span>
            <Button
              v-if="!isBuiltIn(r)"
              variant="ghost"
              size="icon-sm"
              class="shrink-0 opacity-0 group-hover:opacity-100"
              title="Delete rule"
              @click.stop="onDelete(r, $event)"
            >
              <Trash2 class="size-4" />
            </Button>
          </li>
        </ul>
      </div>
    </div>

    <RuleEditor
      v-if="editorOpen"
      :rule="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
