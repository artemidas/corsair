import type { RouteRecordRaw } from "vue-router";

import { ProjectKindStep } from "@/components/form";
import {
  ProjectCreate,
  ProjectCreateKubernetes,
  ProjectCreateContainerImage,
  ProjectList,
  ProjectDetail,
  ProjectEdit,
} from "@/views/project";

const routes: RouteRecordRaw[] = [
  {
    path: "/projects",
    name: "projects",
    component: ProjectList,
    meta: {
      title: "Projects",
      nav: "projects",
    },
  },
  {
    path: "/projects/new",
    component: ProjectCreate,
    meta: {
      title: "New Project",
      nav: "projects",
    },
    children: [
      {
        path: "",
        name: "new-project",
        component: ProjectKindStep,
        meta: {
          title: "New Project",
          nav: "projects",
        },
      },
      {
        path: "kubernetes",
        name: "new-kubernetes-project",
        component: ProjectCreateKubernetes,
        meta: {
          title: "New Kubernetes Project",
          nav: "projects",
        },
      },
      {
        path: "container-image",
        name: "new-container-image-project",
        component: ProjectCreateContainerImage,
        meta: {
          title: "New Container Image Project",
          nav: "projects",
        },
      },
    ],
  },
  {
    path: "/projects/:id",
    name: "project",
    component: ProjectDetail,
    props: true,
    meta: {
      title: "Project",
      nav: "projects",
    },
  },
  {
    path: "/projects/:id/edit",
    name: "edit-project",
    component: ProjectEdit,
    props: true,
    meta: {
      title: "Edit Project",
      nav: "projects",
    },
  },
];

export default routes;
