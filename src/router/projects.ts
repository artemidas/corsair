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
    redirect: { name: "cluster-projects" },
  },
  {
    path: "/projects/clusters",
    name: "cluster-projects",
    component: ProjectList,
    meta: {
      title: "Clusters",
      nav: "clusters",
      projectKind: "kubernetesClusterReview",
    },
  },
  {
    path: "/projects/images",
    name: "image-projects",
    component: ProjectList,
    meta: {
      title: "Images",
      nav: "images",
      projectKind: "containerImageReview",
    },
  },
  {
    path: "/projects/new",
    component: ProjectCreate,
    meta: {
      title: "New Project",
    },
    children: [
      {
        path: "",
        name: "new-project",
        component: ProjectKindStep,
        meta: {
          title: "New Project",
        },
      },
      {
        path: "kubernetes",
        name: "new-kubernetes-project",
        component: ProjectCreateKubernetes,
        meta: {
          title: "New Cluster",
          nav: "clusters",
        },
      },
      {
        path: "container-image",
        name: "new-container-image-project",
        component: ProjectCreateContainerImage,
        meta: {
          title: "New Engagement",
          nav: "images",
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
    },
  },
  {
    path: "/projects/:id/edit",
    name: "edit-project",
    component: ProjectEdit,
    props: true,
    meta: {
      title: "Edit Project",
    },
  },
];

export default routes;
