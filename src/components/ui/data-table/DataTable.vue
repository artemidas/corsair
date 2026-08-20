<script setup lang="ts" generic="TData extends RowData">
import type {
  ColumnDef,
  ExpandedState,
  GroupingState,
  Row,
  RowData,
  SortingState,
} from "@tanstack/vue-table"
import { computed } from "vue"
import { FlexRender, useTable } from "@tanstack/vue-table"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { cn } from "@/lib/utils"
import { features, type DataTableFeatures } from "./features"

const props = withDefaults(defineProps<{
  columns: ColumnDef<DataTableFeatures, TData>[]
  data: TData[]
  getRowId?: (originalRow: TData, index: number) => string
  initialSorting?: SortingState
  initialGrouping?: GroupingState
  initialExpanded?: ExpandedState
  rowClick?: (row: TData) => void
}>(), {
  initialSorting: () => [],
  initialGrouping: () => [],
  initialExpanded: () => ({}),
})

const table = useTable({
  features,
  get data() { return props.data },
  get columns() { return props.columns },
  getRowId: props.getRowId,
  initialState: {
    sorting: props.initialSorting,
    grouping: props.initialGrouping,
    expanded: props.initialExpanded,
  },
})

const pagination = computed(() => table.atoms.pagination.get())
const pageCount = computed(() => table.getPageCount())
const showPagination = computed(() => pageCount.value > 1)

function groupedCell(row: Row<DataTableFeatures, TData>) {
  return row.getVisibleCells().find((cell) => cell.getIsGrouped())
}

function handleRowClick(row: Row<DataTableFeatures, TData>) {
  if (row.getCanExpand()) {
    row.getToggleExpandedHandler()()
    return
  }
  props.rowClick?.(row.original)
}
</script>

<template>
  <div class="flex flex-col gap-4">
    <div class="overflow-hidden rounded-md border">
      <Table>
        <TableHeader>
          <TableRow
            v-for="headerGroup in table.getHeaderGroups()"
            :key="headerGroup.id"
          >
            <TableHead
              v-for="header in headerGroup.headers"
              :key="header.id"
            >
              <FlexRender
                v-if="!header.isPlaceholder"
                :header="header"
              />
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <template v-if="table.getRowModel().rows.length">
            <TableRow
              v-for="row in table.getRowModel().rows"
              :key="row.id"
              :class="cn(
                row.getIsGrouped() && 'bg-muted/50',
                'cursor-pointer',
              )"
              @click="handleRowClick(row)"
            >
              <template v-if="row.getIsGrouped() && groupedCell(row)">
                <TableCell :colspan="columns.length">
                  <FlexRender :cell="groupedCell(row)" />
                </TableCell>
              </template>
              <template v-else>
                <TableCell
                  v-for="cell in row.getVisibleCells()"
                  :key="cell.id"
                >
                  <FlexRender :cell="cell" />
                </TableCell>
              </template>
            </TableRow>
          </template>
          <TableEmpty v-else :colspan="columns.length">
            No results.
          </TableEmpty>
        </TableBody>
      </Table>
    </div>

    <div
      v-if="showPagination"
      class="flex items-center justify-end gap-2"
    >
      <div class="text-sm text-muted-foreground">
        Page {{ pagination.pageIndex + 1 }} of {{ pageCount }}
      </div>
      <Button
        variant="outline"
        size="sm"
        :disabled="!table.getCanPreviousPage()"
        @click="table.previousPage()"
      >
        Previous
      </Button>
      <Button
        variant="outline"
        size="sm"
        :disabled="!table.getCanNextPage()"
        @click="table.nextPage()"
      >
        Next
      </Button>
    </div>
  </div>
</template>
