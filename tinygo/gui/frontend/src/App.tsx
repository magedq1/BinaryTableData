import { useEffect, useMemo, useRef, useState } from 'react';
import {
  Archive,
  Database,
  Download,
  FilePlus2,
  FileSearch,
  FileUp,
  FolderOpen,
  List,
  Plus,
  RefreshCw,
  Replace,
  Save,
  Trash2,
  X
} from 'lucide-react';
import {
  api,
  dialogFilters,
  errorMessage,
  type BlobData,
  type BlobInfo,
  type CellValue,
  type CreateTableRequest,
  type DeflateInfo,
  type FileFilter,
  type ListInfo,
  type TableInfo,
  type TablePage
} from './lib/api';
import { baseName, formatBytes } from './lib/format';
import { Toaster, useToasts } from './components/Toast';
import { StatusBar, Sidebar, TopBar, type RecentFile, type Workspace } from './components/Chrome';
import { TableEditor } from './components/TableEditor';
import { ListEditor } from './components/ListEditor';
import { DeflatePanel } from './components/DeflatePanel';
import { CreateTableDialog } from './components/CreateTableDialog';
import { ListDialog, type ListDialogMode } from './components/ListDialog';
import { CommandPalette, type CommandItem } from './components/CommandPalette';
import { ConfirmDialog } from './components/ConfirmDialog';

interface ConfirmState {
  title: string;
  body: string;
  confirmLabel: string;
  danger?: boolean;
  action: () => Promise<void>;
}

const RECENT_KEY = 'binarytable-studio.recent.v1';

function loadRecent(): RecentFile[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    const recent: RecentFile[] = [];
    for (const entry of parsed) {
      if (!entry || typeof entry.path !== 'string') continue;
      if (entry.kind !== 'table' && entry.kind !== 'list' && entry.kind !== 'deflate') continue;
      recent.push({
        path: entry.path,
        kind: entry.kind,
        max: typeof entry.max === 'number' ? entry.max : undefined,
        compressed: entry.compressed === true
      });
    }
    return recent;
  } catch {
    return [];
  }
}

function storeRecent(files: RecentFile[]) {
  try {
    localStorage.setItem(RECENT_KEY, JSON.stringify(files));
  } catch {
    // Recent files are a convenience; storage failures are non-fatal.
  }
}

function exportName(blob: BlobData): string {
  const extension =
    blob.mime === 'image/png'
      ? '.png'
      : blob.mime === 'image/jpeg'
        ? '.jpg'
        : blob.mime === 'image/gif'
          ? '.gif'
          : blob.mime === 'image/webp'
            ? '.webp'
            : blob.isText
              ? '.txt'
              : '.bin';
  return `blob-${blob.index}${extension}`;
}

export default function App() {
  const { toasts, notify, dismiss } = useToasts();
  const [active, setActive] = useState<Workspace>('table');
  const [table, setTable] = useState<TableInfo | null>(null);
  const [page, setPage] = useState<TablePage | null>(null);
  const [tableStart, setTableStart] = useState(0);
  const [pageSize, setPageSize] = useState(100);
  const [selected, setSelected] = useState<{ row: number; column: number } | null>(null);
  const [searchHits, setSearchHits] = useState<number[]>([]);
  const [list, setList] = useState<ListInfo | null>(null);
  const [selectedBlob, setSelectedBlob] = useState<number | null>(null);
  const [blob, setBlob] = useState<BlobData | null>(null);
  const [deflate, setDeflate] = useState<DeflateInfo | null>(null);
  const [busy, setBusy] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [listDialog, setListDialog] = useState<ListDialogMode | null>(null);
  const [commandOpen, setCommandOpen] = useState(false);
  const [confirm, setConfirm] = useState<ConfirmState | null>(null);
  const [recent, setRecent] = useState<RecentFile[]>(loadRecent);

  const selectedCell = useMemo<CellValue | null>(() => {
    if (!selected || !page) return null;
    const row = page.rows.find((entry) => entry.index === selected.row);
    return row?.cells.find((entry) => entry.column === selected.column) ?? null;
  }, [page, selected]);

  function remember(path: string, kind: Workspace, max?: number, compressed = false) {
    setRecent((current) => {
      const next = [{ path, kind, max, compressed }, ...current.filter((entry) => entry.path !== path)].slice(0, 10);
      storeRecent(next);
      return next;
    });
  }

  function removeRecent(path: string) {
    setRecent((current) => {
      const next = current.filter((entry) => entry.path !== path);
      storeRecent(next);
      return next;
    });
  }

  async function pickFile(title: string, filters: FileFilter[]): Promise<string> {
    try {
      return await api.pickFile(title, filters);
    } catch (error) {
      notify(errorMessage(error), 'error');
      return '';
    }
  }

  async function pickFiles(title: string, filters: FileFilter[]): Promise<string[]> {
    try {
      return await api.pickFiles(title, filters);
    } catch (error) {
      notify(errorMessage(error), 'error');
      return [];
    }
  }

  async function pickSaveFile(title: string, defaultName: string, filters: FileFilter[]): Promise<string> {
    try {
      return await api.pickSaveFile(title, defaultName, filters);
    } catch (error) {
      notify(errorMessage(error), 'error');
      return '';
    }
  }

  async function openTablePath(path: string, readOnly = false) {
    setBusy('Opening table…');
    try {
      const info = readOnly ? await api.openTableReadOnly(path) : await api.openTable(path);
      const rows = await api.getTableRows(info.path, 0, pageSize);
      setTable(info);
      setPage(rows);
      setTableStart(rows.start);
      setSelected(null);
      setSearchHits([]);
      setActive('table');
      remember(info.path, 'table');
      notify(`Opened ${baseName(info.path)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function openTableDialog() {
    const path = await pickFile('Open table', dialogFilters.table);
    if (path) await openTablePath(path);
  }

  async function openCompressedTablePath(path: string) {
    setBusy('Opening compressed table…');
    try {
      const info = await api.openCompressedTable(path);
      const rows = await api.getTableRows(info.path, 0, pageSize);
      setTable(info);
      setPage(rows);
      setTableStart(rows.start);
      setSelected(null);
      setSearchHits([]);
      setActive('table');
      remember(info.path, 'table', undefined, true);
      notify(`Opened compressed table ${baseName(info.path)} (read-only)`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function openCompressedTableDialog() {
    const path = await pickFile('Open a table through DeflateBinaryStreamReader', dialogFilters.deflate);
    if (path) await openCompressedTablePath(path);
  }

  async function loadRows(path: string, start: number, size: number) {
    setBusy('Loading rows…');
    try {
      const rows = await api.getTableRows(path, start, size);
      setPage(rows);
      setTableStart(rows.start);
      setSelected(null);
      return rows;
    } catch (error) {
      notify(errorMessage(error), 'error');
      return null;
    } finally {
      setBusy(null);
    }
  }

  async function changePage(start: number) {
    if (!table) return;
    await loadRows(table.path, Math.max(0, start), pageSize);
  }

  async function changePageSize(size: number) {
    setPageSize(size);
    if (table) await loadRows(table.path, 0, size);
  }

  async function reloadTable() {
    if (!table) return;
    setBusy('Reloading table…');
    try {
      const info = table.compressed ? await api.openCompressedTable(table.path) : await api.openTable(table.path);
      const rows = await api.getTableRows(info.path, tableStart, pageSize);
      setTable(info);
      setPage(rows);
      setSelected(null);
      setSearchHits([]);
      notify('Table reloaded', 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function closeTable() {
    if (!table) return;
    try {
      await api.closeTable(table.path);
    } catch (error) {
      notify(errorMessage(error), 'error');
      return;
    }
    setTable(null);
    setPage(null);
    setSelected(null);
    setSearchHits([]);
    notify('Table closed', 'info');
  }

  function applyCell(cell: CellValue, rowIndex: number) {
    setPage((current) => {
      if (!current) return current;
      return {
        ...current,
        rows: current.rows.map((row) =>
          row.index === rowIndex
            ? { ...row, cells: row.cells.map((entry) => (entry.column === cell.column ? cell : entry)) }
            : row
        )
      };
    });
  }

  async function saveCell(value: string, format: string) {
    if (!table || !selected) return;
    const { row, column } = selected;
    setBusy('Saving cell…');
    try {
      const cell = await api.setTableCell(table.path, row, column, value, format);
      applyCell(cell, row);
      notify('Cell updated', 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function clearCell() {
    if (!table || !selected || !selectedCell) return;
    const { row, column } = selected;
    const format = selectedCell.type === 'RAW' ? 'hex' : 'typed';
    const value = selectedCell.type === 'STRING' || selectedCell.type === 'RAW' ? '' : '0';
    setBusy('Clearing cell…');
    try {
      const cell = await api.setTableCell(table.path, row, column, value, format);
      applyCell(cell, row);
      notify('Cell cleared', 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  function requestClearRow() {
    if (!table || !selected) return;
    setConfirm({
      title: `Clear row ${selected.row}?`,
      body: 'Every column in this row will be zero-filled. The row index and table capacity are unchanged.',
      confirmLabel: 'Clear row',
      danger: true,
      action: async () => {
        setBusy('Clearing row…');
        try {
          await api.clearTableRow(table.path, selected.row);
          await loadRows(table.path, tableStart, pageSize);
          notify('Row cleared', 'success');
        } catch (error) {
          notify(errorMessage(error), 'error');
        } finally {
          setBusy(null);
        }
      }
    });
  }

  async function embedFile() {
    if (!table || !selected || !selectedCell || selectedCell.type !== 'RAW') return;
    const { row, column } = selected;
    const source = await pickFile('Embed a file in this RAW cell', dialogFilters.any);
    if (!source) return;
    setBusy('Embedding file…');
    try {
      const cell = await api.setTableCellFromFile(table.path, row, column, source);
      applyCell(cell, row);
      notify(`Embedded ${baseName(source)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function searchRows(column: number, value: string, format: string) {
    if (!table) return;
    try {
      const hits = await api.searchTableRows(table.path, column, value, format, 500);
      setSearchHits(hits);
      notify(hits.length ? `${hits.length} matching rows` : 'No matching rows', hits.length ? 'success' : 'info');
    } catch (error) {
      setSearchHits([]);
      notify(errorMessage(error), 'error');
    }
  }

  async function jumpToRow(row: number) {
    if (!table) return;
    const start = Math.floor(row / pageSize) * pageSize;
    const rows = await loadRows(table.path, start, pageSize);
    if (!rows) return;
    const target = rows.rows.find((entry) => entry.index === row);
    if (target) {
      setSelected({ row, column: selected?.column ?? target.cells[0]?.column ?? 0 });
    }
  }

  async function createTable(request: CreateTableRequest): Promise<boolean> {
    setBusy('Creating table…');
    try {
      const info = await api.createTable(request);
      const rows = await api.getTableRows(info.path, 0, pageSize);
      setTable(info);
      setPage(rows);
      setTableStart(rows.start);
      setSelected(null);
      setSearchHits([]);
      setActive('table');
      remember(info.path, 'table');
      notify(`Created ${baseName(info.path)}`, 'success');
      return true;
    } catch (error) {
      notify(errorMessage(error), 'error');
      return false;
    } finally {
      setBusy(null);
    }
  }

  async function openListPath(path: string, maxEntries = 0) {
    setBusy('Opening binary list…');
    try {
      const info = await api.openList(path, maxEntries);
      setList(info);
      setSelectedBlob(null);
      setBlob(null);
      setActive('list');
      remember(info.path, 'list', info.maxEntries, false);
      notify(`Opened ${baseName(info.path)} (${info.entries} blobs)`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function openCompressedListPath(path: string, maxEntries = 0) {
    setBusy('Opening compressed binary list…');
    try {
      const info = await api.openCompressedList(path, maxEntries);
      setList(info);
      setSelectedBlob(null);
      setBlob(null);
      setActive('list');
      remember(info.path, 'list', info.maxEntries, true);
      notify(`Opened compressed list ${baseName(info.path)} (read-only)`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function openCompressedListDialog() {
    const path = await pickFile('Open a list through DeflateBinaryStreamReader', dialogFilters.deflate);
    if (path) await openCompressedListPath(path, 0);
  }

  function openListDialog() {
    setListDialog('open');
  }

  function createList() {
    setListDialog('create');
  }

  async function submitListDialog(path: string, maxEntries: number, compressed: boolean, overwrite: boolean): Promise<boolean> {
    const mode = listDialog;
    if (!mode) return false;
    setBusy(mode === 'create' ? 'Creating binary list…' : 'Opening binary list…');
    try {
      let info: ListInfo;
      if (mode === 'create') {
        info = await api.createList({ path, maxEntries, overwrite });
      } else if (compressed) {
        info = await api.openCompressedList(path, maxEntries);
      } else {
        info = await api.openList(path, maxEntries);
      }
      setList(info);
      setSelectedBlob(null);
      setBlob(null);
      setActive('list');
      remember(info.path, 'list', info.maxEntries, info.compressed);
      notify(
        mode === 'create'
          ? `Created ${baseName(info.path)} with ${info.maxEntries} slots`
          : `Opened ${baseName(info.path)} (${info.entries} blobs${info.readOnly ? ', read-only' : ''})`,
        'success'
      );
      return true;
    } catch (error) {
      notify(errorMessage(error), 'error');
      return false;
    } finally {
      setBusy(null);
    }
  }

  async function closeList() {
    if (!list) return;
    try {
      await api.closeList(list.path);
    } catch (error) {
      notify(errorMessage(error), 'error');
      return;
    }
    setList(null);
    setSelectedBlob(null);
    setBlob(null);
    notify('Binary list closed', 'info');
  }

  async function loadBlob(info: ListInfo, index: number) {
    const data = await api.getBlob(info.path, index, 512 * 1024);
    setSelectedBlob(index);
    setBlob(data);
  }

  async function selectBlob(index: number) {
    if (!list) return;
    setSelectedBlob(index);
    setBusy('Reading blob…');
    try {
      await loadBlob(list, index);
    } catch (error) {
      setSelectedBlob(null);
      setBlob(null);
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function refreshList() {
    if (!list) return;
    setBusy('Refreshing list…');
    try {
      const info = list.compressed
        ? await api.openCompressedList(list.path, list.maxEntries)
        : await api.openList(list.path, list.maxEntries);
      setList(info);
      if (selectedBlob !== null && selectedBlob < info.entries) {
        await loadBlob(info, selectedBlob);
      } else {
        setSelectedBlob(null);
        setBlob(null);
      }
      notify('List refreshed', 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function addBlobPaths(paths: string[]) {
    if (!list || paths.length === 0) return;
    if (list.readOnly) {
      notify('This list is a read-only compressed view.', 'info');
      return;
    }
    setBusy(`Embedding ${paths.length} file${paths.length === 1 ? '' : 's'}…`);
    const added: BlobInfo[] = [];
    const failures: string[] = [];
    try {
      for (const path of paths) {
        try {
          added.push(await api.addBlobFromPath(list.path, path));
        } catch (error) {
          failures.push(`${baseName(path)}: ${errorMessage(error)}`);
        }
      }
      const info = await api.openList(list.path, list.maxEntries);
      setList(info);
      if (added.length > 0) {
        await loadBlob(info, added[added.length - 1].index);
        notify(`${added.length} blob${added.length === 1 ? '' : 's'} embedded`, 'success');
      }
      if (failures.length > 0) {
        notify(failures.join(' · '), 'error');
      }
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function addFilesDialog() {
    const paths = await pickFiles('Embed files in the binary list', dialogFilters.any);
    if (paths.length) await addBlobPaths(paths);
  }

  async function exportBlob(index: number | null = selectedBlob) {
    if (!list || index === null) return;
    const defaultName = blob && blob.index === index ? exportName(blob) : `blob-${index}.bin`;
    const path = await pickSaveFile('Export blob', defaultName, dialogFilters.any);
    if (!path) return;
    setBusy('Exporting blob…');
    try {
      await api.exportBlob(list.path, index, path);
      notify(`Exported to ${baseName(path)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function replaceBlob(index: number | null = selectedBlob) {
    if (!list || index === null) return;
    if (list.readOnly) {
      notify('This list is a read-only compressed view.', 'info');
      return;
    }
    const path = await pickFile('Replace blob with a file', dialogFilters.any);
    if (!path) return;
    setBusy('Replacing blob…');
    try {
      const info = await api.replaceBlobFromPath(list.path, index, path);
      setList(info);
      await loadBlob(info, index);
      notify(`Replaced blob #${index}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  function requestDeleteBlob(index: number | null = selectedBlob) {
    if (!list || index === null) return;
    if (list.readOnly) {
      notify('This list is a read-only compressed view.', 'info');
      return;
    }
    setConfirm({
      title: `Delete blob #${index}?`,
      body: 'The list is rebuilt without this entry. Blob indices after it shift down by one. This cannot be undone.',
      confirmLabel: 'Delete blob',
      danger: true,
      action: async () => {
        setBusy('Deleting blob…');
        try {
          const info = await api.deleteBlob(list.path, index);
          setList(info);
          setSelectedBlob(null);
          setBlob(null);
          notify(`Deleted blob #${index}`, 'success');
        } catch (error) {
          notify(errorMessage(error), 'error');
        } finally {
          setBusy(null);
        }
      }
    });
  }

  async function inspectDeflatePath(path: string) {
    setBusy('Inspecting deflate stream…');
    try {
      const info = await api.inspectDeflate(path);
      setDeflate(info);
      setActive('deflate');
      remember(info.path, 'deflate');
      notify(`Inspected ${baseName(info.path)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function inspectDeflateDialog() {
    const path = await pickFile('Inspect a deflate file', dialogFilters.deflate);
    if (path) await inspectDeflatePath(path);
  }

  async function packDeflate(input: string, output: string, blockSize: number, headerSize: number, overwrite: boolean) {
    setBusy('Packing deflate stream…');
    try {
      const info = await api.packDeflate(input, output, blockSize, headerSize, overwrite);
      setDeflate(info);
      setActive('deflate');
      remember(info.path, 'deflate');
      notify(`Packed ${formatBytes(info.originalLength)} into ${formatBytes(info.fileSize)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function unpackDeflate(input: string, output: string, cacheMax: number, overwrite: boolean) {
    setBusy('Unpacking deflate stream…');
    try {
      const info = await api.unpackDeflate(input, output, cacheMax, overwrite);
      setDeflate(info);
      setActive('deflate');
      notify(`Unpacked ${formatBytes(info.outputSize)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  async function runConfirm() {
    const current = confirm;
    if (!current) return;
    setConfirm(null);
    await current.action();
  }

  const dropRef = useRef<(paths: string[]) => void>(() => undefined);
  dropRef.current = (paths: string[]) => {
    if (paths.length === 0) return;
    if (active === 'list') {
      if (list) void addBlobPaths(paths);
      else notify('Open or create a binary list before dropping files.', 'info');
      return;
    }
    if (active === 'table' && table) {
      if (selectedCell?.type === 'RAW' && paths.length === 1) {
        void openEmbeddedDrop(paths[0]);
        return;
      }
      if (paths.length === 1) {
        void openTablePath(paths[0]);
        return;
      }
      notify('Drop one table file at a time.', 'info');
      return;
    }
    if (active === 'deflate') {
      void inspectDeflatePath(paths[0]);
      return;
    }
    if (paths.length === 1) {
      void openTablePath(paths[0]);
    }
  };

  async function openEmbeddedDrop(path: string) {
    if (!table || !selected || !selectedCell || selectedCell.type !== 'RAW') return;
    const { row, column } = selected;
    setBusy('Embedding dropped file…');
    try {
      const cell = await api.setTableCellFromFile(table.path, row, column, path);
      applyCell(cell, row);
      notify(`Embedded ${baseName(path)}`, 'success');
    } catch (error) {
      notify(errorMessage(error), 'error');
    } finally {
      setBusy(null);
    }
  }

  useEffect(() => {
    const runtime = window.runtime;
    if (!runtime?.OnFileDrop) return;
    runtime.OnFileDrop((_x, _y, paths) => dropRef.current(paths), false);
    return () => runtime.OnFileDropOff?.();
  }, []);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      const modifier = event.ctrlKey || event.metaKey;
      if (modifier && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setCommandOpen(true);
      } else if (modifier && event.key.toLowerCase() === 'o') {
        event.preventDefault();
        void openCurrent();
      } else if (modifier && event.key.toLowerCase() === 'n') {
        event.preventDefault();
        setCreateOpen(true);
      } else if (modifier && event.shiftKey && event.key.toLowerCase() === 'l') {
        event.preventDefault();
        void openListDialog();
      } else if (event.key === 'Escape') {
        if (commandOpen) setCommandOpen(false);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  });

  async function openCurrent() {
    if (active === 'table') await openTableDialog();
    else if (active === 'list') await openListDialog();
    else await inspectDeflateDialog();
  }

  function createCurrent() {
    if (active === 'table') setCreateOpen(true);
    else if (active === 'list') void createList();
    else notify('Use the Pack form to create a deflate file.', 'info');
  }

  async function closeCurrent() {
    if (active === 'table') await closeTable();
    else if (active === 'list') await closeList();
    else setDeflate(null);
  }

  const commands: CommandItem[] = [
    {
      id: 'open-table',
      group: 'File',
      label: 'Open table',
      hint: 'Choose a BinaryTable file',
      shortcut: 'Ctrl O',
      icon: <FolderOpen size={15} aria-hidden="true" />,
      run: () => void openTableDialog()
    },
    {
      id: 'create-table',
      group: 'File',
      label: 'Create table',
      hint: 'Define a schema and allocate rows',
      shortcut: 'Ctrl N',
      icon: <FilePlus2 size={15} aria-hidden="true" />,
      run: () => setCreateOpen(true)
    },
    {
      id: 'open-list',
      group: 'File',
      label: 'Open binary list',
      hint: 'Edit variable-length blobs',
      shortcut: 'Ctrl Shift L',
      icon: <List size={15} aria-hidden="true" />,
      run: () => void openListDialog()
    },
    {
      id: 'create-list',
      group: 'File',
      label: 'Create binary list',
      hint: 'Start a new blob container',
      icon: <Plus size={15} aria-hidden="true" />,
      run: () => void createList()
    },
    {
      id: 'inspect-deflate',
      group: 'File',
      label: 'Inspect deflate file',
      hint: 'Read the block index and original length',
      icon: <FileSearch size={15} aria-hidden="true" />,
      run: () => void inspectDeflateDialog()
    },
    {
      id: 'open-compressed-table',
      group: 'File',
      label: 'Open compressed table',
      hint: 'Read a table through DeflateBinaryStreamReader',
      icon: <Archive size={15} aria-hidden="true" />,
      run: () => void openCompressedTableDialog()
    },
    {
      id: 'open-compressed-list',
      group: 'File',
      label: 'Open compressed binary list',
      hint: 'Read a list through DeflateBinaryStreamReader',
      icon: <Archive size={15} aria-hidden="true" />,
      run: () => void openCompressedListDialog()
    },
    {
      id: 'reload',
      group: 'Workspace',
      label: 'Reload current view',
      hint: 'Reload the active file',
      icon: <RefreshCw size={15} aria-hidden="true" />,
      run: () => {
        if (active === 'table') void reloadTable();
        else if (active === 'list') void refreshList();
        else if (deflate) void inspectDeflatePath(deflate.path);
      }
    },
    {
      id: 'close',
      group: 'Workspace',
      label: 'Close current file',
      hint: 'Close the active table or list',
      icon: <X size={15} aria-hidden="true" />,
      run: () => void closeCurrent()
    },
    {
      id: 'show-table',
      group: 'Navigate',
      label: 'Go to tables',
      hint: 'Fixed-schema editor',
      icon: <Database size={15} aria-hidden="true" />,
      run: () => setActive('table')
    },
    {
      id: 'show-list',
      group: 'Navigate',
      label: 'Go to binary lists',
      hint: 'Blob editor',
      icon: <List size={15} aria-hidden="true" />,
      run: () => setActive('list')
    },
    {
      id: 'show-deflate',
      group: 'Navigate',
      label: 'Go to deflate',
      hint: 'Pack and unpack streams',
      icon: <Archive size={15} aria-hidden="true" />,
      run: () => setActive('deflate')
    }
  ];

  if (list) {
    commands.push({
      id: 'export-blob',
      group: 'Binary list',
      label: 'Export selected blob',
      hint: 'Write the selected blob to a file',
      icon: <Download size={15} aria-hidden="true" />,
      run: () => void exportBlob()
    });
    if (!list.readOnly) {
      commands.push(
        {
          id: 'add-blobs',
          group: 'Binary list',
          label: 'Embed files',
          hint: 'Add files to the current list',
          icon: <FileUp size={15} aria-hidden="true" />,
          run: () => void addFilesDialog()
        },
        {
          id: 'replace-blob',
          group: 'Binary list',
          label: 'Replace selected blob',
          hint: 'Rebuild the list with a new file',
          icon: <Replace size={15} aria-hidden="true" />,
          run: () => void replaceBlob()
        },
        {
          id: 'delete-blob',
          group: 'Binary list',
          label: 'Delete selected blob',
          hint: 'Remove and re-index the list',
          icon: <Trash2 size={15} aria-hidden="true" />,
          run: requestDeleteBlob
        }
      );
    }
  }

  if (table && !table.readOnly && selectedCell?.type === 'RAW') {
    commands.push({
      id: 'embed-cell',
      group: 'Table',
      label: 'Embed file in selected RAW cell',
      hint: 'Write a file into the current cell',
      icon: <Save size={15} aria-hidden="true" />,
      run: () => void embedFile()
    });
  }

  return (
    <div className="flex h-screen flex-col bg-bg text-ink">
      <TopBar
        active={active}
        table={table}
        list={list}
        deflate={deflate}
        onOpen={() => void openCurrent()}
        onCreate={createCurrent}
        onCommand={() => setCommandOpen(true)}
        onCloseCurrent={() => void closeCurrent()}
      />

      <div className="flex min-h-0 flex-1">
        <Sidebar
          active={active}
          onSelect={setActive}
          recent={recent}
          onOpenRecent={(file) => {
            if (file.kind === 'table') {
              if (file.compressed) void openCompressedTablePath(file.path);
              else void openTablePath(file.path);
            } else if (file.kind === 'list') {
              if (file.compressed) void openCompressedListPath(file.path, file.max ?? 0);
              else void openListPath(file.path, file.max ?? 0);
            } else {
              void inspectDeflatePath(file.path);
            }
          }}
          onRemoveRecent={removeRecent}
          onOpen={() => void openCurrent()}
          onCreate={() => setCreateOpen(true)}
        />

        <main className="flex min-w-0 flex-1 flex-col">
          {active === 'table' ? (
            <TableEditor
              info={table}
              page={page}
              pageSize={pageSize}
              selected={selected}
              cell={selectedCell}
              searchHits={searchHits}
              busy={Boolean(busy)}
              onOpen={() => void openTableDialog()}
              onOpenCompressed={() => void openCompressedTableDialog()}
              onCreate={() => setCreateOpen(true)}
              onClose={() => void closeTable()}
              onReload={() => void reloadTable()}
              onPageChange={(start) => void changePage(start)}
              onPageSize={(size) => void changePageSize(size)}
              onSelect={(row, column) => setSelected({ row, column })}
              onSaveCell={saveCell}
              onClearCell={clearCell}
              onClearRow={requestClearRow}
              onEmbedFile={() => void embedFile()}
              onSearch={searchRows}
              onClearSearch={() => setSearchHits([])}
              onJumpToRow={(row) => void jumpToRow(row)}
              onCloseInspector={() => setSelected(null)}
            />
          ) : active === 'list' ? (
            <ListEditor
              info={list}
              selectedIndex={selectedBlob}
              blob={blob}
              busy={Boolean(busy)}
              onOpen={() => openListDialog()}
              onOpenCompressed={() => void openCompressedListDialog()}
              onCreate={() => createList()}
              onClose={() => void closeList()}
              onRefresh={() => void refreshList()}
              onAddFiles={() => void addFilesDialog()}
              onSelect={(index) => void selectBlob(index)}
              onExport={(index) => void exportBlob(index)}
              onReplace={(index) => void replaceBlob(index)}
              onDelete={(index) => requestDeleteBlob(index)}
            />
          ) : (
            <DeflatePanel
              info={deflate}
              busy={Boolean(busy)}
              onPickFile={pickFile}
              onPickSave={pickSaveFile}
              onInspect={inspectDeflatePath}
              onPack={packDeflate}
              onUnpack={unpackDeflate}
              onBrowseTable={openCompressedTablePath}
              onBrowseList={(path) => openCompressedListPath(path, 0)}
              onClose={() => setDeflate(null)}
            />
          )}
        </main>
      </div>

      <StatusBar active={active} table={table} list={list} deflate={deflate} busy={busy} />

      <CreateTableDialog
        open={createOpen}
        busy={Boolean(busy)}
        onClose={() => setCreateOpen(false)}
        onCreate={createTable}
        onPickSave={pickSaveFile}
      />

      <ListDialog
        open={listDialog !== null}
        mode={listDialog ?? 'open'}
        busy={Boolean(busy)}
        onClose={() => setListDialog(null)}
        onSubmit={submitListDialog}
        onPickFile={pickFile}
        onPickSave={pickSaveFile}
      />

      <CommandPalette open={commandOpen} onClose={() => setCommandOpen(false)} commands={commands} />

      <ConfirmDialog
        open={Boolean(confirm)}
        title={confirm?.title ?? ''}
        body={confirm?.body ?? ''}
        confirmLabel={confirm?.confirmLabel ?? 'Confirm'}
        danger={confirm?.danger}
        busy={Boolean(busy)}
        onConfirm={() => void runConfirm()}
        onClose={() => setConfirm(null)}
      />

      <Toaster toasts={toasts} dismiss={dismiss} />
    </div>
  );
}
