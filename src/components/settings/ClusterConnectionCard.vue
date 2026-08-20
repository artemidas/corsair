<script setup lang="ts">
import { computed, onMounted, shallowRef } from "vue";
import {
  CircleAlert,
  CircleCheck,
  RefreshCw,
  TriangleAlert,
  Unplug,
} from "@lucide/vue";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Spinner } from "@/components/ui/spinner";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCluster } from "@/composables/useCluster";

const {
  status,
  connecting,
  disconnecting,
  contexts,
  defaultContext,
  contextsError,
  isConnected,
  contextLabel,
  loadContexts,
  connect,
  disconnect,
} = useCluster();

const selectedContext = shallowRef("");
const actionError = shallowRef("");
const refreshing = shallowRef(false);
const busy = computed(() => connecting.value || disconnecting.value);

function onSelectContext(value: unknown) {
  if (typeof value === "string") selectedContext.value = value;
}

async function refreshContexts() {
  refreshing.value = true;
  try {
    await loadContexts();
    if (!selectedContext.value) {
      selectedContext.value =
        status.value.context ??
        defaultContext.value ??
        contexts.value[0] ??
        "";
    }
  } finally {
    refreshing.value = false;
  }
}

async function onConnect() {
  actionError.value = "";
  try {
    await connect(selectedContext.value || null);
  } catch (err) {
    actionError.value = String(err);
  }
}

async function onDisconnect() {
  actionError.value = "";
  await disconnect();
}

onMounted(async () => {
  await loadContexts();
  selectedContext.value =
    status.value.context ??
    defaultContext.value ??
    contexts.value[0] ??
    "";
});
</script>

<template>
  <Card>
    <CardHeader>
      <CardTitle>Cluster connection</CardTitle>
      <CardDescription>
        Connect through your local kubeconfig. Ladon checks the apiserver
        every 5 seconds and shows the result in the header.
      </CardDescription>
      <CardAction>
        <Button
          variant="ghost"
          size="sm"
          :disabled="refreshing"
          @click="refreshContexts"
        >
          <Spinner v-if="refreshing" />
          <RefreshCw v-else />
          Refresh contexts
        </Button>
      </CardAction>
    </CardHeader>

    <CardContent class="flex flex-col gap-4">
      <Alert v-if="actionError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Failed to connect</AlertTitle>
        <AlertDescription>{{ actionError }}</AlertDescription>
      </Alert>

      <Alert v-if="isConnected">
        <CircleCheck color="green" />
        <AlertTitle>Connected</AlertTitle>
        <AlertDescription>
          Using context
          <span class="font-mono">{{ contextLabel }}</span>
        </AlertDescription>
      </Alert>
      <Alert v-else-if="status.connected && !status.healthy" variant="destructive">
        <TriangleAlert />
        <AlertTitle>Cluster unreachable</AlertTitle>
        <AlertDescription>
          Context
          <span class="font-mono">{{ contextLabel }}</span>
          is stored, but the apiserver did not respond.
          <span v-if="status.error"> {{ status.error }}</span>
        </AlertDescription>
      </Alert>
      <Alert v-else>
        <Unplug />
        <AlertTitle>Not connected</AlertTitle>
        <AlertDescription>
          Pick a kubeconfig context and connect to run scans.
        </AlertDescription>
      </Alert>

      <Alert v-if="contextsError" variant="destructive">
        <CircleAlert />
        <AlertTitle>Could not read kubeconfig</AlertTitle>
        <AlertDescription>{{ contextsError }}</AlertDescription>
      </Alert>

      <div class="flex flex-col gap-2">
        <Label for="kube-context">Kubeconfig context</Label>
        <Select
          v-if="contexts.length > 0"
          :model-value="selectedContext"
          @update:model-value="onSelectContext"
        >
          <SelectTrigger id="kube-context" class="w-full font-mono">
            <SelectValue placeholder="Select a context" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem
                v-for="name in contexts"
                :key="name"
                :value="name"
                class="font-mono"
              >
                {{ name }}
                <span
                  v-if="name === defaultContext"
                  class="text-muted-foreground"
                >
                  (current)
                </span>
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
        <Input
          v-else
          id="kube-context"
          v-model="selectedContext"
          class="font-mono"
          placeholder="leave empty to use the active context"
          autocomplete="off"
        />
      </div>
    </CardContent>

    <CardFooter class="flex flex-wrap justify-end gap-2">
      <Button
        v-if="status.connected"
        type="button"
        variant="outline"
        :disabled="busy"
        @click="onDisconnect"
      >
        <Spinner v-if="disconnecting" />
        Disconnect
      </Button>
      <Button type="button" :disabled="busy" @click="onConnect">
        <Spinner v-if="connecting" />
        {{ status.connected ? "Reconnect" : "Connect" }}
      </Button>
    </CardFooter>
  </Card>
</template>
