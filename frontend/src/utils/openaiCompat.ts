/**
 * OpenAI-compatible Chat Completions base URL (…/openai/v1).
 * Used by C-end「我的 Token」与开发者中心，与 GET /tokens/api-usage-guide 中 openai_compat_path 一致。
 */
export function getOpenAICompatBaseURL(): string {
  const base = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.trim() || '/api/v1';
  const normalized = base.endsWith('/') ? base.slice(0, -1) : base;
  if (normalized.startsWith('http://') || normalized.startsWith('https://')) {
    return `${normalized}/openai/v1`;
  }
  if (typeof window !== 'undefined') {
    return `${window.location.origin}${normalized}/openai/v1`;
  }
  return `${normalized}/openai/v1`;
}
