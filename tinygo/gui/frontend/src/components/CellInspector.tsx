import { useEffect, useMemo, useState } from 'react';
import { Braces, Eraser, FileUp, Save, Trash2, X } from 'lucide-react';
import type { CellValue, ColumnInfo, TableInfo } from '../lib/api';
import { formatBytes } from '../lib/format';
import { Badge, Button, Field, IconButton, inputClass, textareaClass, cn } from './ui';

export function CellInspector({
  info,
  selected,
  cell,
  busy,
  readOnly,
  onSave,
  onClearCell,
  onClearRow,
  onEmbedFile,
  onClose
}: {
  info: TableInfo;
  selected: { row: number; column: number } | null;
  cell: CellValue | null;
  busy: boolean;
  readOnly: boolean;
  onSave: (value: string, format: string) => Promise<void>;
  onClearCell: () => Promise<void>;
  onClearRow: () => void;
  onEmbedFile: () => void;
  onClose: () => void;
}) {
  const column = useMemo<ColumnInfo | null>(() => {
    if (!selected) return null;
    return info.columns.find((entry) => entry.index === selected.column) ?? null;
  }, [info.columns, selected]);
  const [format, setFormat] = useState<'typed' | 'hex' | 'text'>('typed');
  const [draft, setDraft] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!cell) return;
    setError('');
    if (cell.type === 'RAW') {
      setFormat('hex');
      setDraft(cell.rawHex);
    } else if (cell.type === 'STRING') {
      setFormat('typed');
      setDraft(cell.value);
    } else {
      setFormat('typed');
      setDraft(cell.value);
    }
  }, [cell]);

  if (!selected || !column || !cell) {
    return (
      <div className="flex h-full flex-col">
        <div className="border-b border-line px-4 py-3">
          <h2 className="text-[12px] font-semibold text-ink">Inspector</h2>
          <p className="mt-0.5 text-[10px] text-sub">Select a cell to edit its value.</p>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto p-4">
          <div className="rounded-lg border border-line bg-panel-2/50 p-3">
            <div className="flex items-center gap-2 text-[11px] font-medium text-sub">
              <Braces size={14} aria-hidden="true" />
              Schema
            </div>
            <pre className="mt-2 overflow-x-auto whitespace-pre-wrap font-mono text-[10px] leading-5 text-sub/80">
              {info.schema}
            </pre>
          </div>
          <div className="mt-4 space-y-2">
            {info.columns.map((entry) => (
              <div key={entry.index} className="flex items-center justify-between gap-2 rounded-md border border-line/60 bg-bg/30 px-2.5 py-1.5">
                <div className="min-w-0">
                  <p className="truncate text-[11px] text-ink">{entry.name}</p>
                  <p className="font-mono text-[9px] text-sub/60">offset {entry.startPos}</p>
                </div>
                <div className="flex shrink-0 items-center gap-1">
                  <Badge tone="muted">{entry.type}</Badge>
                  <span className="font-mono text-[9px] text-sub">{entry.length} B</span>
                </div>
              </div>
            ))}
          </div>
          <p className="mt-4 text-[10px] leading-5 text-sub/60">
            Tip: use the arrow keys to move the selection and Ctrl+K for commands.
          </p>
        </div>
      </div>
    );
  }

  const activeCell = cell;
  const stringBytes = new TextEncoder().encode(draft).length;
  const rawBytes = Math.floor(draft.replace(/\s/g, '').replace(/^0x/i, '').length / 2);

  function validate(): string {
    if (format === 'hex') {
      const clean = draft.replace(/\s/g, '').replace(/^0x/i, '');
      if (clean.length % 2 !== 0) return 'Hex input needs an even number of characters.';
      if (!/^[0-9a-fA-F]*$/.test(clean)) return 'Hex input contains invalid characters.';
      const bytes = clean.length / 2;
      if (bytes > activeCell.length) return `This column holds ${activeCell.length} bytes; the value needs ${bytes}.`;
      return '';
    }
    if (activeCell.type === 'STRING' && stringBytes > activeCell.length) {
      return `This column holds ${activeCell.length} bytes; the UTF-8 value needs ${stringBytes}.`;
    }
    if (format === 'text') {
      const bytes = new TextEncoder().encode(draft).length;
      if (bytes > activeCell.length) {
        return `This column holds ${activeCell.length} bytes; the value needs ${bytes}.`;
      }
    }
    return '';
  }

  async function submit() {
    const message = validate();
    if (message) {
      setError(message);
      return;
    }
    setError('');
    await onSave(draft, format);
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-start justify-between gap-2 border-b border-line px-4 py-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <h2 className="truncate text-[12px] font-semibold text-ink">{cell.name}</h2>
            <Badge tone="accent">{cell.type}</Badge>
          </div>
          <p className="mt-0.5 font-mono text-[10px] text-sub">
            row {selected.row} · column {selected.column} · {cell.length} bytes
          </p>
        </div>
        <IconButton label="Clear selection" onClick={onClose}>
          <X size={15} aria-hidden="true" />
        </IconButton>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {readOnly ? (
          <p className="mb-3 rounded-lg border border-warn/30 bg-warn/10 px-3 py-2 text-[10px] leading-5 text-[#fcd34d]">
            This is a read-only view. Compressed tables are decoded through DeflateBinaryStreamReader and cannot be edited.
          </p>
        ) : null}
        {cell.type === 'RAW' ? (
          <div className="mb-3 flex rounded-lg border border-line bg-bg/40 p-1" role="group" aria-label="RAW value mode">
            <button
              type="button"
              className={cn(
                'flex-1 rounded-md px-2 py-1.5 text-[11px] transition-colors',
                format === 'hex' ? 'bg-muted text-ink' : 'text-sub hover:text-ink'
              )}
              aria-pressed={format === 'hex'}
              onClick={() => {
                if (format === 'text') setDraft(cell.rawHex);
                setFormat('hex');
              }}
            >
              Hex
            </button>
            <button
              type="button"
              className={cn(
                'flex-1 rounded-md px-2 py-1.5 text-[11px] transition-colors',
                format === 'text' ? 'bg-muted text-ink' : 'text-sub hover:text-ink'
              )}
              aria-pressed={format === 'text'}
              onClick={() => {
                if (format === 'hex') setDraft(hexToText(cell.rawHex));
                setFormat('text');
              }}
            >
              Text
            </button>
          </div>
        ) : null}

        <form
          onSubmit={(event) => {
            event.preventDefault();
            void submit();
          }}
        >
          <Field
            label={format === 'hex' ? 'Raw bytes (hex)' : cell.type === 'RAW' ? 'Text bytes' : 'Value'}
            hint={
              format === 'hex'
                ? `${rawBytes} / ${cell.length} bytes`
                : cell.type === 'STRING'
                  ? `${stringBytes} / ${cell.length} UTF-8 bytes`
                  : `Stored in ${cell.length} bytes`
            }
            error={error}
            htmlFor="cell-value"
          >
            {cell.type === 'STRING' || format === 'hex' || format === 'text' ? (
              <textarea
                id="cell-value"
                data-autofocus
                className={textareaClass}
                rows={cell.type === 'RAW' ? 7 : 4}
                spellCheck={false}
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                aria-invalid={Boolean(error)}
                disabled={readOnly}
              />
            ) : (
              <input
                id="cell-value"
                data-autofocus
                className={inputClass}
                inputMode={cell.type === 'DOUBLE' ? 'decimal' : 'numeric'}
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                aria-invalid={Boolean(error)}
                disabled={readOnly}
              />
            )}
          </Field>

          <div className="mt-3 flex flex-wrap gap-2">
            <Button type="submit" variant="primary" size="sm" icon={<Save size={14} aria-hidden="true" />} loading={busy} disabled={readOnly}>
              Save cell
            </Button>
            <Button
              size="sm"
              variant="secondary"
              icon={<Eraser size={14} aria-hidden="true" />}
              onClick={() => void onClearCell()}
              disabled={busy || readOnly}
            >
              Clear cell
            </Button>
            {cell.type === 'RAW' ? (
              <Button size="sm" variant="secondary" icon={<FileUp size={14} aria-hidden="true" />} onClick={onEmbedFile} disabled={busy || readOnly}>
                Embed file
              </Button>
            ) : null}
          </div>
        </form>

        <div className="mt-5 rounded-lg border border-line bg-panel-2/50 p-3">
          <div className="flex items-center justify-between">
            <span className="text-[10px] font-medium uppercase tracking-wider text-sub/70">Raw hex</span>
            <span className="font-mono text-[9px] text-sub/60">{formatBytes(cell.length)}</span>
          </div>
          <pre className="mt-2 max-h-28 overflow-auto whitespace-pre-wrap break-all font-mono text-[10px] leading-5 text-sub/80">
            {cell.rawHex}
          </pre>
        </div>

        <div className="mt-4 rounded-lg border border-danger/20 bg-danger/5 p-3">
          <p className="text-[10px] leading-5 text-sub/80">
            Clearing a row writes zero bytes to every column. The row index and table capacity stay unchanged.
          </p>
          <Button
            size="sm"
            variant="danger"
            className="mt-2"
            icon={<Trash2 size={14} aria-hidden="true" />}
            onClick={onClearRow}
            disabled={busy || readOnly}
          >
            Clear row {selected.row}
          </Button>
        </div>
      </div>
    </div>
  );
}

function hexToText(hex: string): string {
  const clean = hex.replace(/\s/g, '');
  const bytes = new Uint8Array(clean.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = Number.parseInt(clean.slice(i * 2, i * 2 + 2), 16);
  }
  try {
    return new TextDecoder().decode(bytes);
  } catch {
    return '';
  }
}
