import { h } from "vue";
import { ArrowUpDown } from "@lucide/vue";
import { createColumnHelper } from "@tanstack/vue-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { DataTableFeatures } from "@/components/ui/data-table";
import { type Rule } from "@/composables/useRules";
import { severityBadgeVariant } from "@/lib/severity";
import RuleRowActions from "@/components/rule/RuleRowActions.vue";
import RuleTitleCell from "@/components/rule/RuleTitleCell.vue";

const helper = createColumnHelper<DataTableFeatures, Rule>();

const SEVERITY_RANK: Record<Rule["severity"], number> = {
  critical: 0,
  high: 1,
  medium: 2,
  low: 3,
};

function sortableHeader(title: string) {
  return ({
    column,
  }: {
    column: {
      toggleSorting: (desc?: boolean) => void;
      getIsSorted: () => false | "asc" | "desc";
    };
  }) =>
    h(
      Button,
      {
        variant: "ghost",
        onClick: () => column.toggleSorting(column.getIsSorted() === "asc"),
      },
      () => [title, h(ArrowUpDown)],
    );
}

export function createRulesColumns(opts: {
  onDelete: (rule: Rule) => void;
}) {
  return helper.columns([
    helper.accessor("ruleId", {
      header: sortableHeader("Rule ID"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2 font-mono text-sm" }, row.original.ruleId),
    }),
    helper.accessor("title", {
      header: sortableHeader("Rule"),
      cell: ({ row }) => h(RuleTitleCell, { rule: row.original }),
    }),
    helper.accessor("severity", {
      header: sortableHeader("Severity"),
      sortFn: (rowA, rowB) =>
        SEVERITY_RANK[rowA.original.severity] -
        SEVERITY_RANK[rowB.original.severity],
      cell: ({ row }) =>
        h(
          Badge,
          { variant: severityBadgeVariant(row.original.severity) },
          () => row.original.severity,
        ),
    }),
    helper.accessor("resourceType", {
      header: sortableHeader("Resource"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2 font-mono text-sm" }, row.original.resourceType),
    }),
    helper.display({
      id: "actions",
      header: "Actions",
      cell: ({ row }) =>
        h(RuleRowActions, {
          rule: row.original,
          onDelete: opts.onDelete,
        }),
    }),
  ]);
}
