# BinaryTable Studio (Wails GUI)

**English** | [العربية](README.ar.md)

BinaryTable Studio is the desktop editor for this Go/TinyGo port. It is built
with **Wails v2**, **React 19**, **TypeScript**, **Vite**, **Tailwind CSS 4**,
and **Lucide** icons. The core data work stays in the `binarytable` Go package;
the frontend is a thin, accessible shell over Wails bindings.

---

## Features

### Table editor

- Create tables from a visual schema builder (INT, LONG, DOUBLE, STRING, RAW).
- Open existing table files read/write.
- Paginated data grid with sticky headers and keyboard-friendly focus.
- Cell inspector with type-aware inputs, inline validation, and raw hex editing.
- Search any column; jump to the first match and highlight hits.
- Clear a cell or a whole row (zero-fill, preserving row capacity).
- Embed a file directly into a selected RAW cell.
- Drag a table file into the window to open it.

### Binary list editor

- Create a list with an explicit **max entries** value (the create dialog asks
  for it instead of silently using 1000).
- Open an existing list with capacity `0` (auto-detect) or an explicit value,
  including legacy/empty lists where detection is ambiguous.
- Read a compressed list directly through the Deflate stream; the view is
  read-only but export and previews still work.
- Embed multiple files at once, either with the picker or by dropping them.
- Files are streamed into the list (`AddReader`), so large blobs do not need to
  fit in memory.
- Preview images, text, and hex data.
- Export, replace, or delete a blob. Replace and delete rebuild the list
  atomically and keep the original file until the new one is ready.
- Capacity bar and per-entry sizes.
- New lists reserve the index prefix, so later opens can recover the capacity
  even when the list is empty.

### Deflate workbench

- Pack files with the Kotlin-compatible block format.
- Unpack with a configurable LRU cache; this is the GUI equivalent of
  `DeflateBinaryStreamReader`.
- Inspect block size, caller header, block counts, original length, and ratio.
- **Read the decompressed content as a table or binary list without extracting
  it to disk first.** Both views are read-only and use `DeflateReader` as their
  `binarytable.Reader`.
- Drop a file to inspect it.

### Desktop experience

- Dark OLED design system with IBM Plex Sans + JetBrains Mono.
- Command palette (`Ctrl+K`) with fuzzy filtering and keyboard navigation.
- Global drag-and-drop routing based on the active workspace.
- Native file dialogs, local-only processing, no telemetry.
- Toasts, loading states, and accessible focus rings.
- Reduced-motion support and 44px minimum interactive targets.

---

## Requirements

- Go 1.23+
- Wails CLI v2.17+
- Node.js 20+ and npm
- WebView2 (installed by default on current Windows versions)

Install Wails:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
```

---

## Development

```bash
cd tinygo/gui
wails dev
```

The first run installs frontend dependencies. `wails dev` serves Vite with hot
reload and regenerates the bindings in `frontend/wailsjs`.

## Production build

```bash
cd tinygo/gui
wails build -clean
```

The executable is written to `build/bin/BinaryTableStudio.exe` on Windows.
The build embeds `frontend/dist` and does not require Node at runtime.

## Tests

```bash
cd tinygo/gui
go test ./...

cd frontend
npm run build
```

Backend tests cover table edits, search, binary-list add/replace/delete,
capacity inference, Deflate pack/unpack, and overwrite protection.

---

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl+K` / `Cmd+K` | Command palette |
| `Ctrl+O` / `Cmd+O` | Open the active workspace file |
| `Ctrl+N` / `Cmd+N` | New table |
| `Ctrl+Shift+L` | Open a binary list |
| `Esc` | Close the active dialog |

---

## Data behaviour

- `Create table` allocates exactly `schema + rows × rowSize` bytes and truncates
  an existing file only when **Overwrite** is checked.
- `SetString` zero-fills the rest of the column.
- RAW hex editing accepts spaces and an optional `0x` prefix.
- Embedding a file into a RAW cell fails if it is larger than the column.
- Binary list offsets are 32-bit, so the practical file limit is about 2 GiB.
- File names are not part of the Kotlin `BinaryList` format; the editor shows
  indices and generated export names.
- A new list writes `realSize = 0` and reserves `100 + maxEntries × 4` bytes.
  `GuessListCapacity` uses that prefix, or the first data offset, to recover the
  capacity later.
- Compressed table/list views are decoded on demand and are always read-only.
  Use `deflate unpack` (CLI) or the unpack form to create an editable copy.

---

## Architecture

```text
gui/
├── app.go                 Wails app, dialogs, session maps
├── table_service.go       table DTOs and operations
├── list_service.go        binary-list DTOs, rewrite/delete
├── deflate_service.go     pack/unpack/inspect
├── main.go                Wails options and embed
├── frontend/
│   ├── src/App.tsx        workspace shell and actions
│   ├── src/components/    table, list, deflate, dialogs, chrome
│   ├── src/lib/api.ts     typed bridge wrapper
│   └── src/style.css      design tokens and base styles
└── build/                 Wails packaging assets
```

The frontend never touches files directly. Every operation goes through a Wails
binding that uses the same `binarytable` package as the CLI and library.

---

## Design system

Based on the UI/UX Pro Max recommendation for a developer tool:

- Style: dark OLED, high contrast, low light emission.
- Primary surface: `#0F172A`; cards: `#1B2336`; borders: `#334155`.
- Accent: `#22C55E` (code dark + run green).
- Type: IBM Plex Sans for UI, JetBrains Mono for values and schema.
- Effects: restrained glow, visible focus rings, no emoji icons.

---

## Related docs

- [CLI](../docs/cli.md)
- [Data format](../docs/format.md)
- [Deflate](../docs/compression.md)
- [TinyGo notes](../docs/tinygo.md)
