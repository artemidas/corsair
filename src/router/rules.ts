import { RulesListView, RulesDetailView } from "@/views/rules";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/rules",
    name: "rules",
    component: RulesListView,
  },
  {
    path: "/rules/:id",
    name: "rule",
    component: RulesDetailView,
    props: true,
  },
];

export default routes;
