import { useEffect, useMemo, useState, type ReactNode } from 'react';
import { Search } from 'lucide-react';
import { Modal } from './Modal';
import { cn } from './ui';

export interface CommandItem {
  id: string;
  label: string;
  hint?: string;
  group: string;
  icon: ReactNode;
  shortcut?: string;
  run: () => void;
}

export function CommandPalette({
  open,
  onClose,
  commands
}: {
  open: boolean;
  onClose: () => void;
  commands: CommandItem[];
}) {
  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);

  const filtered = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    if (!normalized) return commands;
    return commands.filter((command) => {
      const haystack = `${command.group} ${command.label} ${command.hint ?? ''}`.toLowerCase();
      return haystack.includes(normalized);
    });
  }, [commands, query]);

  useEffect(() => {
    if (open) {
      setQuery('');
      setActive(0);
    }
  }, [open]);

  useEffect(() => {
    setActive(0);
  }, [query]);

  useEffect(() => {
    if (active >= filtered.length) {
      setActive(Math.max(0, filtered.length - 1));
    }
  }, [active, filtered.length]);

  function runCommand(command: CommandItem) {
    onClose();
    window.setTimeout(command.run, 0);
  }

  return (
    <Modal open={open} title="Command palette" description="Run an editor action without leaving the keyboard." onClose={onClose} width="max-w-2xl">
      <div className="flex items-center gap-2 rounded-lg border border-line bg-bg/70 px-3 focus-within:border-accent focus-within:ring-2 focus-within:ring-accent/20">
        <Search size={15} className="text-sub" aria-hidden="true" />
        <label htmlFor="command-query" className="sr-only">
          Filter commands
        </label>
        <input
          id="command-query"
          data-autofocus
          className="h-11 min-w-0 flex-1 border-0 bg-transparent text-[13px] text-ink placeholder:text-sub/60 focus:outline-none"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'ArrowDown') {
              event.preventDefault();
              setActive((current) => Math.min(filtered.length - 1, current + 1));
            } else if (event.key === 'ArrowUp') {
              event.preventDefault();
              setActive((current) => Math.max(0, current - 1));
            } else if (event.key === 'Enter' && filtered[active]) {
              event.preventDefault();
              runCommand(filtered[active]);
            }
          }}
          placeholder="Type a command…"
          aria-controls="command-list"
          aria-activedescendant={filtered[active] ? `command-${filtered[active].id}` : undefined}
        />
        <kbd className="rounded border border-line px-1.5 py-0.5 font-mono text-[9px] text-sub">ESC</kbd>
      </div>

      <div id="command-list" role="listbox" aria-label="Commands" className="mt-2 max-h-[380px] overflow-y-auto">
        {filtered.length === 0 ? (
          <p className="px-3 py-6 text-center text-[12px] text-sub">No command matches this search.</p>
        ) : (
          filtered.map((command, index) => (
            <button
              key={command.id}
              id={`command-${command.id}`}
              type="button"
              role="option"
              aria-selected={index === active}
              className={cn(
                'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors',
                index === active ? 'bg-accent/10 text-ink' : 'text-sub hover:bg-muted/60 hover:text-ink'
              )}
              onMouseEnter={() => setActive(index)}
              onClick={() => runCommand(command)}
            >
              <span className={cn('flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border', index === active ? 'border-accent/30 bg-accent/10 text-accent-2' : 'border-line bg-bg/40 text-sub')}>
                {command.icon}
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-[12px] font-medium">{command.label}</span>
                <span className="block truncate text-[10px] text-sub/70">{command.hint ?? command.group}</span>
              </span>
              {command.shortcut ? <kbd className="rounded border border-line px-1.5 py-0.5 font-mono text-[9px] text-sub">{command.shortcut}</kbd> : null}
            </button>
          ))
        )}
      </div>
    </Modal>
  );
}
