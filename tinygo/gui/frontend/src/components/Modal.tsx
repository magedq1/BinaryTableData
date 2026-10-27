import { useEffect, useId, useRef, type ReactNode } from 'react';
import { X } from 'lucide-react';

export function Modal({
  open,
  title,
  description,
  onClose,
  children,
  footer,
  width = 'max-w-xl'
}: {
  open: boolean;
  title: string;
  description?: string;
  onClose: () => void;
  children: ReactNode;
  footer?: ReactNode;
  width?: string;
}) {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const descriptionId = useId();

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (open && !dialog.open) {
      dialog.showModal();
      window.setTimeout(() => {
        const focusTarget = dialog.querySelector<HTMLElement>('[data-autofocus]');
        focusTarget?.focus();
      }, 20);
    } else if (!open && dialog.open) {
      dialog.close();
    }
  }, [open]);

  return (
    <dialog
      ref={dialogRef}
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      className={`m-auto w-[calc(100vw-2rem)] ${width} rounded-2xl border border-line bg-panel p-0 text-ink shadow-2xl shadow-black/60 backdrop:bg-black/60`}
      onCancel={(event) => {
        event.preventDefault();
        onClose();
      }}
      onClick={(event) => {
        if (event.target === dialogRef.current) onClose();
      }}
    >
      <div className="flex items-start justify-between gap-4 border-b border-line px-5 py-4">
        <div>
          <h2 id={titleId} className="text-[15px] font-semibold text-ink">
            {title}
          </h2>
          {description ? (
            <p id={descriptionId} className="mt-1 text-[12px] leading-5 text-sub">
              {description}
            </p>
          ) : null}
        </div>
        <button
          type="button"
          aria-label="Close dialog"
          className="rounded-lg p-1.5 text-sub transition-colors hover:bg-muted hover:text-ink"
          onClick={onClose}
        >
          <X size={17} aria-hidden="true" />
        </button>
      </div>
      <div className="max-h-[min(70vh,620px)] overflow-y-auto px-5 py-4">{children}</div>
      {footer ? <div className="flex justify-end gap-2 border-t border-line px-5 py-3">{footer}</div> : null}
    </dialog>
  );
}
