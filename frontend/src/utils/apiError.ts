import type { AxiosError } from 'axios';

type ErrorBody = {
  message?: string;
  error?: string;
  code?: string;
  details?: Record<string, unknown>;
};

/** 从后端统一错误 JSON 提取用户可读文案（含可选 hint） */
export function getApiErrorMessage(err: unknown, fallback: string): string {
  if (!err || typeof err !== 'object') return fallback;
  const ax = err as AxiosError<ErrorBody>;
  const d = ax.response?.data;
  if (!d || typeof d !== 'object') return fallback;

  const msg =
    (typeof d.message === 'string' && d.message.trim()) ||
    (typeof d.error === 'string' && d.error.trim()) ||
    '';
  if (!msg) return fallback;

  const hint =
    d.details &&
    typeof d.details === 'object' &&
    typeof (d.details as { hint?: string }).hint === 'string'
      ? String((d.details as { hint: string }).hint).trim()
      : '';
  if (hint) {
    return `${msg}（${hint}）`;
  }
  return msg;
}
