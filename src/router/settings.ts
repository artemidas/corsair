import type { RouteRecordRaw } from "vue-router";

import { SettingsView } from "@/views/settings";

const routes: RouteRecordRaw[] = [
  {
    path: "/settings",
    name: "settings",
    component: SettingsView,
    meta: {
      title: "Settings",
      nav: "settings",
    },
  },
];

export default routes;
