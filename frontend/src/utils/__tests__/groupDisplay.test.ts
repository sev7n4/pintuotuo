import type { Group } from '@/types';
import {
  groupProductSubtitle,
  groupProgressBarStatus,
  groupStatusTagLabel,
  isGroupJoinableByState,
  isGroupPastDeadline,
} from '../groupDisplay';

const baseGroup = (over: Partial<Group>): Group => ({
  id: 1,
  creator_id: 10,
  target_count: 3,
  current_count: 1,
  status: 'active',
  deadline: new Date(Date.now() + 86400000).toISOString(),
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  ...over,
});

describe('groupDisplay', () => {
  test('isGroupPastDeadline', () => {
    expect(isGroupPastDeadline(new Date(Date.now() - 1000).toISOString())).toBe(true);
    expect(isGroupPastDeadline(new Date(Date.now() + 86400000).toISOString())).toBe(false);
  });

  test('groupStatusTagLabel shows 已截止 when active but past deadline', () => {
    const g = baseGroup({ deadline: new Date(Date.now() - 1000).toISOString(), status: 'active' });
    expect(groupStatusTagLabel(g)).toBe('已截止');
  });

  test('groupProgressBarStatus not active for past deadline', () => {
    const g = baseGroup({ deadline: new Date(Date.now() - 1000).toISOString(), status: 'active' });
    expect(groupProgressBarStatus(g)).not.toBe('active');
  });

  test('isGroupJoinableByState false when past deadline', () => {
    const g = baseGroup({ deadline: new Date(Date.now() - 1000).toISOString() });
    expect(isGroupJoinableByState(g)).toBe(false);
  });

  test('groupProductSubtitle prefers sku_name', () => {
    const g = baseGroup({ sku_name: '  Model-A  ', sku_id: 99 });
    expect(groupProductSubtitle(g)).toBe('Model-A');
  });

  test('groupProductSubtitle falls back to sku_id', () => {
    const g = baseGroup({ sku_id: 42 });
    expect(groupProductSubtitle(g)).toBe('SKU #42');
  });
});
