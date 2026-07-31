import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";

import HomeView from "@/views/HomeView.vue";

import projectsRoutes from "./projects";
import rulesRoutes from "./rules";

const routes: RouteRecordRaw[] = [
  { 
    path: "/",
    name: "home",
    component: HomeView,
  },
  ...projectsRoutes,
  ...rulesRoutes,
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
