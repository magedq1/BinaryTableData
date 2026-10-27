import {
  Archive,
  Command,
  Database,
  FilePlus2,
  FolderOpen,
  List,
  Plus,
  ShieldCheck,
  X
} from 'lucide-react';
import type { DeflateInfo, ListInfo, TableInfo } from '../lib/api';
import { baseName, dirName, formatBytes, formatCount, truncateMiddle } from '../lib/format';
import { Badge, Button, IconButton, cn } from './ui';

export type Workspace = 'table' | 'list' | 'deflate';

export interface RecentFile {
  path: string;
  kind: Workspace;
  max?: number;
  compressed?: boolean;
}

export function TopBar({
  active,
  table,
  list,
  deflate,
  onOpen,
  onCreate,
  onCommand,
  onCloseCurrent
}: {
  active: Workspace;
  table: TableInfo | null;
  list: ListInfo | null;
  deflate: DeflateInfo | null;
  onOpen: () => void;
  onCreate: () => void;
  onCommand: () => void;
  onCloseCurrent: () => void;
}) {
  const info = active === 'table' ? table : active === 'list' ? list : deflate;
  const path = info?.path ?? '';
  const title = path ? baseName(path) : active === 'table' ? 'Table workspace' : active === 'list' ? 'Binary list workspace' : 'Deflate workspace';
  const subtitle = path ? truncateMiddle(dirName(path), 72) : 'Open a file or start a new table';

  return (
    <header className="flex h-14 shrink-0 items-center justify-between gap-4 overflow-hidden border-b border-line bg-panel/85 px-4 backdrop-blur">
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-accent/30 bg-accent/10 text-accent-2 shadow-[0_0_24px_rgba(34,197,94,0.08)]">
          <Database size={18} aria-hidden="true" />
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-[13px] font-semibold tracking-tight text-ink">BinaryTable Studio</span>
            <Badge tone="accent">Local</Badge>
          </div>
          <div className="flex min-w-0 items-center gap-1.5 text-[11px] text-sub">
            <span className="truncate font-medium text-sub/90">{title}</span>
            {path ? (
              <>
                <span aria-hidden="true">·</span>
                <span className="truncate text-sub/60" title={path}>
                  {subtitle}
                </span>
              </>
            ) : null}
          </div>
        </div>
      </div>

      <div className="flex shrink-0 items-center gap-1.5">
        <Button size="sm" variant="ghost" icon={<FilePlus2 size={15} aria-hidden="true" />} onClick={onCreate}>
          New
        </Button>
        <Button size="sm" variant="secondary" icon={<FolderOpen size={15} aria-hidden="true" />} onClick={onOpen}>
          Open
        </Button>
        <Button
          size="sm"
          variant="primary"
          icon={<Command size={15} aria-hidden="true" />}
          onClick={onCommand}
          aria-keyshortcuts="Control+K Meta+K"
        >
          Command
        </Button>
        <IconButton label="Close current file" onClick={onCloseCurrent} disabled={!path}>
          <X size={16} aria-hidden="true" />
        </IconButton>
      </div>
    </header>
  );
}

export function Sidebar({
  active,
  onSelect,
  recent,
  onOpenRecent,
  onRemoveRecent,
  onOpen,
  onCreate
}: {
  active: Workspace;
  onSelect: (workspace: Workspace) => void;
  recent: RecentFile[];
  onOpenRecent: (file: RecentFile) => void;
  onRemoveRecent: (path: string) => void;
  onOpen: () => void;
  onCreate: () => void;
}) {
  const nav = [
    { id: 'table' as const, label: 'Tables', hint: 'Fixed-schema rows', icon: Database },
    { id: 'list' as const, label: 'Binary lists', hint: 'Embedded blobs', icon: List },
    { id: 'deflate' as const, label: 'Deflate', hint: 'Pack & inspect', icon: Archive }
  ];

  return (
    <aside className="flex w-[240px] shrink-0 flex-col border-r border-line bg-panel-2/80">
      <nav aria-label="Workspaces" className="space-y-1 p-3">
        {nav.map((item) => {
          const Icon = item.icon;
          const selected = active === item.id;
          return (
            <button
              key={item.id}
              type="button"
              aria-current={selected ? 'page' : undefined}
              onClick={() => onSelect(item.id)}
              className={cn(
                'group flex w-full items-center gap-3 rounded-lg border px-3 py-2.5 text-left transition-colors duration-150',
                selected
                  ? 'border-accent/25 bg-accent/10 text-ink'
                  : 'border-transparent text-sub hover:bg-muted/60 hover:text-ink'
              )}
            >
              <span
                className={cn(
                  'flex h-8 w-8 items-center justify-center rounded-lg border',
                  selected ? 'border-accent/30 bg-accent/10 text-accent-2' : 'border-line bg-bg/40 text-sub group-hover:text-ink'
                )}
              >
                <Icon size={16} aria-hidden="true" />
              </span>
              <span className="min-w-0">
                <span className="block text-[12px] font-medium">{item.label}</span>
                <span className="block truncate text-[10px] text-sub/70">{item.hint}</span>
              </span>
            </button>
          );
        })}
      </nav>

      <div className="mx-3 flex items-center gap-2">
        <Button size="sm" variant="primary" className="flex-1" icon={<Plus size={14} aria-hidden="true" />} onClick={onCreate}>
          New table
        </Button>
        <IconButton label="Open a file" onClick={onOpen}>
          <FolderOpen size={16} aria-hidden="true" />
        </IconButton>
      </div>

      <div className="mt-6 flex min-h-0 flex-1 flex-col px-3">
        <div className="flex items-center justify-between px-1 pb-2">
          <h2 className="text-[10px] font-semibold uppercase tracking-[0.16em] text-sub/70">Recent</h2>
          <span className="text-[10px] text-sub/50">{recent.length}</span>
        </div>
        <div className="min-h-0 space-y-0.5 overflow-y-auto">
          {recent.length === 0 ? (
            <p className="px-1 py-2 text-[11px] leading-5 text-sub/60">Opened files appear here.</p>
          ) : (
            recent.map((file) => (
              <div key={`${file.kind}:${file.path}`} className="group flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => onOpenRecent(file)}
                  className="min-w-0 flex-1 rounded-md px-2 py-1.5 text-left transition-colors hover:bg-muted/60"
                  title={file.path}
                >
                  <span className="block truncate text-[11px] text-sub group-hover:text-ink">{baseName(file.path)}</span>
                  <span className="block truncate text-[9px] text-sub/50">
                    {file.compressed ? `${file.kind} (compressed)` : file.kind} · {truncateMiddle(dirName(file.path), 32)}
                  </span>
                </button>
                <button
                  type="button"
                  aria-label={`Remove ${baseName(file.path)} from recent files`}
                  className="rounded-md p-1 text-sub/0 transition-colors hover:bg-muted hover:text-sub focus:text-sub/80 group-hover:text-sub/70"
                  onClick={() => onRemoveRecent(file.path)}
                >
                  <X size={12} aria-hidden="true" />
                </button>
              </div>
            ))
          )}
        </div>
      </div>

      <div className="border-t border-line/70 p-3">
        <div className="flex items-center gap-2 rounded-lg border border-line bg-bg/40 px-2.5 py-2">
          <ShieldCheck size={14} className="text-accent-2" aria-hidden="true" />
          <div>
            <p className="text-[10px] font-medium text-sub">Local-first editor</p>
            <p className="text-[9px] text-sub/60">Your files never leave this machine.</p>
          </div>
        </div>
      </div>
    </aside>
  );
}

export function StatusBar({
  active,
  table,
  list,
  deflate,
  busy
}: {
  active: Workspace;
  table: TableInfo | null;
  list: ListInfo | null;
  deflate: DeflateInfo | null;
  busy: string | null;
}) {
  let summary = 'Ready';
  if (active === 'table' && table) {
    summary = `${formatCount(table.rows)} rows · ${table.columns.length} columns · ${formatBytes(table.fileSize)}`;
  } else if (active === 'list' && list) {
    summary = `${formatCount(list.entries)} / ${formatCount(list.maxEntries)} blobs · ${formatBytes(list.fileSize)}`;
  } else if (active === 'deflate' && deflate) {
    summary = `${formatCount(deflate.blocks)} blocks · ${formatBytes(deflate.originalLength)} original`;
  }

  return (
    <footer className="flex h-8 shrink-0 items-center justify-between gap-4 border-t border-line bg-panel-2 px-4 text-[10px] text-sub/80">
      <div className="flex min-w-0 items-center gap-2">
        <span className={cn('h-1.5 w-1.5 shrink-0 rounded-full', busy ? 'animate-pulse bg-warn' : 'bg-accent')} aria-hidden="true" />
        <span className="truncate">{busy ?? summary}</span>
      </div>
      <div className="hidden items-center gap-4 md:flex">
        <span><kbd className="rounded border border-line px-1 py-0.5 font-mono text-[9px]">Ctrl K</kbd> commands</span>
        <span><kbd className="rounded border border-line px-1 py-0.5 font-mono text-[9px]">Ctrl O</kbd> open</span>
        <span><kbd className="rounded border border-line px-1 py-0.5 font-mono text-[9px]">Ctrl N</kbd> new</span>
        <span>Drop files anywhere</span>
      </div>
    </footer>
  );
}
