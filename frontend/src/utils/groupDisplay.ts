import type { Group } from '@/types';

export function isGroupPastDeadline(deadline: string): boolean {
  return new Date(deadline).getTime() <= Date.now();
}

/** 仍可尝试参团（不含「已在团内」等业务规则） */
export function isGroupJoinableByState(group: Group): boolean {
  if (group.status !== 'active') return false;
  if (isGroupPastDeadline(group.deadline)) return false;
  return true;
}

/** 卡片/列表上的状态标签：优先展示「已截止」避免与 active 冲突 */
export function groupStatusTagLabel(group: Group): string {
  if (group.status === 'completed') return '已成团';
  if (group.status === 'failed') return '已失败';
  if (group.status === 'active' && isGroupPastDeadline(group.deadline)) return '已截止';
  if (group.status === 'active') return '进行中';
  return group.status;
}

export function groupProgressBarStatus(
  group: Pick<Group, 'status' | 'deadline'>
): 'active' | 'success' | 'normal' | 'exception' {
  if (group.status === 'completed') return 'success';
  if (group.status === 'failed' || (group.status === 'active' && isGroupPastDeadline(group.deadline))) {
    return 'normal';
  }
  if (group.status === 'active') return 'active';
  return 'normal';
}

/** 列表副标题：优先后端 sku_name，否则 sku_id */
export function groupProductSubtitle(group: Group): string {
  const name = group.sku_name?.trim();
  if (name) return name;
  if (group.sku_id) return `SKU #${group.sku_id}`;
  return '';
}
