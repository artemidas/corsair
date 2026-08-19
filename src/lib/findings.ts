import type { Severity } from "@/lib/severity";

export type Finding = {
  id: string;
  scanId?: string;
  ruleId: string;
  ruleTitle: string;
  severity: Severity;
  resourceKind: string;
  resourceName: string;
  namespace: string | null;
  message: string;
};

export type FindingSummary = {
  ruleId: string;
  ruleTitle: string;
  severity: Severity;
  affectedResources: number;
};

export const severityOrder: Record<Severity, number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
};

export function resourceKey(finding: Finding): string {
  return `${finding.resourceKind}/${finding.resourceName}\0${finding.namespace ?? ""}`;
}

export function summarizeFindings(findings: Finding[]): FindingSummary[] {
  const groups = new Map<string, Finding[]>();
  for (const finding of findings) {
    const list = groups.get(finding.ruleId) ?? [];
    list.push(finding);
    groups.set(finding.ruleId, list);
  }

  return [...groups.values()].map((items) => {
    const resources = new Set(items.map(resourceKey));
    const severity = items.reduce(
      (worst, item) =>
        severityOrder[item.severity] < severityOrder[worst]
          ? item.severity
          : worst,
      items[0].severity,
    );
    return {
      ruleId: items[0].ruleId,
      ruleTitle: items[0].ruleTitle,
      severity,
      affectedResources: resources.size,
    };
  });
}
