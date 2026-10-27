import { useState, type FormEvent } from 'react';
import {
  Archive,
  ArrowDownToLine,
  ArrowUpFromLine,
  Database,
  FileSearch,
  FolderOpen,
  List,
  RefreshCw,
  X
} from 'lucide-react';
import type { DeflateInfo, FileFilter } from '../lib/api';
import { formatBytes, formatCount, percent } from '../lib/format';
import { Badge, Button, Field, IconButton, Stat, inputClass } from './ui';

export function DeflatePanel({
  info,
  busy,
  onPickFile,
  onPickSave,
  onInspect,
  onPack,
  onUnpack,
  onBrowseTable,
  onBrowseList,
  onClose
}: {
  info: DeflateInfo | null;
  busy: boolean;
  onPickFile: (title: string, filters: FileFilter[]) => Promise<string>;
  onPickSave: (title: string, defaultName: string, filters: FileFilter[]) => Promise<string>;
  onInspect: (path: string) => Promise<void>;
  onPack: (input: string, output: string, blockSize: number, headerSize: number, overwrite: boolean) => Promise<void>;
  onUnpack: (input: string, output: string, cacheMax: number, overwrite: boolean) => Promise<void>;
  onBrowseTable: (path: string) => Promise<void>;
  onBrowseList: (path: string) => Promise<void>;
  onClose: () => void;
}) {
  const [packInput, setPackInput] = useState('');
  const [packOutput, setPackOutput] = useState('');
  const [blockSize, setBlockSize] = useState(8192);
  const [headerSize, setHeaderSize] = useState(0);
  const [packOverwrite, setPackOverwrite] = useState(false);
  const [unpackInput, setUnpackInput] = useState('');
  const [unpackOutput, setUnpackOutput] = useState('');
  const [cacheMax, setCacheMax] = useState(8);
  const [unpackOverwrite, setUnpackOverwrite] = useState(false);
  const [error, setError] = useState('');

  async function browsePackInput() {
    const path = await onPickFile('Choose a file to compress', [{ displayName: 'All files', pattern: '*.*' }]);
    if (!path) return;
    setPackInput(path);
    if (!packOutput) setPackOutput(`${path}.deflate`);
  }

  async function browsePackOutput() {
    const path = await onPickSave('Save compressed file', 'data.bin.deflate', [{ displayName: 'Deflate files', pattern: '*.deflate' }]);
    if (path) setPackOutput(path);
  }

  async function browseUnpackInput() {
    const path = await onPickFile('Choose a deflate file', [{ displayName: 'Deflate files', pattern: '*.deflate;*.bin;*.*' }]);
    if (!path) return;
    setUnpackInput(path);
    if (!unpackOutput) setUnpackOutput(path.endsWith('.deflate') ? path.slice(0, -'.deflate'.length) : `${path}.raw`);
  }

  async function browseUnpackOutput() {
    const path = await onPickSave('Save uncompressed file', 'data.bin', [{ displayName: 'All files', pattern: '*.*' }]);
    if (path) setUnpackOutput(path);
  }

  async function submitPack(event: FormEvent) {
    event.preventDefault();
    if (!packInput || !packOutput) {
      setError('Choose an input file and an output path.');
      return;
    }
    setError('');
    await onPack(packInput, packOutput, blockSize, headerSize, packOverwrite);
  }

  async function submitUnpack(event: FormEvent) {
    event.preventDefault();
    if (!unpackInput || !unpackOutput) {
      setError('Choose a deflate file and an output path.');
      return;
    }
    setError('');
    await onUnpack(unpackInput, unpackOutput, cacheMax, unpackOverwrite);
  }

  async function inspectInput() {
    const path = await onPickFile('Inspect a deflate file', [{ displayName: 'Deflate files', pattern: '*.deflate;*.bin;*.*' }]);
    if (path) await onInspect(path);
  }

  return (
    <div className="flex h-full min-h-0 min-w-0 flex-col">
      <div className="shrink-0 border-b border-line bg-panel/35 px-4 py-3">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="truncate text-[14px] font-semibold text-ink">
                {info ? info.path.split(/[\\/]/).pop() : 'Deflate workbench'}
              </h1>
              <Badge tone="accent">Block deflate</Badge>
            </div>
            <p className="mt-0.5 max-w-[720px] truncate font-mono text-[10px] text-sub/70" title={info?.path}>
              {info ? info.path : 'Kotlin-compatible DeflateBinaryStreamReader / Writer format'}
            </p>
          </div>
          <div className="flex items-center gap-1.5">
            <Button size="sm" variant="secondary" icon={<FileSearch size={14} aria-hidden="true" />} onClick={inspectInput}>
              Inspect
            </Button>
            {info ? (
              <Button size="sm" variant="ghost" icon={<X size={14} aria-hidden="true" />} onClick={onClose}>
                Close
              </Button>
            ) : null}
          </div>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {info ? (
          <div className="mb-4 rounded-xl border border-line bg-panel/50 p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <div className="flex items-center gap-2">
                  <Archive size={16} className="text-accent-2" aria-hidden="true" />
                  <h2 className="text-[13px] font-semibold text-ink">Stream information</h2>
                </div>
                <p className="mt-1 text-[10px] text-sub">
                  This stream is readable by <span className="font-mono text-sub">DeflateBinaryStreamReader</span>.
                </p>
              </div>
              <div className="flex flex-wrap justify-end gap-1.5">
                <Button
                  size="sm"
                  variant="secondary"
                  icon={<Database size={14} aria-hidden="true" />}
                  onClick={() => void onBrowseTable(info.path)}
                  disabled={busy}
                >
                  Read as table
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  icon={<List size={14} aria-hidden="true" />}
                  onClick={() => void onBrowseList(info.path)}
                  disabled={busy}
                >
                  Read as list
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  icon={<ArrowDownToLine size={14} aria-hidden="true" />}
                  onClick={() => {
                    setUnpackInput(info.path);
                    setUnpackOutput(info.path.endsWith('.deflate') ? info.path.slice(0, -'.deflate'.length) : `${info.path}.raw`);
                  }}
                >
                  Fill unpack form
                </Button>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-6">
              <Stat label="Original" value={formatBytes(info.originalLength)} detail={`${formatCount(info.originalLength)} bytes`} />
              <Stat label="Compressed" value={formatBytes(info.fileSize)} detail={`${formatCount(info.fileSize)} bytes`} />
              <Stat label="Ratio" value={percent(info.ratio)} detail="compressed / original" />
              <Stat label="Blocks" value={`${formatCount(info.blocks)} / ${formatCount(info.maxBlocks)}`} detail="used / reserved" />
              <Stat label="Block size" value={formatBytes(info.blockSize)} detail={`${info.blockSize} bytes`} />
              <Stat label="Header" value={`${info.headerSize} B`} detail="caller data" />
            </div>

            <div className="mt-3">
              <div className="flex items-center justify-between text-[10px] text-sub">
                <span>Compression ratio</span>
                <span className="font-mono">{percent(info.ratio)}</span>
              </div>
              <div className="mt-1.5 h-1.5 overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-accent transition-all duration-300"
                  style={{ width: `${Math.min(100, Math.max(0, info.ratio * 100))}%` }}
                />
              </div>
            </div>
          </div>
        ) : null}

        {error ? (
          <p role="alert" className="mb-3 rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-[11px] text-[#fca5a5]">
            {error}
          </p>
        ) : null}

        <div className="grid min-w-0 gap-4 lg:grid-cols-2">
          <form onSubmit={submitPack} className="rounded-xl border border-line bg-panel/45 p-4">
            <div className="flex items-center gap-2">
              <ArrowUpFromLine size={16} className="text-accent-2" aria-hidden="true" />
              <h2 className="text-[13px] font-semibold text-ink">Pack a file</h2>
            </div>
            <p className="mt-1 text-[10px] text-sub">Compress with the Kotlin-compatible block index.</p>
            <div className="mt-4 space-y-3">
              <Field label="Input file" htmlFor="pack-input">
                <div className="flex gap-1.5">
                  <input id="pack-input" className={inputClass} value={packInput} onChange={(event) => setPackInput(event.target.value)} placeholder="data.bin" />
                  <IconButton label="Browse pack input" onClick={() => void browsePackInput()}>
                    <FolderOpen size={15} aria-hidden="true" />
                  </IconButton>
                </div>
              </Field>
              <Field label="Output file" htmlFor="pack-output">
                <div className="flex gap-1.5">
                  <input id="pack-output" className={inputClass} value={packOutput} onChange={(event) => setPackOutput(event.target.value)} placeholder="data.bin.deflate" />
                  <IconButton label="Browse pack output" onClick={() => void browsePackOutput()}>
                    <FolderOpen size={15} aria-hidden="true" />
                  </IconButton>
                </div>
              </Field>
              <div className="grid grid-cols-2 gap-3">
                <Field label="Block size" htmlFor="pack-block" hint="Default 8192">
                  <input id="pack-block" type="number" min={1} className={inputClass} value={blockSize} onChange={(event) => setBlockSize(Number(event.target.value))} />
                </Field>
                <Field label="Caller header" htmlFor="pack-header" hint="Reserved bytes">
                  <input id="pack-header" type="number" min={0} className={inputClass} value={headerSize} onChange={(event) => setHeaderSize(Number(event.target.value))} />
                </Field>
              </div>
              <label className="flex items-center gap-2 text-[11px] text-sub">
                <input type="checkbox" className="accent-[#22c55e]" checked={packOverwrite} onChange={(event) => setPackOverwrite(event.target.checked)} />
                Overwrite an existing output file
              </label>
              <Button type="submit" variant="primary" icon={<Archive size={14} aria-hidden="true" />} loading={busy} className="w-full">
                Pack
              </Button>
            </div>
          </form>

          <form onSubmit={submitUnpack} className="rounded-xl border border-line bg-panel/45 p-4">
            <div className="flex items-center gap-2">
              <ArrowDownToLine size={16} className="text-accent-2" aria-hidden="true" />
              <h2 className="text-[13px] font-semibold text-ink">Unpack a file</h2>
            </div>
            <p className="mt-1 text-[10px] text-sub">Read a block-deflate stream and restore the original bytes.</p>
            <div className="mt-4 space-y-3">
              <Field label="Deflate file" htmlFor="unpack-input">
                <div className="flex gap-1.5">
                  <input id="unpack-input" className={inputClass} value={unpackInput} onChange={(event) => setUnpackInput(event.target.value)} placeholder="data.bin.deflate" />
                  <IconButton label="Browse unpack input" onClick={() => void browseUnpackInput()}>
                    <FolderOpen size={15} aria-hidden="true" />
                  </IconButton>
                </div>
              </Field>
              <Field label="Output file" htmlFor="unpack-output">
                <div className="flex gap-1.5">
                  <input id="unpack-output" className={inputClass} value={unpackOutput} onChange={(event) => setUnpackOutput(event.target.value)} placeholder="data.bin" />
                  <IconButton label="Browse unpack output" onClick={() => void browseUnpackOutput()}>
                    <FolderOpen size={15} aria-hidden="true" />
                  </IconButton>
                </div>
              </Field>
              <Field label="Block cache" htmlFor="unpack-cache" hint="0 disables caching">
                <input id="unpack-cache" type="number" min={0} max={64} className={inputClass} value={cacheMax} onChange={(event) => setCacheMax(Number(event.target.value))} />
              </Field>
              <label className="flex items-center gap-2 text-[11px] text-sub">
                <input type="checkbox" className="accent-[#22c55e]" checked={unpackOverwrite} onChange={(event) => setUnpackOverwrite(event.target.checked)} />
                Overwrite an existing output file
              </label>
              <Button type="submit" variant="primary" icon={<ArrowDownToLine size={14} aria-hidden="true" />} loading={busy} className="w-full">
                Unpack
              </Button>
            </div>
          </form>
        </div>

        <div className="mt-4 rounded-xl border border-line bg-bg/30 p-3">
          <div className="flex items-center gap-2 text-[10px] text-sub">
            <RefreshCw size={13} aria-hidden="true" />
            Drop a file into the window to inspect it, or pack it with the default settings.
          </div>
        </div>
      </div>
    </div>
  );
}
