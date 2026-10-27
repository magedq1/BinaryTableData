import { useEffect, useState, type FormEvent } from 'react';
import { FolderOpen } from 'lucide-react';
import type { FileFilter } from '../lib/api';
import { Modal } from './Modal';
import { Button, Field, IconButton, inputClass } from './ui';

export type ListDialogMode = 'open' | 'create';

export function ListDialog({
  open,
  mode,
  busy,
  onClose,
  onSubmit,
  onPickFile,
  onPickSave
}: {
  open: boolean;
  mode: ListDialogMode;
  busy: boolean;
  onClose: () => void;
  onSubmit: (path: string, maxEntries: number, compressed: boolean, overwrite: boolean) => Promise<boolean>;
  onPickFile: (title: string, filters: FileFilter[]) => Promise<string>;
  onPickSave: (title: string, defaultName: string, filters: FileFilter[]) => Promise<string>;
}) {
  const [path, setPath] = useState('');
  const [maxEntries, setMaxEntries] = useState(mode === 'create' ? 1000 : 0);
  const [compressed, setCompressed] = useState(false);
  const [overwrite, setOverwrite] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    setPath('');
    setMaxEntries(mode === 'create' ? 1000 : 0);
    setCompressed(false);
    setOverwrite(false);
    setError('');
  }, [open, mode]);

  async function browse() {
    const filters: FileFilter[] = [{ displayName: 'Binary list files', pattern: '*.bin;*.list;*.*' }];
    if (mode === 'create') {
      const selected = await onPickSave('Create binary list', 'blobs.bin', filters);
      if (selected) setPath(selected);
    } else {
      const selected = await onPickFile(compressed ? 'Open compressed binary list' : 'Open binary list', compressed
        ? [{ displayName: 'Deflate files', pattern: '*.deflate;*.bin;*.*' }]
        : filters);
      if (selected) setPath(selected);
    }
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (!path.trim()) {
      setError('Choose a file path.');
      return;
    }
    if (mode === 'create' && (!Number.isFinite(maxEntries) || maxEntries <= 0)) {
      setError('Capacity must be greater than zero.');
      return;
    }
    if (mode === 'open' && (!Number.isFinite(maxEntries) || maxEntries < 0)) {
      setError('Capacity must be zero (auto) or greater.');
      return;
    }
    setError('');
    const done = await onSubmit(path.trim(), Math.floor(maxEntries), compressed, overwrite);
    if (done) onClose();
  }

  const title = mode === 'create' ? 'Create a binary list' : 'Open a binary list';
  const description =
    mode === 'create'
      ? 'Choose the maximum number of blobs. The capacity is stored in the reserved prefix and recovered on later opens.'
      : 'Capacity 0 tries to recover maxListSize from the file. Set it manually for legacy or empty lists.';

  return (
    <Modal
      open={open}
      title={title}
      description={description}
      onClose={onClose}
      width="max-w-xl"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form="list-dialog-form" variant="primary" loading={busy}>
            {mode === 'create' ? 'Create list' : 'Open list'}
          </Button>
        </>
      }
    >
      <form id="list-dialog-form" onSubmit={submit} className="space-y-4">
        {error ? (
          <p role="alert" className="rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-[11px] text-[#fca5a5]">
            {error}
          </p>
        ) : null}
        <Field label="File" htmlFor="list-dialog-path">
          <div className="flex gap-1.5">
            <input
              id="list-dialog-path"
              data-autofocus
              className={inputClass}
              value={path}
              onChange={(event) => setPath(event.target.value)}
              placeholder="C:\data\blobs.bin"
            />
            <IconButton label="Browse list file" onClick={() => void browse()}>
              <FolderOpen size={15} aria-hidden="true" />
            </IconButton>
          </div>
        </Field>
        <Field
          label="Max entries"
          htmlFor="list-dialog-max"
          hint={mode === 'create' ? 'Fixed capacity; data starts after 100 + maxEntries × 4 bytes.' : 'Use 0 to auto-detect.'}
        >
          <input
            id="list-dialog-max"
            type="number"
            min={mode === 'create' ? 1 : 0}
            className={inputClass}
            value={maxEntries}
            onChange={(event) => setMaxEntries(Number(event.target.value))}
          />
        </Field>
        {mode === 'open' ? (
          <label className="flex items-start gap-2 rounded-lg border border-line bg-panel-2/50 px-3 py-2 text-[11px] text-sub">
            <input
              type="checkbox"
              className="mt-0.5 accent-[#22c55e]"
              checked={compressed}
              onChange={(event) => setCompressed(event.target.checked)}
            />
            <span>
              <span className="block font-medium text-ink">Read a compressed list</span>
              <span className="block text-[10px] text-sub/70">
                Opens the file through the DeflateBinaryStreamReader equivalent. The view is read-only.
              </span>
            </span>
          </label>
        ) : (
          <label className="flex items-center gap-2 text-[11px] text-sub">
            <input type="checkbox" className="accent-[#22c55e]" checked={overwrite} onChange={(event) => setOverwrite(event.target.checked)} />
            Overwrite the file if it already exists
          </label>
        )}
      </form>
    </Modal>
  );
}
