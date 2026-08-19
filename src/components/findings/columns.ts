import { h } from "vue";
import { ArrowUpDown } from "@lucide/vue";
import { createColumnHelper } from "@tanstack/vue-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { DataTableFeatures } from "@/components/ui/data-table";
import {
  severityOrder,
  type Finding,
  type FindingSummary,
} from "@/lib/findings";
import { severityBadgeVariant, type Severity } from "@/lib/severity";
import FindingsDetailsLink from "./FindingsDetailsLink.vue";

const summaryHelper = createColumnHelper<DataTableFeatures, FindingSummary>();
const findingHelper = createColumnHelper<DataTableFeatures, Finding>();

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

export function createFindingsSummaryColumns(projectId: string, scanId: string) {
  return summaryHelper.columns([
    summaryHelper.accessor("ruleId", {
      header: sortableHeader("Rule ID"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2" }, row.original.ruleId),
    }),
    summaryHelper.accessor("ruleTitle", {
      header: sortableHeader("Rule"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2" }, row.original.ruleTitle),
    }),
    summaryHelper.accessor("affectedResources", {
      header: sortableHeader("Affected resources"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2" }, row.original.affectedResources),
    }),
    summaryHelper.display({
      id: "details",
      header: "",
      cell: ({ row }) =>
        h(FindingsDetailsLink, {
          projectId,
          scanId,
          ruleId: row.original.ruleId,
        }),
    }),
  ]);
}

export function createRuleFindingsColumns(includeMessage = true) {
  return findingHelper.columns([
    findingHelper.accessor((row) => `${row.resourceKind}/${row.resourceName}`, {
      id: "resource",
      header: sortableHeader("Resource"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2" }, `${row.original.resourceKind}/${row.original.resourceName}`),
    }),
    findingHelper.accessor((row) => row.namespace ?? "-", {
      id: "namespace",
      header: sortableHeader("Namespace"),
      cell: ({ row }) =>
        h("span", { class: "px-3 py-2" }, row.original.namespace ?? "-")
    }),
    
  ]);
}

export const ruleFindingsColumns = createRuleFindingsColumns();
export const ruleResourceColumns = createRuleFindingsColumns(false);
