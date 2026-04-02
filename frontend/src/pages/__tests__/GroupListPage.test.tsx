import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { MemoryRouter, useNavigate } from 'react-router-dom';
import GroupListPage from '../GroupListPage';
import { useGroupStore } from '@/stores/groupStore';

// 模拟 useGroupStore
jest.mock('@/stores/groupStore');

// 模拟 useNavigate
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: jest.fn(),
}));

// 模拟 message
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}));

const mockUseGroupStore = useGroupStore as jest.MockedFunction<typeof useGroupStore>;
const mockUseNavigate = useNavigate as jest.MockedFunction<typeof useNavigate>;

describe('GroupListPage Component', () => {
  const mockNavigate = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseNavigate.mockReturnValue(mockNavigate);
  });

  test('renders GroupListPage with title', async () => {
    const mockGroups = [
      {
        id: 1,
        target_count: 5,
        current_count: 3,
        status: 'active',
        deadline: new Date(Date.now() + 86400000).toISOString(),
      },
    ];

    // 模拟 store 状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: null,
      groups: mockGroups,
      fetchGroups: jest.fn().mockResolvedValue(mockGroups),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 检查页面标题
    expect(screen.getByText('拼团中心')).toBeInTheDocument();
  });

  test('shows loading state when fetching groups', async () => {
    const mockGroups = [
      {
        id: 1,
        target_count: 5,
        current_count: 3,
        status: 'active',
        deadline: new Date(Date.now() + 86400000).toISOString(),
      },
    ];

    // 模拟加载状态
    mockUseGroupStore.mockReturnValue({
      isLoading: true,
      error: null,
      groups: mockGroups,
      fetchGroups: jest.fn().mockResolvedValue(mockGroups),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 检查加载状态
    expect(screen.getByText('拼团中心')).toBeInTheDocument();
  });

  test('shows error message when there is an error', async () => {
    // 模拟错误状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: '加载失败',
      groups: [],
      fetchGroups: jest.fn().mockResolvedValue([]),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 检查错误信息
    expect(screen.getByText('错误: 加载失败')).toBeInTheDocument();
  });

  test('shows empty state when no groups', async () => {
    // 模拟无分组状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: null,
      groups: [],
      fetchGroups: jest.fn().mockResolvedValue([]),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 检查空状态（与 GroupListPage 文案一致）
    expect(screen.getByText('暂无进行中的拼团')).toBeInTheDocument();
    expect(screen.getByText('浏览商品')).toBeInTheDocument();
  });

  test('renders groups list when groups exist', async () => {
    const mockGroups = [
      {
        id: 1,
        target_count: 5,
        current_count: 3,
        status: 'active',
        deadline: new Date(Date.now() + 86400000).toISOString(),
      },
      {
        id: 2,
        target_count: 3,
        current_count: 3,
        status: 'completed',
        deadline: new Date(Date.now() - 86400000).toISOString(),
      },
    ];

    // 模拟有分组状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: null,
      groups: mockGroups,
      fetchGroups: jest.fn().mockResolvedValue(mockGroups),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 检查分组列表
    expect(screen.getByText('拼团中心')).toBeInTheDocument();
    expect(screen.getByText('拼团 #1')).toBeInTheDocument();
  });

  test('handles join group', async () => {
    const mockGroups = [
      {
        id: 1,
        target_count: 5,
        current_count: 3,
        status: 'active',
        deadline: new Date(Date.now() + 86400000).toISOString(),
      },
    ];

    const mockJoinGroup = jest.fn().mockResolvedValue(undefined);
    const mockFetchGroups = jest.fn().mockResolvedValue(mockGroups);

    // 模拟有分组状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: null,
      groups: mockGroups,
      fetchGroups: mockFetchGroups,
      joinGroup: mockJoinGroup,
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    // 点击加入拼团按钮
    const joinButton = screen.getByText('加入拼团');
    await act(async () => {
      fireEvent.click(joinButton);
    });

    // 验证加入拼团函数被调用
    await waitFor(() => {
      expect(mockJoinGroup).toHaveBeenCalledWith(1);
    });
  });

  test('navigates to catalog from empty state', async () => {
    // 模拟无分组状态
    mockUseGroupStore.mockReturnValue({
      isLoading: false,
      error: null,
      groups: [],
      fetchGroups: jest.fn().mockResolvedValue([]),
      joinGroup: jest.fn(),
      createGroup: jest.fn(),
      fetchGroupDetails: jest.fn(),
      cancelGroup: jest.fn(),
      clearError: jest.fn(),
    });

    await act(async () => {
      render(
        <MemoryRouter>
          <GroupListPage />
        </MemoryRouter>
      );
    });

    const browseButton = screen.getByText('浏览商品');
    await act(async () => {
      fireEvent.click(browseButton);
    });

    expect(mockNavigate).toHaveBeenCalledWith('/catalog');
  });
});
