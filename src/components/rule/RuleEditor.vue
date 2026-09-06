<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { CircleAlert } from "@lucide/vue";
import {
  DEFAULT_REGO,
  RESOURCE_TYPES,
  useRules,
  type Rule,
  type RuleInput,
  type RuleSeverity,
} from "@/composables/useRules";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
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

const props = defineProps<{
  rule: Rule | null;
}>();

const emit = defineEmits<{
  close: [];
}>();

const { createRule, updateRule } = useRules();

const isEdit = computed(() => props.rule !== null);
const title = computed(() => (isEdit.value ? "Edit rule" : "New rule"));
const isSubmitting = ref(false);
const submitError = ref<string | null>(null);

const form = useForm<RuleInput>({
  validationSchema: toTypedSchema(
    z.object({
      title: z.string().trim().min(1, "Title is required."),
      description: z.string(),
      severity: z.enum(["critical", "high", "medium", "low"]),
      resourceType: z.string().min(1, "Resource type is required."),
      rego: z.string().trim().min(1, "Rego policy is required."),
    }),
  ),
  initialValues: {
    title: "",
    description: "",
    severity: "medium" as RuleSeverity,
    resourceType: "Pod",
    rego: DEFAULT_REGO,
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
        rego: props.rule.rego,
      },
    });
  } else {
    form.resetForm({
      values: {
        title: "",
        description: "",
        severity: "medium",
        resourceType: "Pod",
        rego: DEFAULT_REGO,
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
    <DialogContent class="sm:max-w-2xl">
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

        <FormField v-slot="{ componentField }" name="rego">
          <FormItem>
            <FormLabel>
              Rego policy
              <span class="text-xs font-normal text-muted-foreground">
                package
                <span class="font-mono">ladon</span>, define
                <span class="font-mono">violation</span>
              </span>
            </FormLabel>
            <FormControl>
              <Textarea
                v-bind="componentField"
                placeholder="package ladon"
                class="min-h-48"
                spellcheck="false"
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
