export function formatBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) return '—';
  if (value < 1024) return `${value} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let size = value;
  let unit = -1;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${size.toFixed(size >= 10 ? 1 : 2)} ${units[unit]}`;
}

export function formatCount(value: number): string {
  return new Intl.NumberFormat('en-US').format(value);
}

export function percent(value: number): string {
  if (!Number.isFinite(value)) return '—';
  return `${(value * 100).toFixed(value > 1 ? 0 : 1)}%`;
}

export function baseName(path: string): string {
  if (!path) return 'No file';
  const parts = path.split(/[\\/]/);
  return parts[parts.length - 1] || path;
}

export function dirName(path: string): string {
  if (!path) return '';
  const index = Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\'));
  return index > 0 ? path.slice(0, index) : path;
}

export function truncateMiddle(value: string, max = 48): string {
  if (value.length <= max) return value;
  const side = Math.floor((max - 1) / 2);
  return `${value.slice(0, side)}…${value.slice(value.length - side)}`;
}

export function displayCellValue(value: string, empty: boolean): string {
  if (empty) return '—';
  if (!value) return '""';
  return value;
}
