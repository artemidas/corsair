export {};

declare module "vue-router" {
  interface RouteMeta {
    title: string;
    nav?: "home" | "clusters" | "images" | "rules" | "settings";
    projectKind?: "kubernetesClusterReview" | "containerImageReview";
  }
}