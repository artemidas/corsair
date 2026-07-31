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
  },
  {
    path: "/projects/new",
    name: "new-project",
    component: ProjectCreate,
  },
  {
    path: "/projects/:id",
    name: "project",
    component: ProjectDetail,
    props: true,
  },
  {
    path: "/projects/:id/edit",
    name: "edit-project",
    component: ProjectEdit,
    props: true,
  },
]

export default routes;