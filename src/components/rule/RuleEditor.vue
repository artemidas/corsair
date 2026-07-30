<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import {
  RESOURCE_TYPES,
  type CustomRule,
  type CustomRuleInput,
  type RuleSeverity,
} from "@/composables/useCustomRules";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useForm } from "vee-validate";
import { toTypedSchema } from "@vee-validate/zod";
import { z } from "zod";
import { useCustomRules } from "@/composables/useCustomRules";

const props = defineProps<{
  rule: CustomRule | null;
}>();

const emit = defineEmits<{
  close: [];
}>();

const dialogRef = ref<HTMLDialogElement | null>(null);

const { createRule, updateRule } = useCustomRules();

const isEdit = computed(() => props.rule !== null);
const title = computed(() => (isEdit.value ? "Edit rule" : "New rule"));
const isSubmitting = ref(false);
const submitError = ref<string | null>(null);

const form = useForm<CustomRuleInput>({
  validationSchema: toTypedSchema(
    z.object({
      title: z.string().trim().min(1, "Title is required."),
      description: z.string(),
      severity: z.enum(["critical", "high", "medium", "low"]),
      resourceType: z.string().min(1, "Resource type is required."),
      fieldPath: z.string().trim().min(1, "Field path is required."),
      expectedValue: z.string(),
    }),
  ),
  initialValues: {
    title: "",
    description: "",
    severity: "medium" as RuleSeverity,
    resourceType: "Pod",
    fieldPath: "",
    expectedValue: "",
  },
});

const onSubmit = form.handleSubmit(async (values) => {
  isSubmitting.value = true;
  submitError.value = null;
  try {
    if (isEdit.value && props.rule) {
      await updateRule(props.rule.id, values);
    } else {
      await createRule(values);
    }
    close();
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : String(err);
  } finally {
    isSubmitting.value = false;
  }
});

function reset() {
  if (props.rule) {
    form.resetForm({
      values: {
        title: props.rule.title,
        description: props.rule.description,
        severity: props.rule.severity,
        resourceType: props.rule.resourceType,
        fieldPath: props.rule.fieldPath,
        expectedValue: props.rule.expectedValue,
      },
    });
  } else {
    form.resetForm({
      values: {
        title: "",
        description: "",
        severity: "medium",
        resourceType: "Pod",
        fieldPath: "",
        expectedValue: "",
      },
    });
  }
}

watch(
  () => props.rule,
  () => reset(),
  { immediate: true },
);

onMounted(() => {
  dialogRef.value?.showModal();
});

function close() {
  dialogRef.value?.close();
  emit("close");
}
</script>

<template>
  <dialog ref="dialogRef" class="modal" @close="emit('close')">
    <div class="modal-box max-w-lg">
      <h3 class="text-lg font-bold">{{ title }}</h3>

      <div v-if="submitError" class="alert alert-error mt-4 text-sm" role="alert">
        <span>{{ submitError }}</span>
      </div>

      <form class="mt-4 flex flex-col gap-4" @submit.prevent="onSubmit">
        <FormField v-slot="{ field }" name="title">
          <FormItem>
            <FormLabel>Title</FormLabel>
            <FormControl>
              <Input v-bind="field" placeholder="No privileged containers" />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ field }" name="description">
          <FormItem>
            <FormLabel>Description</FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="Optional — what's the impact of this check?"
              />
            </FormControl>
          </FormItem>
        </FormField>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <FormField v-slot="{ field }" name="severity">
            <FormItem>
              <FormLabel>Severity</FormLabel>
              <FormControl>
                <select
                  v-bind="field"
                  class="select select-bordered w-full"
                  :value="field.value"
                  @change="(e) => field.onChange((e.target as HTMLSelectElement).value)"
                >
                  <option value="critical">critical</option>
                  <option value="high">high</option>
                  <option value="medium">medium</option>
                  <option value="low">low</option>
                </select>
              </FormControl>
            </FormItem>
          </FormField>

          <FormField v-slot="{ field }" name="resourceType">
            <FormItem>
              <FormLabel>Resource</FormLabel>
              <FormControl>
                <select
                  v-bind="field"
                  class="select select-bordered w-full"
                  :value="field.value"
                  @change="(e) => field.onChange((e.target as HTMLSelectElement).value)"
                >
                  <option v-for="r in RESOURCE_TYPES" :key="r" :value="r">
                    {{ r }}
                  </option>
                </select>
              </FormControl>
            </FormItem>
          </FormField>
        </div>

        <FormField v-slot="{ field }" name="fieldPath">
          <FormItem>
            <FormLabel>
              Field path
              <span class="text-xs font-normal text-muted-foreground">
                (use <span class="font-mono">[*]</span> for arrays)
              </span>
            </FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="spec.containers[*].securityContext.privileged"
                class="font-mono"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ field }" name="expectedValue">
          <FormItem>
            <FormLabel>Expected value</FormLabel>
            <FormControl>
              <Input
                v-bind="field"
                placeholder="true"
                class="font-mono"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <div class="modal-action mt-2">
          <Button
            type="button"
            variant="ghost"
            :disabled="isSubmitting"
            @click="close"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="isSubmitting">
            <span v-if="isSubmitting" class="loading loading-spinner loading-sm"></span>
            {{ isEdit ? "Save" : "Create" }}
          </Button>
        </div>
      </form>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button>close</button>
    </form>
  </dialog>
</template>
