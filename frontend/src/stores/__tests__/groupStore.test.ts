import { useGroupStore } from '../groupStore';
import { groupService } from '@/services/group';
import { Group } from '@/types';

jest.mock('@/services/group');

const mockedGroupService = groupService as jest.Mocked<typeof groupService>;

const mockGroup: Group = {
  id: 1,
  product_id: 1,
  creator_id: 1,
  current_count: 2,
  target_count: 5,
  status: 'active',
  deadline: '2024-12-31T23:59:59Z',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const mockGroup2: Group = {
  id: 2,
  product_id: 2,
  creator_id: 2,
  current_count: 3,
  target_count: 4,
  status: 'active',
  deadline: '2024-12-31T23:59:59Z',
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

describe('GroupStore', () => {
  beforeEach(() => {
    useGroupStore.setState({
      groups: [],
      currentGroup: null,
      total: 0,
      isLoading: false,
      error: null,
    });
    jest.clearAllMocks();
  });

  describe('initial state', () => {
    it('should have correct initial values', () => {
      const state = useGroupStore.getState();
      expect(state.groups).toEqual([]);
      expect(state.currentGroup).toBeNull();
      expect(state.total).toBe(0);
      expect(state.isLoading).toBe(false);
      expect(state.error).toBeNull();
    });
  });

  describe('fetchGroups', () => {
    it('should fetch groups successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: {
            data: [mockGroup, mockGroup2],
            total: 2,
            page: 1,
            per_page: 20,
          },
        },
      };

      mockedGroupService.listGroups.mockResolvedValueOnce(mockResponse as any);

      const store = useGroupStore.getState();
      const result = await store.fetchGroups();

      expect(result).toEqual([mockGroup, mockGroup2]);
      const newState = useGroupStore.getState();
      expect(newState.groups).toEqual([mockGroup, mockGroup2]);
      expect(newState.total).toBe(2);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch groups error', async () => {
      const errorMessage = 'Failed to fetch groups';
      mockedGroupService.listGroups.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      const result = await store.fetchGroups();

      expect(result).toBeNull();
      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('fetchGroupByID', () => {
    it('should fetch group by ID successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: mockGroup,
        },
      };

      mockedGroupService.getGroupByID.mockResolvedValueOnce(mockResponse as any);

      const store = useGroupStore.getState();
      await store.fetchGroupByID(1);

      const newState = useGroupStore.getState();
      expect(newState.currentGroup).toEqual(mockGroup);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle fetch group by ID error', async () => {
      const errorMessage = 'Failed to fetch group';
      mockedGroupService.getGroupByID.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      await store.fetchGroupByID(1);

      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('createGroup', () => {
    it('should create group successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: { group: mockGroup, order_id: 100 },
        },
      };

      mockedGroupService.createGroup.mockResolvedValueOnce(mockResponse as any);

      const store = useGroupStore.getState();
      await store.createGroup(1, 5, '2024-12-31T23:59:59Z');

      const newState = useGroupStore.getState();
      expect(newState.groups).toEqual([mockGroup]);
      expect(newState.currentGroup).toEqual(mockGroup);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle create group error', async () => {
      const errorMessage = 'Failed to create group';
      mockedGroupService.createGroup.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      await expect(store.createGroup(1, 5, '2024-12-31T23:59:59Z')).rejects.toThrow(errorMessage);

      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('joinGroup', () => {
    it('should join group successfully', async () => {
      // 先添加一个分组
      useGroupStore.setState({ groups: [mockGroup] });

      const updatedGroup = { ...mockGroup, current_count: 3 };
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: { group: updatedGroup, order_id: 200 },
        },
      };

      mockedGroupService.joinGroup.mockResolvedValueOnce(mockResponse as any);

      const store = useGroupStore.getState();
      await store.joinGroup(1);

      const newState = useGroupStore.getState();
      expect(newState.groups[0].current_count).toBe(3);
      expect(newState.currentGroup).toEqual(updatedGroup);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle join group error', async () => {
      const errorMessage = 'Failed to join group';
      mockedGroupService.joinGroup.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      await expect(store.joinGroup(1)).rejects.toThrow(errorMessage);

      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('cancelGroup', () => {
    it('should cancel group successfully', async () => {
      // 先添加一个分组
      useGroupStore.setState({ groups: [mockGroup, mockGroup2] });

      mockedGroupService.cancelGroup.mockResolvedValueOnce({
        data: { code: 0, message: 'success' },
      } as any);

      const store = useGroupStore.getState();
      await store.cancelGroup(1);

      const newState = useGroupStore.getState();
      expect(newState.groups).toEqual([mockGroup2]);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle cancel group error', async () => {
      const errorMessage = 'Failed to cancel group';
      mockedGroupService.cancelGroup.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      await expect(store.cancelGroup(1)).rejects.toThrow(errorMessage);

      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('getGroupProgress', () => {
    it('should get group progress successfully', async () => {
      const mockResponse = {
        data: {
          code: 0,
          message: 'success',
          data: mockGroup,
        },
      };

      mockedGroupService.getGroupProgress.mockResolvedValueOnce(mockResponse as any);

      const store = useGroupStore.getState();
      await store.getGroupProgress(1);

      const newState = useGroupStore.getState();
      expect(newState.currentGroup).toEqual(mockGroup);
      expect(newState.isLoading).toBe(false);
    });

    it('should handle get group progress error', async () => {
      const errorMessage = 'Failed to get group progress';
      mockedGroupService.getGroupProgress.mockRejectedValueOnce(new Error(errorMessage));

      const store = useGroupStore.getState();
      await store.getGroupProgress(1);

      const newState = useGroupStore.getState();
      expect(newState.error).toBe(errorMessage);
      expect(newState.isLoading).toBe(false);
    });
  });

  describe('clearError', () => {
    it('should clear error', () => {
      useGroupStore.setState({ error: 'Test error' });

      const store = useGroupStore.getState();
      store.clearError();

      // 重新获取 store 状态来验证错误是否被清除
      const updatedStore = useGroupStore.getState();
      expect(updatedStore.error).toBeNull();
    });
  });
});
