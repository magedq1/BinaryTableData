import { useMemo, useState, type FormEvent } from 'react';
import {
  Archive,
  ChevronLeft,
  ChevronRight,
  Database,
  FilePlus2,
  FolderOpen,
  RefreshCw,
  Search,
  X
} from 'lucide-react';
import type { CellValue, TableInfo, TablePage } from '../lib/api';
import { formatBytes, formatCount } from '../lib/format';
import { CellInspector } from './CellInspector';
import { TableGrid } from './TableGrid';
import { Badge, Button, EmptyState, IconButton, Stat, cn } from './ui';

export function TableEditor({
  info,
  page,
  pageSize,
  selected,
  cell,
  searchHits,
  busy,
  onOpen,
  onOpenCompressed,
  onCreate,
  onClose,
  onReload,
  onPageChange,
  onPageSize,
  onSelect,
  onSaveCell,
  onClearCell,
  onClearRow,
  onEmbedFile,
  onSearch,
  onClearSearch,
  onJumpToRow,
  onCloseInspector
}: {
  info: TableInfo | null;
  page: TablePage | null;
  pageSize: number;
  selected: { row: number; column: number } | null;
  cell: CellValue | null;
  searchHits: number[];
  busy: boolean;
  onOpen: () => void;
  onOpenCompressed: () => void;
  onCreate: () => void;
  onClose: () => void;
  onReload: () => void;
  onPageChange: (start: number) => void;
  onPageSize: (size: number) => void;
  onSelect: (row: number, column: number) => void;
  onSaveCell: (value: string, format: string) => Promise<void>;
  onClearCell: () => Promise<void>;
  onClearRow: () => void;
  onEmbedFile: () => void;
  onSearch: (column: number, value: string, format: string) => Promise<void>;
  onClearSearch: () => void;
  onJumpToRow: (row: number) => void;
  onCloseInspector: () => void;
}) {
  const [searchColumn, setSearchColumn] = useState<number>(-1);
  const [searchValue, setSearchValue] = useState('');
  const [searching, setSearching] = useState(false);
  const hitSet = useMemo(() => new Set(searchHits), [searchHits]);

  if (!info) {
    return (
      <EmptyState
        icon={<Database size={24} aria-hidden="true" />}
        title="Open a BinaryTable"
        body="Edit fixed-schema rows, inspect cells, search values, and embed files into RAW columns. Drop a table file anywhere in the window to open it."
        actions={
          <>
            <Button variant="primary" icon={<FolderOpen size={15} aria-hidden="true" />} onClick={onOpen}>
              Open table
            </Button>
            <Button variant="secondary" icon={<FilePlus2 size={15} aria-hidden="true" />} onClick={onCreate}>
              Create table
            </Button>
            <Button variant="ghost" icon={<Archive size={15} aria-hidden="true" />} onClick={onOpenCompressed}>
              Compressed table
            </Button>
          </>
        }
      />
    );
  }

  const activeColumn = info.columns.find((column) => column.index === searchColumn) ?? info.columns[0];
  const first = page && page.rows.length > 0 ? page.start + 1 : 0;
  const last = page ? page.start + page.rows.length : 0;
  const canPrev = info.rows > 0 && (page?.start ?? 0) > 0;
  const canNext = page ? page.start + pageSize < info.rows : false;

  async function submitSearch(event: FormEvent) {
    event.preventDefault();
    if (!activeColumn) return;
    if (!searchValue.trim()) {
      onClearSearch();
      return;
    }
    setSearching(true);
    try {
      await onSearch(activeColumn.index, searchValue, activeColumn.type === 'RAW' ? 'hex' : 'typed');
    } finally {
      setSearching(false);
    }
  }

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col">
      <div className="shrink-0 border-b border-line bg-panel/35 px-4 py-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="truncate text-[14px] font-semibold text-ink">{info.path.split(/[\\/]/).pop()}</h1>
              <Badge tone={info.readOnly ? 'warn' : 'accent'}>
                {info.compressed ? 'Compressed · read-only' : info.readOnly ? 'Read only' : 'Read/write'}
              </Badge>
            </div>
            <p className="mt-0.5 max-w-[720px] truncate font-mono text-[10px] text-sub/70" title={info.path}>
              {info.path}
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            <IconButton label="Reload rows" onClick={onReload} disabled={busy}>
              <RefreshCw size={15} aria-hidden="true" />
            </IconButton>
            <Button size="sm" variant="ghost" icon={<Archive size={14} aria-hidden="true" />} onClick={onOpenCompressed}>
              Compressed
            </Button>
            <Button size="sm" variant="secondary" icon={<FilePlus2 size={14} aria-hidden="true" />} onClick={onCreate}>
              New table
            </Button>
            <Button size="sm" variant="ghost" icon={<X size={14} aria-hidden="true" />} onClick={onClose}>
              Close
            </Button>
          </div>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-6">
          <Stat label="Rows" value={formatCount(info.rows)} detail={`row ${formatBytes(info.rowSize)}`} />
          <Stat label="Columns" value={formatCount(info.columns.length)} detail={`${info.rowSize} bytes per row`} />
          <Stat label="File" value={formatBytes(info.fileSize)} detail={`capacity ${formatBytes(info.capacity)}`} />
          <Stat label="Schema" value={formatBytes(info.schemaSize)} detail="header bytes" />
          <Stat label="Page" value={`${first}–${last}`} detail={`${formatCount(pageSize)} per page`} />
          <Stat label="Matches" value={searchHits.length ? formatCount(searchHits.length) : '—'} detail={searchHits.length ? 'rows found' : 'no active search'} />
        </div>
      </div>

      <div className="shrink-0 border-b border-line bg-panel-2/40 px-4 py-2">
        <div className="flex flex-wrap items-center gap-2">
          <form className="flex min-w-0 flex-1 items-center gap-2" onSubmit={submitSearch}>
            <div className="flex h-9 min-w-0 flex-1 items-center gap-2 rounded-lg border border-line bg-bg/60 px-2.5 focus-within:border-accent focus-within:ring-2 focus-within:ring-accent/20">
              <Search size={14} className="shrink-0 text-sub" aria-hidden="true" />
              <label htmlFor="table-search-column" className="sr-only">
                Search column
              </label>
              <select
                id="table-search-column"
                className="max-w-[150px] shrink-0 border-0 bg-transparent py-1 text-[11px] text-sub focus:outline-none"
                value={activeColumn?.index ?? 0}
                onChange={(event) => setSearchColumn(Number(event.target.value))}
              >
                {info.columns.map((column) => (
                  <option key={column.index} value={column.index}>
                    {column.name}
                  </option>
                ))}
              </select>
              <label htmlFor="table-search-value" className="sr-only">
                Search value
              </label>
              <input
                id="table-search-value"
                className="min-w-0 flex-1 border-0 bg-transparent text-[12px] text-ink placeholder:text-sub/50 focus:outline-none"
                placeholder={activeColumn?.type === 'RAW' ? 'hex bytes…' : 'Search value…'}
                value={searchValue}
                onChange={(event) => setSearchValue(event.target.value)}
              />
              {searchValue || searchHits.length > 0 ? (
                <button
                  type="button"
                  aria-label="Clear search"
                  className="rounded p-0.5 text-sub hover:text-ink"
                  onClick={() => {
                    setSearchValue('');
                    onClearSearch();
                  }}
                >
                  <X size={13} aria-hidden="true" />
                </button>
              ) : null}
            </div>
            <Button type="submit" size="sm" variant="secondary" loading={searching}>
              Find
            </Button>
          </form>

          {searchHits.length > 0 ? (
            <div className="flex items-center gap-1.5">
              <Badge tone="accent">{searchHits.length} hits</Badge>
              <Button
                size="sm"
                variant="ghost"
                icon={<ChevronRight size={14} aria-hidden="true" />}
                onClick={() => onJumpToRow(searchHits[0])}
              >
                First hit
              </Button>
            </div>
          ) : null}

          <div className="ml-auto flex items-center gap-1.5">
            <label htmlFor="page-size" className="sr-only">
              Rows per page
            </label>
            <select
              id="page-size"
              className="h-9 rounded-lg border border-line bg-bg/60 px-2 text-[11px] text-sub focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/20"
              value={pageSize}
              onChange={(event) => onPageSize(Number(event.target.value))}
            >
              {[50, 100, 250, 500].map((size) => (
                <option key={size} value={size}>
                  {size} rows
                </option>
              ))}
            </select>
            <IconButton label="Previous page" onClick={() => onPageChange(Math.max(0, (page?.start ?? 0) - pageSize))} disabled={!canPrev || busy}>
              <ChevronLeft size={15} aria-hidden="true" />
            </IconButton>
            <IconButton label="Next page" onClick={() => onPageChange((page?.start ?? 0) + pageSize)} disabled={!canNext || busy}>
              <ChevronRight size={15} aria-hidden="true" />
            </IconButton>
          </div>
        </div>

        <div className="mt-2 flex items-center gap-1.5 overflow-x-auto pb-0.5">
          {info.columns.map((column) => (
            <button
              key={column.index}
              type="button"
              className={cn(
                'shrink-0 rounded-md border px-2 py-1 font-mono text-[9px] transition-colors',
                activeColumn?.index === column.index
                  ? 'border-accent/40 bg-accent/10 text-accent-2'
                  : 'border-line bg-bg/30 text-sub hover:border-line-strong hover:text-ink'
              )}
              onClick={() => setSearchColumn(column.index)}
              title={`${column.name} · ${column.type} · ${column.length} bytes`}
            >
              {column.name}
              <span className="ml-1.5 text-sub/50">{column.type}</span>
            </button>
          ))}
        </div>
      </div>

      <div className="flex min-h-0 flex-1">
        <div className="min-w-0 flex-1 bg-bg/20">
          {page ? (
            <TableGrid page={page} selected={selected} onSelect={onSelect} searchHits={hitSet} />
          ) : (
            <div className="flex h-full items-center justify-center text-[12px] text-sub">Loading rows…</div>
          )}
        </div>
        <aside className="hidden w-[300px] shrink-0 border-l border-line bg-panel/50 lg:block 2xl:w-[360px]">
          <CellInspector
            info={info}
            selected={selected}
            cell={cell}
            busy={busy}
            readOnly={info.readOnly}
            onSave={onSaveCell}
            onClearCell={onClearCell}
            onClearRow={onClearRow}
            onEmbedFile={onEmbedFile}
            onClose={onCloseInspector}
          />
        </aside>
      </div>
    </div>
  );
}
