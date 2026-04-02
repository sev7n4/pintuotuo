import { groupService } from '../group';
import api from '../api';
import type { AxiosResponse } from 'axios';

jest.mock('../api');

const mockApi = api as jest.Mocked<typeof api>;

const createMockResponse = <T>(data: T): AxiosResponse<T> => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
});

describe('groupService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('createGroup calls api.post with correct parameters', async () => {
    const mockData = {
      sku_id: 1,
      target_count: 5,
      deadline: '2024-12-31T23:59:59Z',
    };
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        target_count: 5,
        current_count: 1,
        deadline: '2024-12-31T23:59:59Z',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Group created successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.createGroup(mockData);

    expect(mockApi.post).toHaveBeenCalledWith('/groups', mockData);
    expect(result.data).toEqual(mockResponse);
  });

  test('listGroups calls api.get with correct parameters', async () => {
    const mockPage = 1;
    const mockPerPage = 10;
    const mockResponse = {
      success: true,
      data: {
        items: [
          {
            id: 1,
            product_id: 1,
            target_count: 5,
            current_count: 1,
            deadline: '2024-12-31T23:59:59Z',
            status: 'active',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        page: 1,
        per_page: 10,
      },
      message: 'Groups retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.listGroups(mockPage, mockPerPage);

    expect(mockApi.get).toHaveBeenCalledWith('/groups', {
      params: { page: mockPage, per_page: mockPerPage },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('listGroups calls api.get without parameters when page and per_page are not provided', async () => {
    const mockResponse = {
      success: true,
      data: {
        items: [],
        total: 0,
        page: 1,
        per_page: 10,
      },
      message: 'Groups retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.listGroups();

    expect(mockApi.get).toHaveBeenCalledWith('/groups', {
      params: { page: undefined, per_page: undefined },
    });
    expect(result.data).toEqual(mockResponse);
  });

  test('getGroupByID calls api.get with correct parameters', async () => {
    const mockGroupId = 1;
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        target_count: 5,
        current_count: 1,
        deadline: '2024-12-31T23:59:59Z',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Group retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.getGroupByID(mockGroupId);

    expect(mockApi.get).toHaveBeenCalledWith(`/groups/${mockGroupId}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('joinGroup calls api.post with correct parameters', async () => {
    const mockGroupId = 1;
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        target_count: 5,
        current_count: 2,
        deadline: '2024-12-31T23:59:59Z',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Joined group successfully',
    };

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.joinGroup(mockGroupId);

    expect(mockApi.post).toHaveBeenCalledWith(`/groups/${mockGroupId}/join`, {});
    expect(result.data).toEqual(mockResponse);
  });

  test('cancelGroup calls api.delete with correct parameters', async () => {
    const mockGroupId = 1;
    const mockResponse = {
      success: true,
      data: undefined,
      message: 'Group cancelled successfully',
    };

    mockApi.delete.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.cancelGroup(mockGroupId);

    expect(mockApi.delete).toHaveBeenCalledWith(`/groups/${mockGroupId}`);
    expect(result.data).toEqual(mockResponse);
  });

  test('getGroupProgress calls api.get with correct parameters', async () => {
    const mockGroupId = 1;
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        product_id: 1,
        target_count: 5,
        current_count: 3,
        deadline: '2024-12-31T23:59:59Z',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'Group progress retrieved successfully',
    };

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse));

    const result = await groupService.getGroupProgress(mockGroupId);

    expect(mockApi.get).toHaveBeenCalledWith(`/groups/${mockGroupId}/progress`);
    expect(result.data).toEqual(mockResponse);
  });
});
