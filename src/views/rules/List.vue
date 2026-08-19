<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { BookCheck, CircleAlert, Plus, Trash2 } from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { RuleEditor } from "@/components/rule";
import {
  useCustomRules,
  isBuiltIn,
  type CustomRule,
} from "@/composables/useCustomRules";
import { severityBadgeVariant } from "@/lib/severity";
import { confirm } from "@tauri-apps/plugin-dialog";

const router = useRouter();
const { userRules, builtInRules, loading, loadError, deleteRule } = useCustomRules();

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
</script>

<template>
  <div>
    <Card>
      <CardHeader>
        <CardTitle>Rules</CardTitle>
        <CardDescription>
          Your own matchers plus the built-in checks that ship with Corsair.
          Built-in rules are read-only.
        </CardDescription>
        <CardAction>
          <Button size="sm" @click="openNew">
            <Plus />
            New rule
          </Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Alert v-if="loadError" variant="destructive">
          <CircleAlert />
          <AlertTitle>Could not load rules</AlertTitle>
          <AlertDescription>{{ loadError }}</AlertDescription>
        </Alert>
        <div
          v-else-if="loading && items.length === 0"
          class="flex flex-col gap-2"
        >
          <Skeleton class="h-16 w-full" />
          <Skeleton class="h-16 w-full" />
        </div>
        <Empty v-else-if="items.length === 0" class="border-0 p-6 md:p-8">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <BookCheck />
            </EmptyMedia>
            <EmptyTitle>No rules yet</EmptyTitle>
            <EmptyDescription>
              Click "New rule" to add one.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
        <ul v-else class="flex flex-col gap-2">
          <li
            v-for="r in items"
            :key="r.id"
            class="group flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors hover:bg-accent"
            @click="open(r)"
          >
            <BookCheck class="size-5 shrink-0 text-muted-foreground" />
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span class="truncate font-medium">{{ r.title }}</span>
                <Badge v-if="isBuiltIn(r)" variant="outline">Built-in</Badge>
              </div>
              <div
                v-if="r.description"
                class="truncate text-xs text-muted-foreground"
              >
                {{ r.description }}
              </div>
              <div class="mt-1 font-mono text-xs text-muted-foreground">
                {{ r.resourceType }} · {{ r.fieldPath }} == {{ r.expectedValue }}
              </div>
            </div>
            <Badge :variant="severityBadgeVariant(r.severity)">
              {{ r.severity }}
            </Badge>
            <Button
              v-if="!isBuiltIn(r)"
              variant="ghost"
              size="icon-sm"
              class="shrink-0 opacity-0 group-hover:opacity-100"
              title="Delete rule"
              @click.stop="onDelete(r, $event)"
            >
              <Trash2 />
            </Button>
          </li>
        </ul>
      </CardContent>
    </Card>

    <RuleEditor
      v-if="editorOpen"
      :rule="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
