import { computed, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";

export type ProjectKind = "kubernetesClusterReview" | "containerImageReview";

export interface ProjectConfig {
  context: string | null;
  image: string | null;
}

export interface Project {
  id: string;
  name: string;
  kind: ProjectKind;
  config: ProjectConfig;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectInput {
  name: string;
  kind: ProjectKind;
  config: ProjectConfig;
}

interface Connection {
  context: string | null;
}

const projects = ref<Project[]>([]);
const selectedProjectId = ref<string | null>(null);
const loading = ref(false);
const loadError = ref("");
const connection = ref<Connection | null>(null);

export function useProjects() {
  const selectedProject = computed<Project | null>(
    () => projects.value.find((p) => p.id === selectedProjectId.value) ?? null,
  );

  async function loadProjects() {
    loading.value = true;
    loadError.value = "";
    try {
      projects.value = await invoke<Project[]>("list_projects");
    } catch (err) {
      loadError.value = String(err);
    } finally {
      loading.value = false;
    }
  }

  async function createProject(input: ProjectInput): Promise<Project> {
    const created = await invoke<Project>("create_project", { input });
    projects.value = [...projects.value, created];
    return created;
  }

  async function updateProject(id: string, input: ProjectInput): Promise<Project> {
    const updated = await invoke<Project>("update_project", { id, input });
    projects.value = projects.value.map((p) => (p.id === id ? updated : p));
    return updated;
  }

  async function deleteProject(id: string) {
    await invoke("delete_project", { id });
    projects.value = projects.value.filter((p) => p.id !== id);
    if (selectedProjectId.value === id) {
      selectedProjectId.value = null;
    }
  }

  function selectProject(id: string | null) {
    selectedProjectId.value = id;
  }

  async function refreshConnection() {
    connection.value = await invoke<Connection | null>("active_context");
  }

  function setConnection(ctx: string | null) {
    connection.value = { context: ctx };
  }

  function isConnectedTo(project: Project): boolean {
    if (project.kind !== "kubernetesClusterReview") return false;
    return connection.value?.context === (project.config.context ?? null);
  }

  return {
    projects,
    selectedProjectId,
    selectedProject,
    loading,
    loadError,
    connection,
    loadProjects,
    createProject,
    updateProject,
    deleteProject,
    selectProject,
    refreshConnection,
    setConnection,
    isConnectedTo,
  };
}
