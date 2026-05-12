/**
 * 卖场卡片：无商品主图时，用厂商标识加载 **MIT** 图标（与 spus.model_provider / model_providers.code 对齐）。
 *
 * 来源：[@lobehub/icons-static-svg](https://www.npmjs.com/package/@lobehub/icons-static-svg)（[lobe-icons](https://github.com/lobehub/lobe-icons)），
 * 覆盖智谱、阶跃、Moonshot、MiniMax 等 Simple Icons 不收录或易 404 的品牌。
 * 版本号固定，升级时改 `LOBE_ICONS_PKG_VERSION` 并 smoke 测几个 URL。
 *
 * 未映射的 code 返回 null，由 UI 回退到文字占位；后台新增厂商时在此表补一行即可。
 *
 * 生产环境曾出现的 code 一览（须逐项有映射或有意 null）：openrouter、siliconflow 及迁移种子中的各厂商等。
 */
const LOBE_ICONS_PKG_VERSION = '1.90.0';
const LOBE_ICONS_BASE = `https://unpkg.com/@lobehub/icons-static-svg@${LOBE_ICONS_PKG_VERSION}/icons`;

/** model_provider / provider code → Lobe Icons 文件名（不含 .svg） */
const PROVIDER_LOBE_SLUG: Record<string, string> = {
  openrouter: 'openrouter',
  // 生产 code siliconflow：Lobe 包无 siliconflow.svg，暂用 siliconcloud（可改 public/brand 自托管）
  siliconflow: 'siliconcloud',
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
  zhipu: 'zhipu',
  zhipuai: 'zhipu',
  zai: 'zai',
  moonshot: 'moonshot',
  moonshotai: 'moonshot',
  minimax: 'minimax',
  xai: 'xai',
  // 字节跳动 / 豆包系（迁移 model_providers.code = bytedance）
  bytedance: 'bytedance',
  // 火山引擎（Lobe：volcengine.svg）
  volcengine: 'volcengine',
  // 豆包产品线（若 SPU 单独写 doubao）
  doubao: 'doubao',
  alibaba: 'alibaba',
  alibabacloud: 'alibabacloud',
  qwen: 'qwen',
  baidu: 'baidu',
  tencent: 'tencent',
  tencentqq: 'tencent',
  aws: 'aws',
  amazon: 'aws',
  amazonaws: 'aws',
  stepfun: 'stepfun',
};

/** 卡片背景：低饱和、偏中性底，避免与深色 SVG 徽标抢对比 */
const PROVIDER_SURFACE: Record<string, string> = {
  openai:
    'radial-gradient(120% 80% at 18% 12%, rgba(16, 163, 127, 0.09) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  anthropic:
    'radial-gradient(120% 80% at 18% 12%, rgba(217, 119, 87, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  google:
    'radial-gradient(120% 80% at 18% 12%, rgba(66, 133, 244, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  gemini:
    'radial-gradient(120% 80% at 18% 12%, rgba(66, 133, 244, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  deepseek:
    'radial-gradient(120% 80% at 18% 12%, rgba(77, 107, 254, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  meta: 'radial-gradient(120% 80% at 18% 12%, rgba(8, 102, 255, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  mistral:
    'radial-gradient(120% 80% at 18% 12%, rgba(250, 70, 22, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  cohere:
    'radial-gradient(120% 80% at 18% 12%, rgba(217, 78, 222, 0.06) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  zhipu:
    'radial-gradient(120% 80% at 18% 12%, rgba(22, 119, 255, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  zhipuai:
    'radial-gradient(120% 80% at 18% 12%, rgba(22, 119, 255, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  moonshot:
    'radial-gradient(120% 80% at 18% 12%, rgba(99, 102, 241, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  moonshotai:
    'radial-gradient(120% 80% at 18% 12%, rgba(99, 102, 241, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  xai: 'radial-gradient(120% 80% at 18% 12%, rgba(0, 0, 0, 0.04) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  stepfun:
    'radial-gradient(120% 80% at 18% 12%, rgba(59, 130, 246, 0.06) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  openrouter:
    'radial-gradient(120% 80% at 18% 12%, rgba(99, 102, 241, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  siliconflow:
    'radial-gradient(120% 80% at 18% 12%, rgba(20, 184, 166, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  siliconcloud:
    'radial-gradient(120% 80% at 18% 12%, rgba(20, 184, 166, 0.08) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  bytedance:
    'radial-gradient(120% 80% at 18% 12%, rgba(254, 44, 85, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  volcengine:
    'radial-gradient(120% 80% at 18% 12%, rgba(59, 130, 246, 0.07) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
  doubao:
    'radial-gradient(120% 80% at 18% 12%, rgba(254, 44, 85, 0.06) 0%, transparent 55%), linear-gradient(165deg, #fafafa 0%, #f0f0f0 100%)',
};

function normalizeProviderCode(code: string): string {
  return code.trim().toLowerCase().replace(/\s+/g, '').replace(/_/g, '');
}

/** 返回 Lobe Icons 的图标文件名（不含扩展名），供 URL 与测试使用 */
export function resolveProviderIconSlug(providerCode: string): string | null {
  const raw = providerCode.trim().toLowerCase();
  if (PROVIDER_LOBE_SLUG[raw]) {
    return PROVIDER_LOBE_SLUG[raw];
  }
  const compact = normalizeProviderCode(providerCode);
  if (PROVIDER_LOBE_SLUG[compact]) {
    return PROVIDER_LOBE_SLUG[compact];
  }
  return null;
}

/** Lobe Icons（MIT）SVG CDN；未映射则 null，避免无效请求 */
export function getProviderLogoUrl(providerCode: string): string | null {
  const slug = resolveProviderIconSlug(providerCode);
  if (!slug) return null;
  return `${LOBE_ICONS_BASE}/${slug}.svg`;
}

export function getProviderCardSurfaceStyle(providerCode: string): string | undefined {
  const raw = providerCode.trim().toLowerCase();
  if (PROVIDER_SURFACE[raw]) return PROVIDER_SURFACE[raw];
  const compact = normalizeProviderCode(providerCode);
  const key = Object.keys(PROVIDER_SURFACE).find((k) => normalizeProviderCode(k) === compact);
  if (key) return PROVIDER_SURFACE[key];
  return undefined;
}
