import { computed } from "vue";
import { useRoute, type RouteLocationRaw } from "vue-router";
import { useProjects } from "@/composables/useProjects";
import { useRules } from "@/composables/useRules";

export interface BreadcrumbCrumb {
  label: string;
  to?: RouteLocationRaw;
}

export function usePageHeader() {
  const route = useRoute();
  const { getProjectById } = useProjects();
  const { getRuleById } = useRules();

  const crumbs = computed<BreadcrumbCrumb[]>(() => {
    const id = typeof route.params.id === "string" ? route.params.id : undefined;
    const projectName = id ? (getProjectById(id)?.name ?? "Project") : "Project";
    const ruleTitle = id ? (getRuleById(id)?.title ?? "Rule") : "Rule";

    switch (route.name) {
      case "home":
        return [{ label: "Home" }];
      case "projects":
        return [{ label: "Projects" }];
      case "new-project":
        return [
          { label: "Projects", to: { name: "projects" } },
          { label: "New Project" },
        ];
      case "project":
        return [
          { label: "Projects", to: { name: "projects" } },
          { label: projectName },
        ];
      case "edit-project":
        return [
          { label: "Projects", to: { name: "projects" } },
          {
            label: projectName,
            to: id ? { name: "project", params: { id } } : { name: "projects" },
          },
          { label: "Edit" },
        ];
      case "finding": {
        const ruleId =
          typeof route.params.ruleId === "string"
            ? route.params.ruleId
            : "Finding";
        return [
          { label: "Projects", to: { name: "projects" } },
          {
            label: projectName,
            to: id ? { name: "project", params: { id } } : { name: "projects" },
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
        return [{ label: route.meta.title ?? "Corsair" }];
    }
  });

  const title = computed(
    () => crumbs.value[crumbs.value.length - 1]?.label ?? "Corsair",
  );
  const nav = computed(() => route.meta.nav);

  return { title, nav, crumbs };
}
