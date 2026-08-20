import { computed } from "vue";
import { useRoute, type RouteLocationRaw } from "vue-router";
import {
  listLabelForKind,
  listRouteForKind,
  useProjects,
  type ProjectKind,
} from "@/composables/useProjects";
import { useRules } from "@/composables/useRules";

export interface BreadcrumbCrumb {
  label: string;
  to?: RouteLocationRaw;
}

function kindListCrumb(kind: ProjectKind | undefined): BreadcrumbCrumb {
  const resolved = kind ?? "kubernetesClusterReview";
  return {
    label: listLabelForKind(resolved),
    to: listRouteForKind(resolved),
  };
}

export function usePageHeader() {
  const route = useRoute();
  const { getProjectById } = useProjects();
  const { getRuleById } = useRules();

  const projectId = computed(() =>
    typeof route.params.id === "string" ? route.params.id : undefined,
  );
  const project = computed(() =>
    projectId.value ? getProjectById(projectId.value) : null,
  );
  const projectKind = computed<ProjectKind | undefined>(
    () => route.meta.projectKind ?? project.value?.kind,
  );

  const crumbs = computed<BreadcrumbCrumb[]>(() => {
    const id = projectId.value;
    const projectName = project.value?.name ?? "Project";
    const ruleTitle = id ? (getRuleById(id)?.title ?? "Rule") : "Rule";
    const listCrumb = kindListCrumb(projectKind.value);

    switch (route.name) {
      case "home":
        return [{ label: "Home" }];
      case "cluster-projects":
        return [{ label: "Clusters" }];
      case "image-projects":
        return [{ label: "Images" }];
      case "new-project":
        return [{ label: "New Project" }];
      case "new-kubernetes-project":
        return [listCrumb, { label: "New cluster" }];
      case "new-container-image-project":
        return [listCrumb, { label: "New engagement" }];
      case "project":
        return [listCrumb, { label: projectName }];
      case "edit-project":
        return [
          listCrumb,
          {
            label: projectName,
            to: id ? { name: "project", params: { id } } : listCrumb.to,
          },
          { label: "Edit" },
        ];
      case "finding": {
        const ruleId =
          typeof route.params.ruleId === "string"
            ? route.params.ruleId
            : "Finding";
        return [
          listCrumb,
          {
            label: projectName,
            to: id ? { name: "project", params: { id } } : listCrumb.to,
          },
          { label: ruleId },
        ];
      }
      case "rules":
        return [{ label: "Rules" }];
      case "rule":
        return [
          { label: "Rules", to: { name: "rules" } },
          { label: ruleTitle },
        ];
      case "settings":
        return [{ label: "Settings" }];
      default:
        return [{ label: route.meta.title ?? "Ladon" }];
    }
  });

  const title = computed(
    () => crumbs.value[crumbs.value.length - 1]?.label ?? "Ladon",
  );
  const nav = computed(() => {
    if (route.meta.nav) return route.meta.nav;
    const kind = projectKind.value;
    if (kind === "containerImageReview") return "images";
    if (kind === "kubernetesClusterReview") return "clusters";
    return undefined;
  });

  return { title, nav, crumbs };
}
