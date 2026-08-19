import type { RouteRecordRaw } from "vue-router";

import {
  ProjectCreate,
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
    name: "new-project",
    component: ProjectCreate,
    meta: {
      title: "New Project",
      nav: "projects",
    },
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
]

export default routes;