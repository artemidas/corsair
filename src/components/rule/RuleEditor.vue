<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { CircleAlert } from "@lucide/vue";
import {
  RESOURCE_TYPES,
  type CustomRule,
  type CustomRuleInput,
  type RuleSeverity,
} from "@/composables/useCustomRules";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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

function close() {
  emit("close");
}

function onOpenChange(open: boolean) {
  if (!open) close();
}
</script>

<template>
  <Dialog :open="true" @update:open="onOpenChange">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>{{ title }}</DialogTitle>
      </DialogHeader>

      <Alert v-if="submitError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Could not save rule</AlertTitle>
        <AlertDescription>{{ submitError }}</AlertDescription>
      </Alert>

      <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
        <FormField v-slot="{ componentField }" name="title">
          <FormItem>
            <FormLabel>Title</FormLabel>
            <FormControl>
              <Input
                v-bind="componentField"
                placeholder="No privileged containers"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="description">
          <FormItem>
            <FormLabel>Description</FormLabel>
            <FormControl>
              <Input
                v-bind="componentField"
                placeholder="Optional — what's the impact of this check?"
              />
            </FormControl>
          </FormItem>
        </FormField>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <FormField v-slot="{ componentField }" name="severity">
            <FormItem>
              <FormLabel>Severity</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger class="w-full">
                    <SelectValue placeholder="Select severity" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem value="critical">critical</SelectItem>
                    <SelectItem value="high">high</SelectItem>
                    <SelectItem value="medium">medium</SelectItem>
                    <SelectItem value="low">low</SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </FormItem>
          </FormField>

          <FormField v-slot="{ componentField }" name="resourceType">
            <FormItem>
              <FormLabel>Resource</FormLabel>
              <Select v-bind="componentField">
                <FormControl>
                  <SelectTrigger class="w-full">
                    <SelectValue placeholder="Select resource" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectGroup>
                    <SelectItem
                      v-for="r in RESOURCE_TYPES"
                      :key="r"
                      :value="r"
                    >
                      {{ r }}
                    </SelectItem>
                  </SelectGroup>
                </SelectContent>
              </Select>
            </FormItem>
          </FormField>
        </div>

        <FormField v-slot="{ componentField }" name="fieldPath">
          <FormItem>
            <FormLabel>
              Field path
              <span class="text-xs font-normal text-muted-foreground">
                (use <span class="font-mono">[*]</span> for arrays)
              </span>
            </FormLabel>
            <FormControl>
              <Input
                v-bind="componentField"
                placeholder="spec.containers[*].securityContext.privileged"
                class="font-mono"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <FormField v-slot="{ componentField }" name="expectedValue">
          <FormItem>
            <FormLabel>Expected value</FormLabel>
            <FormControl>
              <Input
                v-bind="componentField"
                placeholder="true"
                class="font-mono"
              />
            </FormControl>
            <FormMessage />
          </FormItem>
        </FormField>

        <DialogFooter>
          <Button
            type="button"
            variant="ghost"
            :disabled="isSubmitting"
            @click="close"
          >
            Cancel
          </Button>
          <Button type="submit" :disabled="isSubmitting">
            <Spinner v-if="isSubmitting" />
            {{ isEdit ? "Save" : "Create" }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
