import { computed, ref } from "vue";
import { invoke } from "@tauri-apps/api/core";
import type { Finding } from "@/lib/findings";

export type ScanStatus = "completed" | "failed";

export type Scan = {
  id: string;
  projectId: string;
  status: ScanStatus;
  context: string | null;
  error: string | null;
  findingCount: number;
  startedAt: string;
  finishedAt: string;
};

export type ScanResult = {
  scan: Scan;
  findings: Finding[];
};

const scansByProject = ref<Record<string, Scan[]>>({});
const findingsByScan = ref<Record<string, Finding[]>>({});
const selectedScanIdByProject = ref<Record<string, string>>({});

export function formatScanLabel(scan: Scan): string {
  const when = new Date(scan.startedAt).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
  if (scan.status === "failed") {
    return `${when} · failed`;
  }
  const n = scan.findingCount;
  return `${when} · ${n} ${n === 1 ? "finding" : "findings"}`;
}

export function useScans() {
  const scans = computed(() => scansByProject.value);
  const selectedScanIds = computed(() => selectedScanIdByProject.value);

  function scansFor(projectId: string): Scan[] {
    return scansByProject.value[projectId] ?? [];
  }

  function selectedScanId(projectId: string): string | undefined {
    return selectedScanIdByProject.value[projectId];
  }

  function selectedScan(projectId: string): Scan | null {
    const id = selectedScanId(projectId);
    if (!id) return null;
    return scansFor(projectId).find((scan) => scan.id === id) ?? null;
  }

  function findingsFor(scanId: string): Finding[] {
    return findingsByScan.value[scanId] ?? [];
  }

  async function loadFindings(scanId: string): Promise<Finding[]> {
    const findings = await invoke<Finding[]>("list_scan_findings", { scanId });
    findingsByScan.value = { ...findingsByScan.value, [scanId]: findings };
    return findings;
  }

  async function selectScan(projectId: string, scanId: string | undefined) {
    if (!scanId) {
      const next = { ...selectedScanIdByProject.value };
      delete next[projectId];
      selectedScanIdByProject.value = next;
      return;
    }
    if (!(scanId in findingsByScan.value)) {
      await loadFindings(scanId);
    }
    selectedScanIdByProject.value = {
      ...selectedScanIdByProject.value,
      [projectId]: scanId,
    };
  }

  async function loadScans(projectId: string) {
    const list = await invoke<Scan[]>("list_scans", { projectId });
    scansByProject.value = { ...scansByProject.value, [projectId]: list };

    const current = selectedScanIdByProject.value[projectId];
    const stillThere = current && list.some((scan) => scan.id === current);
    await selectScan(projectId, stillThere ? current : list[0]?.id);
  }

  async function runScan(projectId: string): Promise<ScanResult> {
    const result = await invoke<ScanResult>("run_scan", { projectId });
    findingsByScan.value = {
      ...findingsByScan.value,
      [result.scan.id]: result.findings,
    };
    const rest = scansFor(projectId).filter((scan) => scan.id !== result.scan.id);
    scansByProject.value = {
      ...scansByProject.value,
      [projectId]: [result.scan, ...rest],
    };
    await selectScan(projectId, result.scan.id);
    return result;
  }

  return {
    scans,
    selectedScanIds,
    scansFor,
    selectedScanId,
    selectedScan,
    findingsFor,
    loadScans,
    loadFindings,
    selectScan,
    runScan,
  };
}
