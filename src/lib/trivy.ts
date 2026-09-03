export type TrivyScanOptions = {
  scanners: string[];
  severity: string[];
  ignoreUnfixed: boolean;
  skipDbUpdate: boolean;
  extraArgs: string[];
};

export const defaultTrivyScanOptions = (): TrivyScanOptions => ({
  scanners: ["vuln"],
  severity: [],
  ignoreUnfixed: false,
  skipDbUpdate: true,
  extraArgs: [],
});

export const trivyScanners = [
  { id: "vuln", label: "Vulnerabilities" },
  { id: "misconfig", label: "Misconfigurations" },
  { id: "secret", label: "Secrets" },
  { id: "license", label: "Licenses" },
] as const;

export const trivySeverities = [
  "CRITICAL",
  "HIGH",
  "MEDIUM",
  "LOW",
  "UNKNOWN",
] as const;

export function parseExtraArgs(text: string): string[] {
  return text
    .trim()
    .split(/\s+/)
    .map((part) => part.trim())
    .filter(Boolean);
}

export function toBoundTrivyOptions(options: TrivyScanOptions) {
  return {
    scanners: options.scanners,
    severity: options.severity,
    ignoreUnfixed: options.ignoreUnfixed,
    skipDbUpdate: options.skipDbUpdate,
    extraArgs: options.extraArgs,
  };
}
