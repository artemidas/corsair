import { computed, readonly, ref, shallowRef } from "vue";
import { invoke } from "@tauri-apps/api/core";

export interface ClusterStatus {
  connected: boolean;
  healthy: boolean;
  context: string | null;
  error: string | null;
}

export interface KubeContexts {
  current: string | null;
  contexts: string[];
}

const POLL_MS = 5000;

const disconnected: ClusterStatus = {
  connected: false,
  healthy: false,
  context: null,
  error: null,
};

const status = ref<ClusterStatus>({ ...disconnected });
const connecting = shallowRef(false);
const disconnecting = shallowRef(false);
const contexts = ref<string[]>([]);
const defaultContext = shallowRef<string | null>(null);
const contextsError = shallowRef("");

let pollTimer: ReturnType<typeof setInterval> | null = null;
let probeInFlight = false;

function applyStatus(next: ClusterStatus) {
  status.value = next;
}

async function probe() {
  if (connecting.value || disconnecting.value || probeInFlight) return;
  probeInFlight = true;
  try {
    applyStatus(await invoke<ClusterStatus>("probe_cluster"));
  } catch (err) {
    applyStatus({
      ...status.value,
      healthy: false,
      error: String(err),
    });
  } finally {
    probeInFlight = false;
  }
}

export function useCluster() {
  const isConnected = computed(
    () => status.value.connected && status.value.healthy,
  );
  const contextLabel = computed(
    () => status.value.context ?? "active context",
  );

  async function loadContexts() {
    contextsError.value = "";
    try {
      const result = await invoke<KubeContexts>("list_kube_contexts");
      contexts.value = result.contexts;
      defaultContext.value = result.current;
    } catch (err) {
      contexts.value = [];
      defaultContext.value = null;
      contextsError.value = String(err);
    }
  }

  async function connect(context: string | null) {
    connecting.value = true;
    try {
      applyStatus(
        await invoke<ClusterStatus>("connect_cluster", {
          context: context || null,
        }),
      );
    } catch (err) {
      applyStatus({
        connected: status.value.connected,
        healthy: status.value.healthy,
        context: status.value.context,
        error: String(err),
      });
      throw err;
    } finally {
      connecting.value = false;
    }
  }

  async function disconnect() {
    disconnecting.value = true;
    try {
      applyStatus(await invoke<ClusterStatus>("disconnect_cluster"));
    } finally {
      disconnecting.value = false;
    }
  }

  function startPolling() {
    if (pollTimer !== null) return;
    void probe();
    pollTimer = setInterval(() => {
      void probe();
    }, POLL_MS);
  }

  function stopPolling() {
    if (pollTimer === null) return;
    clearInterval(pollTimer);
    pollTimer = null;
  }

  return {
    status: readonly(status),
    connecting: readonly(connecting),
    disconnecting: readonly(disconnecting),
    contexts: readonly(contexts),
    defaultContext: readonly(defaultContext),
    contextsError: readonly(contextsError),
    isConnected,
    contextLabel,
    loadContexts,
    connect,
    disconnect,
    probe,
    startPolling,
    stopPolling,
  };
}
