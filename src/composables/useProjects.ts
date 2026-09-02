import { ref } from "vue";
import type { RouteLocationRaw } from "vue-router";
import {
  CreateProject as createProjectBound,
  DeleteProject as deleteProjectBound,
  ListProjects,
  UpdateProject as updateProjectBound,
} from "@/bindings/ladon/project/service";
import type {
  Project as BoundProject,
  ProjectInput as BoundProjectInput,
  ProjectKind as BoundProjectKind,
} from "@/bindings/ladon/project/models";

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

function toBoundInput(input: ProjectInput): BoundProjectInput {
  return {
    name: input.name,
    kind: input.kind as BoundProjectKind,
    config: {
      context: input.config.context,
      image: input.config.image,
      images: input.config.images ?? [],
    },
  };
}

function fromBound(project: BoundProject): Project {
  return {
    id: project.id,
    name: project.name,
    kind: project.kind as ProjectKind,
    config: {
      context: project.config.context,
      image: project.config.image,
      images: project.config.images ?? [],
    },
    createdAt: project.createdAt,
    updatedAt: project.updatedAt,
  };
}

const projects = ref<Project[]>([]);
const loading = ref(false);
const loadError = ref("");

export function useProjects() {
  async function loadProjects() {
    loading.value = true;
    loadError.value = "";
    try {
      projects.value = ((await ListProjects()) ?? []).map(fromBound);
    } catch (err) {
      loadError.value = String(err);
    } finally {
      loading.value = false;
    }
  }

  async function createProject(input: ProjectInput): Promise<Project> {
    const created = fromBound(await createProjectBound(toBoundInput(input)));
    projects.value = [...projects.value, created];
    return created;
  }

  async function updateProject(id: string, input: ProjectInput): Promise<Project> {
    const updated = fromBound(await updateProjectBound(id, toBoundInput(input)));
    projects.value = projects.value.map((p) => (p.id === id ? updated : p));
    return updated;
  }

  async function deleteProject(id: string) {
    await deleteProjectBound(id);
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
