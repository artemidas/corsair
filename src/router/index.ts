import { createRouter, createWebHistory, type RouteRecordRaw } from "vue-router";
import Home from "@/views/Home.vue";
import ProjectDetailView from "@/views/ProjectDetailView.vue";
import BuiltInRulesView from "@/views/BuiltInRulesView.vue";
import CustomRulesView from "@/views/CustomRulesView.vue";
import RuleDetailView from "@/views/RuleDetailView.vue";

const routes: RouteRecordRaw[] = [
  { path: "/", name: "home", component: Home },

  { path: "/projects/:id", name: "project", component: ProjectDetailView, props: true },

  {
    path: "/rules/built-in",
    name: "built-in-rules",
    component: BuiltInRulesView,
  },
  {
    path: "/rules/built-in/:id",
    name: "built-in-rule",
    component: RuleDetailView,
    props: true,
    meta: { source: "built-in" },
  },

  {
    path: "/rules/custom",
    name: "custom-rules",
    component: CustomRulesView,
  },
  {
    path: "/rules/custom/:id",
    name: "custom-rule",
    component: RuleDetailView,
    props: true,
    meta: { source: "custom" },
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
