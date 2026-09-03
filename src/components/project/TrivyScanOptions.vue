<script setup lang="ts">
import { computed } from "vue";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  defaultTrivyScanOptions,
  parseExtraArgs,
  trivyScanners,
  trivySeverities,
  type TrivyScanOptions,
} from "@/lib/trivy";

const options = defineModel<TrivyScanOptions>({
  default: defaultTrivyScanOptions,
});

const extraArgsText = computed({
  get: () => options.value.extraArgs.join(" "),
  set: (value: string) => {
    options.value = {
      ...options.value,
      extraArgs: parseExtraArgs(value),
    };
  },
});

function toggleScanner(id: string, checked: boolean) {
  const current = new Set(options.value.scanners);
  if (checked) {
    current.add(id);
  } else {
    current.delete(id);
  }
  options.value = {
    ...options.value,
    scanners: [...current],
  };
}

function toggleSeverity(id: string, checked: boolean) {
  const current = new Set(options.value.severity);
  if (checked) {
    current.add(id);
  } else {
    current.delete(id);
  }
  options.value = {
    ...options.value,
    severity: [...current],
  };
}

function setIgnoreUnfixed(checked: boolean) {
  options.value = {
    ...options.value,
    ignoreUnfixed: checked,
  };
}

function setSkipDbUpdate(checked: boolean) {
  options.value = {
    ...options.value,
    skipDbUpdate: checked,
  };
}
</script>

<template>
  <div class="flex flex-col gap-4 rounded-md border p-4">
    <div class="flex flex-col gap-2">
      <Label>Scanners</Label>
      <div class="flex flex-wrap gap-4">
        <label
          v-for="scanner in trivyScanners"
          :key="scanner.id"
          class="flex items-center gap-2 text-sm"
        >
          <input
            type="checkbox"
            class="size-4 accent-primary"
            :checked="options.scanners.includes(scanner.id)"
            @change="toggleScanner(scanner.id, ($event.target as HTMLInputElement).checked)"
          />
          {{ scanner.label }}
        </label>
      </div>
    </div>

    <div class="flex flex-col gap-2">
      <Label>Severity filter</Label>
      <p class="text-xs text-muted-foreground">
        Leave all unchecked to report every severity.
      </p>
      <div class="flex flex-wrap gap-4">
        <label
          v-for="severity in trivySeverities"
          :key="severity"
          class="flex items-center gap-2 text-sm"
        >
          <input
            type="checkbox"
            class="size-4 accent-primary"
            :checked="options.severity.includes(severity)"
            @change="toggleSeverity(severity, ($event.target as HTMLInputElement).checked)"
          />
          {{ severity }}
        </label>
      </div>
    </div>

    <label class="flex items-center gap-2 text-sm">
      <input
        type="checkbox"
        class="size-4 accent-primary"
        :checked="options.ignoreUnfixed"
        @change="setIgnoreUnfixed(($event.target as HTMLInputElement).checked)"
      />
      Ignore unfixed vulnerabilities
    </label>

    <label class="flex items-center gap-2 text-sm">
      <input
        type="checkbox"
        class="size-4 accent-primary"
        :checked="options.skipDbUpdate"
        @change="setSkipDbUpdate(($event.target as HTMLInputElement).checked)"
      />
      Skip vulnerability DB update
    </label>
    <p class="-mt-2 text-xs text-muted-foreground">
      Keep enabled after the first successful scan to avoid macOS Keychain prompts. The DB is downloaded automatically on first run.
    </p>

    <div class="flex flex-col gap-2">
      <Label for="trivy-extra-args">Extra Trivy flags</Label>
      <Input
        id="trivy-extra-args"
        v-model="extraArgsText"
        placeholder="--skip-db-update --timeout 5m"
        autocomplete="off"
        class="font-mono text-sm"
      />
      <p class="text-xs text-muted-foreground">
        Space-separated flags passed through to
        <span class="font-mono">trivy image</span> before the image reference.
      </p>
    </div>
  </div>
</template>
