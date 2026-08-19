import type { BadgeVariants } from "@/components/ui/badge";

export type Severity = "critical" | "high" | "medium" | "low";

const variants: Record<Severity, NonNullable<BadgeVariants["variant"]>> = {
  critical: "destructive",
  high: "default",
  medium: "secondary",
  low: "outline",
};

export function severityBadgeVariant(
  severity: Severity,
): BadgeVariants["variant"] {
  return variants[severity];
}
