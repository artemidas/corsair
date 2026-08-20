<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import {
  BookCheck,
  CircleAlert,
  Download,
  Plus,
  Upload,
} from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Item,
  ItemContent,
  ItemTitle,
  ItemDescription,
  ItemActions,
} from "@/components/ui/item";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { RuleEditor, RulesDataTable } from "@/components/rule";
import {
  useCustomRules,
  YAML_FILTERS,
  type CustomRule,
  type ImportSummary,
} from "@/composables/useCustomRules";
import { confirm, open, save } from "@tauri-apps/plugin-dialog";

const router = useRouter();
const {
  userRules,
  builtInRules,
  loading,
  loadError,
  deleteRule,
  importRules,
  exportRules,
} = useCustomRules();

const items = computed<CustomRule[]>(() => [
  ...userRules.value,
  ...builtInRules.value,
]);

const editorOpen = ref(false);
const editorTarget = ref<CustomRule | null>(null);
const actionBusy = ref(false);
const actionError = ref("");
const actionMessage = ref("");

function openRule(rule: CustomRule) {
  router.push({ name: "rule", params: { id: rule.id } });
}

function openNew() {
  editorTarget.value = null;
  editorOpen.value = true;
}

function clearActionStatus() {
  actionError.value = "";
  actionMessage.value = "";
}

async function onDelete(rule: CustomRule) {
  const confirmed = await confirm(`Delete rule "${rule.title}"?`, {
    title: "Delete rule",
    kind: "warning",
  });
  if (!confirmed) return;
  clearActionStatus();
  try {
    await deleteRule(rule.id);
  } catch (err) {
    actionError.value = String(err);
  }
}

async function onImport() {
  const selected = await open({
    multiple: false,
    title: "Import rules",
    filters: YAML_FILTERS,
  });
  if (!selected || Array.isArray(selected)) return;
  clearActionStatus();
  actionBusy.value = true;
  try {
    const summary = await importRules(selected, "merge");
    actionMessage.value = formatImportSummary(summary);
  } catch (err) {
    actionError.value = String(err);
  } finally {
    actionBusy.value = false;
  }
}

async function onExport() {
  const dest = await save({
    title: "Export rules",
    defaultPath: "corsair-rules.yaml",
    filters: YAML_FILTERS,
  });
  if (!dest) return;
  clearActionStatus();
  actionBusy.value = true;
  try {
    const count = await exportRules(dest, "user");
    actionMessage.value =
      count === 1 ? "Exported 1 rule." : `Exported ${count} rules.`;
  } catch (err) {
    actionError.value = String(err);
  } finally {
    actionBusy.value = false;
  }
}

function formatImportSummary(summary: ImportSummary): string {
  const parts = [`Created ${summary.created}, updated ${summary.updated}.`];
  if (summary.skipped.length > 0) {
    const skipped = summary.skipped
      .map((s) => `${s.id || s.title} (${s.reason})`)
      .join("; ");
    parts.push(`Skipped ${summary.skipped.length}: ${skipped}`);
  }
  return parts.join(" ");
}
</script>

<template>
  <div>
    <Item>
      <ItemContent>
        <ItemTitle>Rules</ItemTitle>
        <ItemDescription>
          Your own matchers plus the built-in checks that ship with Corsair.
          Built-in rules are read-only. Import and export user rules as YAML.
        </ItemDescription>
      </ItemContent>
      <ItemActions>
        <Button
          variant="outline"
          size="sm"
          :disabled="actionBusy"
          @click="onImport"
        >
          <Upload />
          Import
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="actionBusy || userRules.length === 0"
          @click="onExport"
        >
          <Download />
          Export
        </Button>
        <Button size="sm" :disabled="actionBusy" @click="openNew">
          <Plus />
          New rule
        </Button>
      </ItemActions>
    </Item>

    <Alert v-if="loadError" variant="destructive">
      <CircleAlert />
      <AlertTitle>Could not load rules</AlertTitle>
      <AlertDescription>{{ loadError }}</AlertDescription>
    </Alert>
    <template v-else>
      <Alert v-if="actionError" variant="destructive" class="mb-3">
        <CircleAlert />
        <AlertTitle>Rules action failed</AlertTitle>
        <AlertDescription>{{ actionError }}</AlertDescription>
      </Alert>
      <Alert v-else-if="actionMessage" class="mb-3">
        <BookCheck />
        <AlertTitle>{{ actionMessage }}</AlertTitle>
      </Alert>
      <div
        v-if="loading && items.length === 0"
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
            Click "New rule" to add one, or import a YAML file.
          </EmptyDescription>
        </EmptyHeader>
      </Empty>
      <RulesDataTable
        v-else
        :data="items"
        @row-click="openRule"
        @delete="onDelete"
      />
    </template>
    <RuleEditor
      v-if="editorOpen"
      :rule="editorTarget"
      @close="editorOpen = false"
    />
  </div>
</template>
