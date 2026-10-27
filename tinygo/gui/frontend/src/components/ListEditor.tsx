import { useRef, useState, type DragEvent } from 'react';
import {
  Archive,
  Download,
  FilePlus2,
  FileUp,
  FolderOpen,
  HardDrive,
  List,
  RefreshCw,
  Replace,
  Trash2,
  X
} from 'lucide-react';
import type { BlobData, ListInfo } from '../lib/api';
import { formatBytes, formatCount, percent } from '../lib/format';
import { Badge, Button, EmptyState, IconButton, Stat, cn } from './ui';

export function ListEditor({
  info,
  selectedIndex,
  blob,
  busy,
  onOpen,
  onOpenCompressed,
  onCreate,
  onClose,
  onRefresh,
  onAddFiles,
  onSelect,
  onExport,
  onReplace,
  onDelete
}: {
  info: ListInfo | null;
  selectedIndex: number | null;
  blob: BlobData | null;
  busy: boolean;
  onOpen: () => void;
  onOpenCompressed: () => void;
  onCreate: () => void;
  onClose: () => void;
  onRefresh: () => void;
  onAddFiles: () => void;
  onSelect: (index: number) => void;
  onExport: (index: number) => void;
  onReplace: (index: number) => void;
  onDelete: (index: number) => void;
}) {
  const [dragging, setDragging] = useState(false);
  const dragDepth = useRef(0);

  function dragEnter(event: DragEvent) {
    event.preventDefault();
    dragDepth.current++;
    setDragging(true);
  }
  function dragLeave(event: DragEvent) {
    event.preventDefault();
    dragDepth.current = Math.max(0, dragDepth.current - 1);
    if (dragDepth.current === 0) setDragging(false);
  }
  function dragOver(event: DragEvent) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }
  function drop(event: DragEvent) {
    event.preventDefault();
    dragDepth.current = 0;
    setDragging(false);
    // Wails resolves absolute paths and delivers them through the window-level
    // file-drop callback; opening a second picker here would be wrong.
  }

  if (!info) {
    return (
      <EmptyState
        icon={<List size={24} aria-hidden="true" />}
        title="Open a binary list"
        body="Store variable-length blobs next to your tables. Drop files into the window or use the picker to embed images, documents, and other assets."
        actions={
          <>
            <Button variant="primary" icon={<FolderOpen size={15} aria-hidden="true" />} onClick={onOpen}>
              Open list
            </Button>
            <Button variant="secondary" icon={<FilePlus2 size={15} aria-hidden="true" />} onClick={onCreate}>
              New list
            </Button>
            <Button variant="ghost" icon={<Archive size={15} aria-hidden="true" />} onClick={onOpenCompressed}>
              Compressed list
            </Button>
          </>
        }
      />
    );
  }

  const usage = info.maxEntries > 0 ? info.entries / info.maxEntries : 0;

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col">
      <div className="shrink-0 border-b border-line bg-panel/35 px-4 py-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="truncate text-[14px] font-semibold text-ink">{info.path.split(/[\\/]/).pop()}</h1>
              <Badge tone={info.readOnly ? 'warn' : 'accent'}>
                {info.compressed ? 'Compressed · read-only' : info.readOnly ? 'Read only' : 'Binary list'}
              </Badge>
            </div>
            <p className="mt-0.5 max-w-[720px] truncate font-mono text-[10px] text-sub/70" title={info.path}>
              {info.path}
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            <IconButton label="Reload list" onClick={onRefresh} disabled={busy}>
              <RefreshCw size={15} aria-hidden="true" />
            </IconButton>
            <Button size="sm" variant="ghost" icon={<Archive size={14} aria-hidden="true" />} onClick={onOpenCompressed}>
              Compressed
            </Button>
            <Button size="sm" variant="secondary" icon={<FilePlus2 size={14} aria-hidden="true" />} onClick={onCreate}>
              New list
            </Button>
            <Button size="sm" variant="ghost" icon={<X size={14} aria-hidden="true" />} onClick={onClose}>
              Close
            </Button>
          </div>
        </div>

        <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
          <Stat label="Entries" value={`${formatCount(info.entries)} / ${formatCount(info.maxEntries)}`} detail={`${percent(usage)} capacity used`} />
          <Stat label="File" value={formatBytes(info.fileSize)} detail={`index ${formatBytes(info.indexSize)}`} />
          <Stat label="Schema" value={`${info.schemaSize} B`} detail="fixed prefix" />
          <Stat label="Selected" value={selectedIndex === null ? '—' : `#${selectedIndex}`} detail={blob ? formatBytes(blob.size) : 'no blob selected'} />
        </div>

        <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-muted">
          <div
            className={cn('h-full rounded-full transition-all duration-300', usage > 0.9 ? 'bg-warn' : 'bg-accent')}
            style={{ width: `${Math.min(100, Math.max(0, usage * 100))}%` }}
            role="progressbar"
            aria-valuemin={0}
            aria-valuemax={info.maxEntries}
            aria-valuenow={info.entries}
            aria-label="List capacity used"
          />
        </div>
      </div>

      <div className="flex min-h-0 flex-1">
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="p-4">
            <div
              className={cn(
                'drop-target flex flex-col items-center justify-center rounded-xl border border-dashed border-line-strong bg-panel-2/40 px-4 py-5 text-center',
                dragging && 'border-accent bg-accent/5'
              )}
              data-dragging={dragging}
              onDragEnter={dragEnter}
              onDragLeave={dragLeave}
              onDragOver={dragOver}
              onDrop={drop}
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-line bg-bg/40 text-accent-2">
                <HardDrive size={18} aria-hidden="true" />
              </div>
              <p className="mt-3 text-[12px] font-medium text-ink">
                {info.readOnly ? 'Read-only compressed list' : 'Embed files into this list'}
              </p>
              <p className="mt-0.5 text-[10px] text-sub">
                {info.readOnly
                  ? 'This list is decoded through DeflateBinaryStreamReader. Export and preview work; editing is disabled.'
                  : 'Drop files anywhere in the window, or pick them explicitly.'}
              </p>
              <div className="mt-3 flex flex-wrap justify-center gap-2">
                <Button size="sm" variant="primary" icon={<FileUp size={14} aria-hidden="true" />} onClick={onAddFiles} loading={busy} disabled={info.readOnly}>
                  Add files
                </Button>
                <span className="self-center text-[10px] text-sub/60">
                  {info.readOnly ? 'Decompress on the fly' : 'Files are copied into the list'}
                </span>
              </div>
            </div>
          </div>

          <div className="min-h-0 flex-1 overflow-auto px-4 pb-4">
            {info.blobs.length === 0 ? (
              <div className="flex h-40 items-center justify-center rounded-lg border border-line bg-bg/20 text-[12px] text-sub">
                This list is empty. Add a file to create the first blob.
              </div>
            ) : (
              <table className="w-full border-separate border-spacing-0 text-left">
                <caption className="sr-only">Binary list entries</caption>
                <thead className="sticky top-0 z-10">
                  <tr>
                    <th scope="col" className="border-b border-line bg-panel px-3 py-2 text-[10px] font-medium uppercase tracking-wider text-sub/70">
                      Index
                    </th>
                    <th scope="col" className="border-b border-line bg-panel px-3 py-2 text-[10px] font-medium uppercase tracking-wider text-sub/70">
                      Size
                    </th>
                    <th scope="col" className="border-b border-line bg-panel px-3 py-2 text-[10px] font-medium uppercase tracking-wider text-sub/70">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {info.blobs.map((entry) => {
                    const selected = selectedIndex === entry.index;
                    return (
                      <tr key={entry.index} className={cn('group', selected && 'bg-accent/5')}>
                        <th scope="row" className="border-b border-line/50 px-3 py-1.5 text-left font-mono text-[11px] font-normal">
                          <button
                            type="button"
                            className={cn('rounded px-1 py-0.5 text-left transition-colors hover:text-accent-2', selected ? 'text-accent-2' : 'text-ink')}
                            onClick={() => onSelect(entry.index)}
                            aria-pressed={selected}
                          >
                            #{entry.index}
                          </button>
                        </th>
                        <td className="border-b border-line/50 px-3 py-1.5 font-mono text-[11px] text-sub">
                          {formatBytes(entry.size)}
                          {entry.name ? <span className="ml-2 text-[10px] text-sub/60">{entry.name}</span> : null}
                        </td>
                        <td className="border-b border-line/50 px-3 py-1.5">
                          <div className="flex items-center gap-1 opacity-70 transition-opacity group-hover:opacity-100">
                            <IconButton label={`Preview blob ${entry.index}`} onClick={() => onSelect(entry.index)}>
                              <List size={14} aria-hidden="true" />
                            </IconButton>
                            <IconButton label={`Export blob ${entry.index}`} onClick={() => onExport(entry.index)}>
                              <Download size={14} aria-hidden="true" />
                            </IconButton>
                            <IconButton label={`Replace blob ${entry.index}`} onClick={() => onReplace(entry.index)} disabled={info.readOnly}>
                              <Replace size={14} aria-hidden="true" />
                            </IconButton>
                            <IconButton label={`Delete blob ${entry.index}`} onClick={() => onDelete(entry.index)} disabled={info.readOnly}>
                              <Trash2 size={14} aria-hidden="true" />
                            </IconButton>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </div>
        </div>

        <aside className="hidden w-[340px] shrink-0 border-l border-line bg-panel/50 lg:block 2xl:w-[380px]">
          <BlobPreview
            info={info}
            blob={blob}
            selectedIndex={selectedIndex}
            busy={busy}
            readOnly={info.readOnly}
            onExport={onExport}
            onReplace={onReplace}
            onDelete={onDelete}
          />
        </aside>
      </div>
    </div>
  );
}

function BlobPreview({
  info,
  blob,
  selectedIndex,
  busy,
  readOnly,
  onExport,
  onReplace,
  onDelete
}: {
  info: ListInfo;
  blob: BlobData | null;
  selectedIndex: number | null;
  busy: boolean;
  readOnly: boolean;
  onExport: (index: number) => void;
  onReplace: (index: number) => void;
  onDelete: (index: number) => void;
}) {
  if (selectedIndex === null || !blob) {
    return (
      <div className="flex h-full flex-col">
        <div className="border-b border-line px-4 py-3">
          <h2 className="text-[12px] font-semibold text-ink">Blob preview</h2>
          <p className="mt-0.5 text-[10px] text-sub">Select an entry to inspect its bytes.</p>
        </div>
        <div className="flex flex-1 items-center justify-center p-6 text-center">
          <div>
            <List size={28} className="mx-auto text-sub/50" aria-hidden="true" />
            <p className="mt-3 text-[11px] leading-5 text-sub/70">
              Images and text render inline. Binary blobs show a hex preview and can be exported.
            </p>
          </div>
        </div>
      </div>
    );
  }

  const isImage = blob.mime.startsWith('image/') && Boolean(blob.base64);
  return (
    <div className="flex h-full flex-col">
      <div className="flex items-start justify-between gap-2 border-b border-line px-4 py-3">
        <div className="min-w-0">
          <h2 className="text-[12px] font-semibold text-ink">Blob #{blob.index}</h2>
          <p className="mt-0.5 font-mono text-[10px] text-sub">{formatBytes(blob.size)} · {blob.mime}</p>
        </div>
        <Badge tone="muted">{info.maxEntries} max</Badge>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {isImage ? (
          <div className="overflow-hidden rounded-xl border border-line bg-[#0b1120] p-2">
            <img
              src={`data:${blob.mime};base64,${blob.base64}`}
              alt={`Preview of blob ${blob.index}`}
              className="mx-auto max-h-64 rounded-lg object-contain"
            />
          </div>
        ) : blob.isText ? (
          <div className="rounded-xl border border-line bg-bg/50">
            <div className="flex items-center justify-between border-b border-line px-3 py-2">
              <span className="text-[10px] font-medium uppercase tracking-wider text-sub/70">Text preview</span>
              {blob.truncated ? <Badge tone="warn">Truncated</Badge> : null}
            </div>
            <pre className="max-h-80 overflow-auto whitespace-pre-wrap break-words p-3 font-mono text-[11px] leading-5 text-sub">
              {blob.text || '(empty)'}
            </pre>
          </div>
        ) : (
          <div className="rounded-xl border border-line bg-bg/50">
            <div className="flex items-center justify-between border-b border-line px-3 py-2">
              <span className="text-[10px] font-medium uppercase tracking-wider text-sub/70">Hex preview</span>
              <span className="font-mono text-[9px] text-sub/60">{blob.hex.length / 2} bytes shown</span>
            </div>
            <pre className="max-h-80 overflow-auto whitespace-pre-wrap break-all p-3 font-mono text-[10px] leading-5 text-[#a7f3d0]">
              {blob.hex || '(empty)'}
            </pre>
          </div>
        )}

        <div className="mt-4 flex flex-wrap gap-2">
          <Button size="sm" variant="primary" icon={<Download size={14} aria-hidden="true" />} onClick={() => onExport(blob.index)} disabled={busy}>
            Export
          </Button>
          <Button size="sm" variant="secondary" icon={<Replace size={14} aria-hidden="true" />} onClick={() => onReplace(blob.index)} disabled={busy || readOnly}>
            Replace
          </Button>
          <Button size="sm" variant="danger" icon={<Trash2 size={14} aria-hidden="true" />} onClick={() => onDelete(blob.index)} disabled={busy || readOnly}>
            Delete
          </Button>
        </div>

        <p className="mt-4 text-[10px] leading-5 text-sub/60">
          Blob offsets are 32-bit and names are not stored in the Kotlin format. Export uses a generated file name.
        </p>
      </div>
    </div>
  );
}
