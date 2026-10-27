import { useMemo } from 'react';
import type { CellValue, TablePage } from '../lib/api';
import { displayCellValue } from '../lib/format';
import { cn } from './ui';

function cellText(cell: CellValue): string {
  if (cell.type === 'RAW') {
    return cell.value || '—';
  }
  return displayCellValue(cell.value, cell.empty);
}

function columnWidth(cell: CellValue): string {
  if (cell.type === 'RAW') return 'min-w-[220px]';
  if (cell.type === 'STRING') return 'min-w-[180px]';
  if (cell.type === 'DOUBLE') return 'min-w-[130px]';
  return 'min-w-[110px]';
}

export function TableGrid({
  page,
  selected,
  onSelect,
  searchHits
}: {
  page: TablePage;
  selected: { row: number; column: number } | null;
  onSelect: (row: number, column: number) => void;
  searchHits: Set<number>;
}) {
  const columns = useMemo(() => page.rows[0]?.cells ?? [], [page.rows]);

  if (page.rows.length === 0) {
    return (
      <div className="flex h-full items-center justify-center text-[12px] text-sub">
        This table has no rows. Increase the row capacity when creating it.
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto">
      <table className="w-full border-separate border-spacing-0 text-left">
        <caption className="sr-only">
          Table rows {page.start + 1} to {page.start + page.rows.length} of {page.total}
        </caption>
        <thead className="sticky top-0 z-20">
          <tr>
            <th
              scope="col"
              className="sticky left-0 z-30 w-14 border-b border-r border-line bg-panel px-2 py-2 text-right text-[10px] font-medium uppercase tracking-wider text-sub/70"
            >
              #
            </th>
            {columns.map((cell) => (
              <th
                key={cell.column}
                scope="col"
                className={cn(
                  'border-b border-line bg-panel px-3 py-2 align-bottom',
                  columnWidth(cell)
                )}
              >
                <div className="flex items-center justify-between gap-2">
                  <span className="truncate text-[11px] font-semibold text-ink" title={cell.name}>
                    {cell.name}
                  </span>
                  <span className="shrink-0 rounded border border-line bg-bg/40 px-1.5 py-0.5 font-mono text-[9px] text-sub">
                    {cell.type}
                  </span>
                </div>
                <div className="mt-0.5 font-mono text-[9px] text-sub/50">{cell.length} B</div>
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {page.rows.map((row) => {
            const hit = searchHits.has(row.index);
            return (
              <tr key={row.index} className="grid-row group" data-hit={hit}>
                <th
                  scope="row"
                  className={cn(
                    'sticky left-0 z-10 border-b border-r border-line/70 bg-panel-2 px-2 py-1 text-right font-mono text-[10px] font-normal text-sub/70',
                    hit && 'text-accent-2'
                  )}
                >
                  {row.index}
                </th>
                {row.cells.map((cell) => {
                  const isSelected = selected?.row === row.index && selected.column === cell.column;
                  return (
                    <td key={cell.column} className="border-b border-line/45 p-0">
                      <button
                        type="button"
                        className={cn(
                          'grid-cell flex h-8 w-full items-center overflow-hidden px-3 text-left font-mono text-[11px]',
                          cell.type === 'RAW' ? 'text-[#a7f3d0]' : 'text-ink',
                          cell.empty && 'text-sub/35'
                        )}
                        data-selected={isSelected}
                        aria-label={`Row ${row.index}, ${cell.name}, ${cell.type}, ${cellText(cell)}`}
                        aria-pressed={isSelected}
                        onClick={() => onSelect(row.index, cell.column)}
                      >
                        <span className="truncate" title={cellText(cell)}>
                          {cellText(cell)}
                        </span>
                      </button>
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
