/**
 * 卖场卡片：无商品主图时，用厂商标识生成 Simple Icons CDN 徽标（与 spus.model_provider 对齐）。
 * 未收录的厂商返回 null，由 UI 回退到文字占位。
 */
const SIMPLE_ICONS_BASE = 'https://cdn.jsdelivr.net/npm/simple-icons@v11/icons';

/** model_provider / provider code → simple-icons 文件名（不含 .svg） */
const PROVIDER_ICON_SLUG: Record<string, string> = {
  openai: 'openai',
  anthropic: 'anthropic',
  google: 'google',
  gemini: 'google',
  'google-ai': 'google',
  deepseek: 'deepseek',
  meta: 'meta',
  mistral: 'mistral',
  mistralai: 'mistral',
  cohere: 'cohere',
  groq: 'groq',
  huggingface: 'huggingface',
  microsoft: 'microsoft',
  nvidia: 'nvidia',
  ibm: 'ibm',
  zhipu: 'zhipuai',
  zhipuai: 'zhipuai',
  moonshot: 'moonshotai',
  moonshotai: 'moonshotai',
  minimax: 'minimax',
  xai: 'x',
  bytedance: 'bytedance',
  alibaba: 'alibabacloud',
  alibabacloud: 'alibabacloud',
  qwen: 'alibabacloud',
  baidu: 'baidu',
  tencent: 'tencentqq',
  aws: 'amazonaws',
  amazon: 'amazonaws',
};

/** 卡片背景浅色渐变（与徽标搭配） */
const PROVIDER_SURFACE: Record<string, string> = {
  openai: 'linear-gradient(145deg, rgba(16, 163, 127, 0.12) 0%, #f5f5f5 72%)',
  anthropic: 'linear-gradient(145deg, rgba(217, 119, 87, 0.14) 0%, #f5f5f5 72%)',
  google: 'linear-gradient(145deg, rgba(66, 133, 244, 0.12) 0%, #f5f5f5 72%)',
  gemini: 'linear-gradient(145deg, rgba(66, 133, 244, 0.12) 0%, #f5f5f5 72%)',
  deepseek: 'linear-gradient(145deg, rgba(77, 107, 254, 0.12) 0%, #f5f5f5 72%)',
  meta: 'linear-gradient(145deg, rgba(8, 102, 255, 0.1) 0%, #f5f5f5 72%)',
  mistral: 'linear-gradient(145deg, rgba(250, 70, 22, 0.1) 0%, #f5f5f5 72%)',
  cohere: 'linear-gradient(145deg, rgba(217, 78, 222, 0.1) 0%, #f5f5f5 72%)',
  zhipu: 'linear-gradient(145deg, rgba(22, 119, 255, 0.1) 0%, #f5f5f5 72%)',
  zhipuai: 'linear-gradient(145deg, rgba(22, 119, 255, 0.1) 0%, #f5f5f5 72%)',
  moonshot: 'linear-gradient(145deg, rgba(99, 102, 241, 0.12) 0%, #f5f5f5 72%)',
  moonshotai: 'linear-gradient(145deg, rgba(99, 102, 241, 0.12) 0%, #f5f5f5 72%)',
  xai: 'linear-gradient(145deg, rgba(0, 0, 0, 0.06) 0%, #f5f5f5 72%)',
};

function normalizeProviderCode(code: string): string {
  return code
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '')
    .replace(/_/g, '');
}

export function resolveProviderIconSlug(providerCode: string): string | null {
  const raw = providerCode.trim().toLowerCase();
  if (PROVIDER_ICON_SLUG[raw]) {
    return PROVIDER_ICON_SLUG[raw];
  }
  const compact = normalizeProviderCode(providerCode);
  if (PROVIDER_ICON_SLUG[compact]) {
    return PROVIDER_ICON_SLUG[compact];
  }
  return null;
}

/** Simple Icons 官方 SVG（CDN）；若 slug 未配置则返回 null，避免大量 404 */
export function getProviderLogoUrl(providerCode: string): string | null {
  const slug = resolveProviderIconSlug(providerCode);
  if (!slug) return null;
  return `${SIMPLE_ICONS_BASE}/${slug}.svg`;
}

export function getProviderCardSurfaceStyle(providerCode: string): string | undefined {
  const raw = providerCode.trim().toLowerCase();
  if (PROVIDER_SURFACE[raw]) return PROVIDER_SURFACE[raw];
  const compact = normalizeProviderCode(providerCode);
  const key = Object.keys(PROVIDER_SURFACE).find((k) => normalizeProviderCode(k) === compact);
  if (key) return PROVIDER_SURFACE[key];
  return undefined;
}
