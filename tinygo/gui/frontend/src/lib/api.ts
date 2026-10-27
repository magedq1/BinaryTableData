export type ColumnType = 'INT' | 'LONG' | 'DOUBLE' | 'STRING' | 'RAW';

export interface FileFilter {
  displayName: string;
  pattern: string;
}

export interface ColumnInfo {
  index: number;
  name: string;
  type: ColumnType;
  length: number;
  startPos: number;
}

export interface TableInfo {
  path: string;
  readOnly: boolean;
  compressed: boolean;
  rows: number;
  rowSize: number;
  schemaSize: number;
  capacity: number;
  fileSize: number;
  schema: string;
  columns: ColumnInfo[];
}

export interface CellValue {
  column: number;
  name: string;
  type: ColumnType;
  length: number;
  value: string;
  rawHex: string;
  empty: boolean;
}

export interface RowValue {
  index: number;
  cells: CellValue[];
}

export interface TablePage {
  start: number;
  total: number;
  limit: number;
  rows: RowValue[];
}

export interface ColumnSpec {
  name: string;
  type: string;
  length: number;
}

export interface CreateTableRequest {
  path: string;
  rows: number;
  columns: ColumnSpec[];
  overwrite: boolean;
}

export interface BlobInfo {
  index: number;
  size: number;
  name?: string;
}

export interface ListInfo {
  path: string;
  readOnly: boolean;
  compressed: boolean;
  entries: number;
  maxEntries: number;
  schemaSize: number;
  indexSize: number;
  fileSize: number;
  blobs: BlobInfo[];
}

export interface BlobData {
  index: number;
  size: number;
  mime: string;
  isText: boolean;
  text: string;
  hex: string;
  base64: string;
  truncated: boolean;
}

export interface CreateListRequest {
  path: string;
  maxEntries: number;
  overwrite: boolean;
}

export interface DeflateInfo {
  path: string;
  fileSize: number;
  blockSize: number;
  headerSize: number;
  maxBlocks: number;
  blocks: number;
  originalLength: number;
  outputSize: number;
  ratio: number;
}

type Bridge = Record<string, (...args: unknown[]) => Promise<unknown>>;

declare global {
  interface Window {
    go?: { main?: { App?: Bridge } };
    runtime?: {
      OnFileDrop?: (callback: (x: number, y: number, paths: string[]) => void, useDropTarget?: boolean) => void;
      OnFileDropOff?: () => void;
    };
  }
}

function call<T>(method: string, ...args: unknown[]): Promise<T> {
  const bridge = window.go?.main?.App;
  if (!bridge) {
    return Promise.reject(new Error('The Wails bridge is not available. Run this window through Wails.'));
  }
  const fn = bridge[method];
  if (!fn) {
    return Promise.reject(new Error(`Wails method ${method} is not bound.`));
  }
  return fn(...args) as Promise<T>;
}

export const api = {
  version: () => call<string>('Version'),

  pickFile: (title: string, filters: FileFilter[]) => call<string>('PickFile', title, filters),
  pickFiles: (title: string, filters: FileFilter[]) => call<string[]>('PickFiles', title, filters),
  pickSaveFile: (title: string, defaultFilename: string, filters: FileFilter[]) =>
    call<string>('PickSaveFile', title, defaultFilename, filters),

  openTable: (path: string) => call<TableInfo>('OpenTable', path),
  openTableReadOnly: (path: string) => call<TableInfo>('OpenTableReadOnly', path),
  openCompressedTable: (path: string) => call<TableInfo>('OpenCompressedTable', path),
  closeTable: (path: string) => call<void>('CloseTable', path),
  createTable: (request: CreateTableRequest) => call<TableInfo>('CreateTable', request),
  getTableRows: (path: string, start: number, limit: number) =>
    call<TablePage>('GetTableRows', path, start, limit),
  setTableCell: (path: string, row: number, column: number, value: string, format: string) =>
    call<CellValue>('SetTableCell', path, row, column, value, format),
  setTableCellFromFile: (path: string, row: number, column: number, sourcePath: string) =>
    call<CellValue>('SetTableCellFromFile', path, row, column, sourcePath),
  clearTableRow: (path: string, row: number) => call<void>('ClearTableRow', path, row),
  searchTableRows: (path: string, column: number, value: string, format: string, limit: number) =>
    call<number[]>('SearchTableRows', path, column, value, format, limit),

  openList: (path: string, maxEntries: number) => call<ListInfo>('OpenList', path, maxEntries),
  openCompressedList: (path: string, maxEntries: number) => call<ListInfo>('OpenCompressedList', path, maxEntries),
  createList: (request: CreateListRequest) => call<ListInfo>('CreateList', request),
  closeList: (path: string) => call<void>('CloseList', path),
  addBlobFromPath: (path: string, sourcePath: string) => call<BlobInfo>('AddBlobFromPath', path, sourcePath),
  addBlobFromBase64: (path: string, name: string, dataBase64: string) =>
    call<BlobInfo>('AddBlobFromBase64', path, name, dataBase64),
  getBlob: (path: string, index: number, previewLimit: number) =>
    call<BlobData>('GetBlob', path, index, previewLimit),
  exportBlob: (path: string, index: number, outputPath: string) =>
    call<void>('ExportBlob', path, index, outputPath),
  deleteBlob: (path: string, index: number) => call<ListInfo>('DeleteBlob', path, index),
  replaceBlobFromPath: (path: string, index: number, sourcePath: string) =>
    call<ListInfo>('ReplaceBlobFromPath', path, index, sourcePath),

  inspectDeflate: (path: string) => call<DeflateInfo>('InspectDeflate', path),
  packDeflate: (input: string, output: string, blockSize: number, headerSize: number, overwrite: boolean) =>
    call<DeflateInfo>('PackDeflate', input, output, blockSize, headerSize, overwrite),
  unpackDeflate: (input: string, output: string, cacheMax: number, overwrite: boolean) =>
    call<DeflateInfo>('UnpackDeflate', input, output, cacheMax, overwrite)
};

export const dialogFilters = {
  table: [{ displayName: 'BinaryTable files', pattern: '*.db;*.bin;*.dat;*.*' }],
  list: [{ displayName: 'Binary list files', pattern: '*.bin;*.list;*.*' }],
  deflate: [{ displayName: 'Deflate files', pattern: '*.deflate;*.bin;*.*' }],
  any: [{ displayName: 'All files', pattern: '*.*' }],
  images: [{ displayName: 'Images', pattern: '*.png;*.jpg;*.jpeg;*.gif;*.webp;*.bmp;*.ico' }]
};

export function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  return JSON.stringify(error);
}
