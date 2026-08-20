import { shallowRef } from "vue";
import { invoke } from "@tauri-apps/api/core";

export interface LocalImage {
  id: string;
  reference: string;
  repository: string;
  tag: string;
  size: string;
}

export interface LocalImageList {
  runtime: string;
  images: LocalImage[];
}

export function useLocalImages() {
  const images = shallowRef<LocalImage[]>([]);
  const runtime = shallowRef<string | null>(null);
  const loading = shallowRef(false);
  const loadError = shallowRef("");

  async function loadImages() {
    loading.value = true;
    loadError.value = "";
    try {
      const result = await invoke<LocalImageList>("list_local_images");
      images.value = result.images;
      runtime.value = result.runtime;
    } catch (err) {
      images.value = [];
      runtime.value = null;
      loadError.value = String(err);
    } finally {
      loading.value = false;
    }
  }

  return { images, runtime, loading, loadError, loadImages };
}
