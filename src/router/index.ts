import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import HomeView from "@/views/HomeView.vue";

import findingsRoutes from "./findings";
import projectsRoutes from "./projects";
import rulesRoutes from "./rules";
import settingsRoutes from "./settings";

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    name: "home",
    component: HomeView,
    meta: {
      title: "Home",
      nav: "home",
    },
  },
  ...projectsRoutes,
  ...findingsRoutes,
  ...rulesRoutes,
  ...settingsRoutes,
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
