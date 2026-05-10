import type { APIUsageGuideItem } from '@/types';

/** 请求体 model 字段推荐写法（与权益目录一致） */
export function modelValueFromItem(it: APIUsageGuideItem): string {
  return (it.provider_slash_example || `${it.provider_code}/${it.model_example}`).trim();
}
