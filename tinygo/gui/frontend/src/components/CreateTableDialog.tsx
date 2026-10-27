import { useEffect, useState, type FormEvent } from 'react';
import { FolderOpen, Plus, Trash2 } from 'lucide-react';
import type { ColumnSpec, CreateTableRequest, FileFilter } from '../lib/api';
import { Button, Field, IconButton, inputClass, selectClass } from './ui';
import { Modal } from './Modal';

const typeDefaults: Record<string, number> = {
  int: 4,
  long: 8,
  double: 8,
  string: 64,
  raw: 1024
};

export function CreateTableDialog({
  open,
  busy,
  onClose,
  onCreate,
  onPickSave
}: {
  open: boolean;
  busy: boolean;
  onClose: () => void;
  onCreate: (request: CreateTableRequest) => Promise<boolean>;
  onPickSave: (title: string, defaultName: string, filters: FileFilter[]) => Promise<string>;
}) {
  const [path, setPath] = useState('');
  const [rows, setRows] = useState(1000);
  const [columns, setColumns] = useState<ColumnSpec[]>([
    { name: '_id', type: 'int', length: 4 },
    { name: 'name', type: 'string', length: 64 }
  ]);
  const [overwrite, setOverwrite] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open) return;
    setError('');
  }, [open]);

  function updateColumn(index: number, patch: Partial<ColumnSpec>) {
    setColumns((current) => current.map((column, i) => (i === index ? { ...column, ...patch } : column)));
  }

  function addColumn() {
    setColumns((current) => [...current, { name: `column_${current.length + 1}`, type: 'string', length: 64 }]);
  }

  function removeColumn(index: number) {
    setColumns((current) => current.filter((_, i) => i !== index));
  }

  async function browsePath() {
    const selected = await onPickSave('Create table file', 'new-table.db', [{ displayName: 'BinaryTable files', pattern: '*.db;*.bin;*.dat;*.*' }]);
    if (selected) setPath(selected);
  }

  function validate(): string {
    if (!path.trim()) return 'Choose an output path for the table.';
    if (!Number.isFinite(rows) || rows < 0) return 'Row capacity must be zero or greater.';
    if (columns.length === 0) return 'Add at least one column.';
    const names = new Set<string>();
    for (const column of columns) {
      const name = column.name.trim();
      if (!name) return 'Every column needs a name.';
      if (names.has(name)) return `Duplicate column name: ${name}`;
      names.add(name);
      if ((column.type === 'string' || column.type === 'raw') && (!Number.isFinite(column.length) || column.length <= 0)) {
        return `Column ${name} needs a positive byte length.`;
      }
    }
    return '';
  }

  async function submit(event: FormEvent) {
    event.preventDefault();
    const message = validate();
    if (message) {
      setError(message);
      return;
    }
    setError('');
    const created = await onCreate({
      path,
      rows: Math.floor(rows),
      columns: columns.map((column) => ({
        ...column,
        name: column.name.trim(),
        length: column.type === 'int' ? 4 : column.type === 'long' || column.type === 'double' ? 8 : Math.floor(column.length)
      })),
      overwrite
    });
    if (created) {
      setPath('');
      setRows(1000);
      setColumns([
        { name: '_id', type: 'int', length: 4 },
        { name: 'name', type: 'string', length: 64 }
      ]);
      setOverwrite(false);
      onClose();
    }
  }

  return (
    <Modal
      open={open}
      title="Create a table"
      description="Define a fixed schema. The file is allocated at schema size plus rows × row size."
      onClose={onClose}
      width="max-w-3xl"
      footer={
        <>
          <Button variant="ghost" onClick={onClose} disabled={busy}>
            Cancel
          </Button>
          <Button type="submit" form="create-table-form" variant="primary" loading={busy}>
            Create table
          </Button>
        </>
      }
    >
      <form id="create-table-form" onSubmit={submit} className="space-y-4">
        {error ? (
          <p role="alert" className="rounded-lg border border-danger/30 bg-danger/10 px-3 py-2 text-[11px] text-[#fca5a5]">
            {error}
          </p>
        ) : null}

        <div className="grid gap-4 sm:grid-cols-[1fr_140px]">
          <Field label="Output file" htmlFor="create-path">
            <div className="flex gap-1.5">
              <input
                id="create-path"
                data-autofocus
                className={inputClass}
                value={path}
                onChange={(event) => setPath(event.target.value)}
                placeholder="C:\data\users.db"
              />
              <IconButton label="Browse output file" onClick={() => void browsePath()}>
                <FolderOpen size={15} aria-hidden="true" />
              </IconButton>
            </div>
          </Field>
          <Field label="Rows" htmlFor="create-rows">
            <input
              id="create-rows"
              type="number"
              min={0}
              className={inputClass}
              value={rows}
              onChange={(event) => setRows(Number(event.target.value))}
            />
          </Field>
        </div>

        <div>
          <div className="flex items-center justify-between">
            <h3 className="text-[12px] font-medium text-sub">Columns</h3>
            <Button size="sm" variant="ghost" icon={<Plus size={14} aria-hidden="true" />} onClick={addColumn}>
              Add column
            </Button>
          </div>
          <div className="mt-2 space-y-2">
            {columns.map((column, index) => (
              <div key={index} className="grid grid-cols-[1fr_130px_110px_40px] items-end gap-2 rounded-lg border border-line bg-panel-2/50 p-2.5">
                <Field label={`Name ${index + 1}`} htmlFor={`column-name-${index}`}>
                  <input
                    id={`column-name-${index}`}
                    className={inputClass}
                    value={column.name}
                    onChange={(event) => updateColumn(index, { name: event.target.value })}
                  />
                </Field>
                <Field label="Type" htmlFor={`column-type-${index}`}>
                  <select
                    id={`column-type-${index}`}
                    className={selectClass}
                    value={column.type}
                    onChange={(event) => {
                      const type = event.target.value;
                      updateColumn(index, { type, length: typeDefaults[type] ?? 0 });
                    }}
                  >
                    <option value="int">INT · 4 B</option>
                    <option value="long">LONG · 8 B</option>
                    <option value="double">DOUBLE · 8 B</option>
                    <option value="string">STRING</option>
                    <option value="raw">RAW</option>
                  </select>
                </Field>
                <Field label="Length" htmlFor={`column-length-${index}`}>
                  <input
                    id={`column-length-${index}`}
                    type="number"
                    min={1}
                    className={inputClass}
                    value={column.length}
                    disabled={column.type === 'int' || column.type === 'long' || column.type === 'double'}
                    onChange={(event) => updateColumn(index, { length: Number(event.target.value) })}
                  />
                </Field>
                <IconButton
                  label={`Remove column ${column.name || index + 1}`}
                  onClick={() => removeColumn(index)}
                  disabled={columns.length <= 1}
                >
                  <Trash2 size={14} aria-hidden="true" />
                </IconButton>
              </div>
            ))}
          </div>
        </div>

        <label className="flex items-center gap-2 text-[11px] text-sub">
          <input type="checkbox" className="accent-[#22c55e]" checked={overwrite} onChange={(event) => setOverwrite(event.target.checked)} />
          Overwrite the file if it already exists
        </label>
      </form>
    </Modal>
  );
}
