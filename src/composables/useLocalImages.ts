import { shallowRef } from "vue";
import { invoke } from "@tauri-apps/api/core";

export interface LocalImage {
  id: string;
  reference: string;
  repository: string;
  tag: string;
  size: string;
  sizeBytes: number | null;
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

export function formatBytes(n: number): string {
  const kb = 1024;
  const mb = 1024 * kb;
  const gb = 1024 * mb;
  if (n >= gb) return `${(n / gb).toFixed(1)} GB`;
  if (n >= mb) return `${(n / mb).toFixed(1)} MB`;
  if (n >= kb) return `${(n / kb).toFixed(1)} KB`;
  return `${Math.round(n)} B`;
}

export function normalizeImageRef(reference: string): string {
  return reference
    .trim()
    .replace(/^docker\.io\//, "")
    .replace(/^library\//, "");
}

export function matchLocalImage(
  reference: string,
  locals: LocalImage[],
): LocalImage | undefined {
  const exact = locals.find((image) => image.reference === reference);
  if (exact) return exact;
  const normalized = normalizeImageRef(reference);
  return locals.find(
    (image) => normalizeImageRef(image.reference) === normalized,
  );
}

function shortId(id: string): string {
  return id.replace(/^sha256:/, "").slice(0, 12);
}

export function imageRowStats(
  reference: string,
  locals: LocalImage[],
): {
  reference: string;
  size: string;
  sizeBytes: number | null;
  id: string;
} {
  const local = matchLocalImage(reference, locals);
  return {
    reference,
    size: local?.size ?? "",
    sizeBytes: local?.sizeBytes ?? null,
    id: local ? shortId(local.id) : "",
  };
}
