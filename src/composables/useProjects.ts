import { ref } from "vue";
import { invoke } from "@tauri-apps/api/core";
import type { RouteLocationRaw } from "vue-router";

export type ProjectKind = "kubernetesClusterReview" | "containerImageReview";

export interface ProjectConfig {
  context: string | null;
  /** @deprecated Folded into `images` on read. */
  image?: string | null;
  images?: string[];
}

export function imagesFor(config: ProjectConfig): string[] {
  const listed = (config.images ?? [])
    .map((image) => image.trim())
    .filter(Boolean);
  if (listed.length > 0) return [...new Set(listed)];
  const single = config.image?.trim();
  return single ? [single] : [];
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

export function listRouteForKind(kind: ProjectKind): RouteLocationRaw {
  return kind === "containerImageReview"
    ? { name: "image-projects" }
    : { name: "cluster-projects" };
}

export function listLabelForKind(kind: ProjectKind): string {
  return kind === "containerImageReview" ? "Images" : "Clusters";
}

const projects = ref<Project[]>([]);
const loading = ref(false);
const loadError = ref("");

export function useProjects() {
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
  }

  function getProjectById(id: string): Project | null {
    return projects.value.find((p) => p.id === id) ?? null;
  }

  return {
    projects,
    loading,
    loadError,
    loadProjects,
    createProject,
    updateProject,
    deleteProject,
    getProjectById,
  };
}
