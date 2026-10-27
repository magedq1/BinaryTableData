import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { LoaderCircle } from 'lucide-react';

export function cn(...values: Array<string | false | null | undefined>): string {
  return values.filter(Boolean).join(' ');
}

const buttonVariants = {
  primary:
    'bg-accent text-[#052e16] hover:bg-accent-2 active:bg-[#16a34a] shadow-[0_0_24px_rgba(34,197,94,0.12)]',
  secondary: 'bg-muted text-ink hover:bg-[#313b52] border border-line',
  ghost: 'bg-transparent text-sub hover:text-ink hover:bg-muted/70 border border-transparent',
  danger: 'bg-danger/15 text-[#fecaca] hover:bg-danger/25 border border-danger/40'
} as const;

const buttonSizes = {
  sm: 'h-9 px-3 text-[12px] gap-1.5',
  md: 'h-10 px-4 text-[13px] gap-2'
} as const;

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: keyof typeof buttonVariants;
  size?: keyof typeof buttonSizes;
  loading?: boolean;
  icon?: ReactNode;
}

export function Button({
  variant = 'secondary',
  size = 'md',
  loading = false,
  icon,
  className,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      type="button"
      className={cn(
        'inline-flex items-center justify-center rounded-lg font-medium transition-colors duration-150',
        'disabled:cursor-not-allowed disabled:opacity-45',
        buttonVariants[variant],
        buttonSizes[size],
        className
      )}
      disabled={disabled || loading}
      {...props}
    >
      {loading ? <LoaderCircle size={15} className="animate-spin" aria-hidden="true" /> : icon}
      {children}
    </button>
  );
}

export interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  label: string;
  children: ReactNode;
  active?: boolean;
}

export function IconButton({ label, children, active, className, ...props }: IconButtonProps) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      className={cn(
        'inline-flex h-10 w-10 items-center justify-center rounded-lg border transition-colors duration-150',
        active
          ? 'border-accent/50 bg-accent/12 text-accent-2'
          : 'border-transparent bg-transparent text-sub hover:bg-muted/70 hover:text-ink',
        'disabled:cursor-not-allowed disabled:opacity-40',
        className
      )}
      {...props}
    >
      {children}
    </button>
  );
}

export const inputClass =
  'h-10 w-full rounded-lg border border-line bg-bg/70 px-3 text-[13px] text-ink placeholder:text-sub/60 ' +
  'transition-colors duration-150 hover:border-line-strong focus:border-accent focus:outline-none ' +
  'focus:ring-2 focus:ring-accent/25 disabled:opacity-50';

export const selectClass =
  'h-10 w-full rounded-lg border border-line bg-bg/70 px-3 text-[13px] text-ink ' +
  'transition-colors duration-150 hover:border-line-strong focus:border-accent focus:outline-none ' +
  'focus:ring-2 focus:ring-accent/25';

export const textareaClass =
  'w-full resize-y rounded-lg border border-line bg-bg/70 px-3 py-2 font-mono text-[12px] leading-6 ' +
  'text-ink placeholder:text-sub/60 transition-colors duration-150 hover:border-line-strong ' +
  'focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/25';

export function Field({
  label,
  hint,
  error,
  htmlFor,
  children
}: {
  label: string;
  hint?: string;
  error?: string;
  htmlFor?: string;
  children: ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <label htmlFor={htmlFor} className="block text-[12px] font-medium text-sub">
        {label}
      </label>
      {children}
      {error ? (
        <p role="alert" className="text-[11px] text-[#fca5a5]">
          {error}
        </p>
      ) : hint ? (
        <p className="text-[11px] text-sub/80">{hint}</p>
      ) : null}
    </div>
  );
}

export function Badge({
  children,
  tone = 'neutral'
}: {
  children: ReactNode;
  tone?: 'neutral' | 'accent' | 'warn' | 'danger' | 'muted';
}) {
  const tones = {
    neutral: 'border-line bg-muted/60 text-sub',
    accent: 'border-accent/40 bg-accent/10 text-accent-2',
    warn: 'border-warn/40 bg-warn/10 text-[#fcd34d]',
    danger: 'border-danger/40 bg-danger/10 text-[#fca5a5]',
    muted: 'border-line/70 bg-bg/50 text-sub'
  } as const;
  return (
    <span className={cn('inline-flex items-center gap-1 rounded-md border px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide', tones[tone])}>
      {children}
    </span>
  );
}

export function Stat({
  label,
  value,
  detail
}: {
  label: string;
  value: ReactNode;
  detail?: ReactNode;
}) {
  return (
    <div className="rounded-lg border border-line bg-panel-2/60 px-3 py-2">
      <div className="text-[10px] uppercase tracking-[0.14em] text-sub">{label}</div>
      <div className="mt-0.5 font-mono text-[13px] text-ink">{value}</div>
      {detail ? <div className="mt-0.5 text-[10px] text-sub">{detail}</div> : null}
    </div>
  );
}

export function EmptyState({
  icon,
  title,
  body,
  actions
}: {
  icon: ReactNode;
  title: string;
  body: string;
  actions?: ReactNode;
}) {
  return (
    <div className="flex h-full items-center justify-center p-8">
      <div className="max-w-lg text-center">
        <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl border border-line bg-panel-2 text-accent-2">
          {icon}
        </div>
        <h2 className="mt-5 text-lg font-semibold text-ink">{title}</h2>
        <p className="mt-1.5 text-[13px] leading-6 text-sub">{body}</p>
        {actions ? <div className="mt-5 flex flex-wrap items-center justify-center gap-2">{actions}</div> : null}
      </div>
    </div>
  );
}
