import { FindingDetailView } from "@/views/findings";
import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
  {
    path: "/projects/:id/scans/:scanId/findings/:ruleId",
    name: "finding",
    component: FindingDetailView,
    props: true,
    meta: {
      title: "Finding",
    },
  },
];

export default routes;
