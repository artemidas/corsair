import type { ProjectKind } from "@/composables/useProjects";

export interface ProjectFormValues {
  name: string;
  kind: ProjectKind;
  context: string;
  image: string;
}
