import { computed, ref } from "vue";
import {
  GetScan,
  ListScanFindings,
  ListScans,
  PreviewScan,
  RunScan as runScanBound,
} from "@/bindings/ladon/scan/service";
import type {
  Finding as BoundFinding,
  Scan as BoundScan,
  ScanResult as BoundResult,
} from "@/bindings/ladon/scan/models";
import type { Finding } from "@/lib/findings";
import type { Severity } from "@/lib/severity";

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

function fromScan(scan: BoundScan): Scan {
  return {
    id: scan.id,
    projectId: scan.projectId,
    status: scan.status as unknown as ScanStatus,
    context: scan.context,
    error: scan.error,
    findingCount: scan.findingCount,
    startedAt: scan.startedAt,
    finishedAt: scan.finishedAt,
  };
}

function fromFinding(finding: BoundFinding): Finding {
  return {
    id: finding.id,
    scanId: finding.scanId || undefined,
    ruleId: finding.ruleId,
    ruleTitle: finding.ruleTitle,
    severity: finding.severity as unknown as Severity,
    resourceKind: finding.resourceKind,
    resourceName: finding.resourceName,
    namespace: finding.namespace,
    message: finding.message,
  };
}

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

  async function getScan(id: string): Promise<Scan | null> {
    const scan = await GetScan(id);
    return scan ? fromScan(scan) : null;
  }

  async function loadFindings(scanId: string): Promise<Finding[]> {
    const findings = ((await ListScanFindings(scanId)) ?? []).map(fromFinding);
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
    const list = ((await ListScans(projectId)) ?? []).map(fromScan);
    scansByProject.value = { ...scansByProject.value, [projectId]: list };

    const current = selectedScanIdByProject.value[projectId];
    const stillThere = current && list.some((scan) => scan.id === current);
    await selectScan(projectId, stillThere ? current : list[0]?.id);
  }

  async function runScan(projectId: string): Promise<ScanResult> {
    const raw: BoundResult = await runScanBound(projectId);
    const result: ScanResult = {
      scan: fromScan(raw.scan),
      findings: (raw.findings ?? []).map(fromFinding),
    };
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

  async function previewScan(): Promise<Finding[]> {
    return ((await PreviewScan()) ?? []).map(fromFinding);
  }

  return {
    scans,
    selectedScanIds,
    scansFor,
    selectedScanId,
    selectedScan,
    findingsFor,
    getScan,
    loadScans,
    loadFindings,
    selectScan,
    runScan,
    previewScan,
  };
}
