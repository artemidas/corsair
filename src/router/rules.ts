import { RulesListView, RulesDetailView } from "@/views/rules";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/rules",
    name: "rules",
    component: RulesListView,
    meta: {
      title: "Rules",
      nav: "rules",
    },
  },
  {
    path: "/rules/:id",
    name: "rule",
    component: RulesDetailView,
    props: true,
    meta: {
      title: "Rule",
      nav: "rules",
    },
  },
];

export default routes;
