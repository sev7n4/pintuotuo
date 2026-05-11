import type { Group } from '@/types';

/** 解析 GET /groups 列表体：兼容根级分页与多包一层 data */
export function parseGroupsListPayload(raw: unknown): { list: Group[]; total: number } {
  let list: Group[] = [];
  let total = 0;

  if (raw && typeof raw === 'object') {
    const obj = raw as Record<string, unknown>;
    if (Array.isArray(obj.data)) {
      list = obj.data as Group[];
      total = Number(obj.total) || list.length;
    } else if (obj.data && typeof obj.data === 'object') {
      const inner = obj.data as Record<string, unknown>;
      if (Array.isArray(inner.data)) {
        list = inner.data as Group[];
        total = Number(inner.total) || list.length;
      }
    }
  }

  return { list, total };
}
