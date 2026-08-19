export {};

declare module "vue-router" {
  interface RouteMeta {
    title: string;
    nav: "home" | "projects" | "rules" | "settings";
  }
}