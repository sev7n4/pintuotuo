import { parseGroupsListPayload } from '../groupListPayload';
import type { Group } from '@/types';

const g1 = { id: 1 } as Group;
const g2 = { id: 2 } as Group;

describe('parseGroupsListPayload', () => {
  it('parses Go ListGroups flat body', () => {
    expect(
      parseGroupsListPayload({
        total: 2,
        page: 1,
        per_page: 20,
        data: [g1, g2],
      })
    ).toEqual({ list: [g1, g2], total: 2 });
  });

  it('parses wrapped APIResponse shape', () => {
    expect(
      parseGroupsListPayload({
        code: 0,
        message: 'ok',
        data: {
          data: [g1],
          total: 1,
          page: 1,
          per_page: 20,
        },
      })
    ).toEqual({ list: [g1], total: 1 });
  });

  it('returns empty for invalid', () => {
    expect(parseGroupsListPayload(null)).toEqual({ list: [], total: 0 });
    expect(parseGroupsListPayload({})).toEqual({ list: [], total: 0 });
  });
});
