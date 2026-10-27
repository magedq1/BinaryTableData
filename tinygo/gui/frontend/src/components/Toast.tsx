import { useCallback, useEffect, useRef, useState } from 'react';
import { AlertTriangle, CheckCircle2, Info, X } from 'lucide-react';

export type ToastKind = 'success' | 'error' | 'info';

export interface ToastItem {
  id: number;
  kind: ToastKind;
  message: string;
}

let nextToastId = 1;

export function useToasts() {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const timers = useRef<Map<number, number>>(new Map());

  const dismiss = useCallback((id: number) => {
    const timer = timers.current.get(id);
    if (timer) window.clearTimeout(timer);
    timers.current.delete(id);
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const notify = useCallback(
    (message: string, kind: ToastKind = 'info') => {
      const id = nextToastId++;
      setToasts((current) => [...current.slice(-3), { id, kind, message }]);
      const timer = window.setTimeout(() => dismiss(id), kind === 'error' ? 7000 : 4200);
      timers.current.set(id, timer);
    },
    [dismiss]
  );

  useEffect(() => {
    const map = timers.current;
    return () => {
      map.forEach((timer) => window.clearTimeout(timer));
      map.clear();
    };
  }, []);

  return { toasts, notify, dismiss };
}

const tone = {
  success: {
    border: 'border-accent/35',
    icon: <CheckCircle2 size={17} className="text-accent-2" aria-hidden="true" />,
    label: 'Success'
  },
  error: {
    border: 'border-danger/40',
    icon: <AlertTriangle size={17} className="text-[#fca5a5]" aria-hidden="true" />,
    label: 'Error'
  },
  info: {
    border: 'border-line-strong',
    icon: <Info size={17} className="text-sub" aria-hidden="true" />,
    label: 'Information'
  }
} as const;

export function Toaster({ toasts, dismiss }: { toasts: ToastItem[]; dismiss: (id: number) => void }) {
  return (
    <div
      className="pointer-events-none fixed bottom-12 right-4 z-[80] flex w-[min(420px,calc(100vw-2rem))] flex-col gap-2"
      aria-live="polite"
      aria-atomic="false"
    >
      {toasts.map((toast) => {
        const style = tone[toast.kind];
        return (
          <div
            key={toast.id}
            className={`animate-toast-in pointer-events-auto flex items-start gap-3 rounded-xl border ${style.border} bg-panel/95 p-3 shadow-2xl shadow-black/30 backdrop-blur`}
            role={toast.kind === 'error' ? 'alert' : 'status'}
          >
            <div className="mt-0.5 shrink-0">{style.icon}</div>
            <div className="min-w-0 flex-1">
              <p className="sr-only">{style.label}</p>
              <p className="break-words text-[12px] leading-5 text-ink">{toast.message}</p>
            </div>
            <button
              type="button"
              aria-label="Dismiss notification"
              className="rounded-md p-1 text-sub transition-colors hover:bg-muted hover:text-ink"
              onClick={() => dismiss(toast.id)}
            >
              <X size={14} aria-hidden="true" />
            </button>
          </div>
        );
      })}
    </div>
  );
}
