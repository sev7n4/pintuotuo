import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { GroupProgressPage } from '../GroupProgressPage';
import { useGroupStore } from '@/stores/groupStore';
import type { Group } from '@/types';

const mockNavigate = jest.fn();

jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
  useParams: () => ({ id: '1' }),
}));

jest.mock('@/stores/groupStore', () => ({
  useGroupStore: jest.fn(),
}));

const mockGroup: Group = {
  id: 1,
  product_id: 100,
  creator_id: 1,
  target_count: 2,
  current_count: 1,
  status: 'active',
  deadline: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

describe('GroupProgressPage', () => {
  beforeEach(() => {
    mockNavigate.mockClear();
  });

  describe('UT-GP-001: renders group progress page with loading state', () => {
    it('should show loading spinner when loading', () => {
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: null,
        isLoading: true,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      const spinner = document.querySelector('.ant-spin');
      expect(spinner).toBeInTheDocument();
    });
  });

  describe('UT-GP-002: displays group info when loaded', () => {
    it('should display group information', () => {
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/拼团进度/)).toBeInTheDocument();
      expect(screen.getByText(/订单号/)).toBeInTheDocument();
    });
  });

  describe('UT-GP-003: displays member list correctly', () => {
    it('should show current and target member count', () => {
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/人成团/)).toBeInTheDocument();
    });
  });

  describe('UT-GP-004: displays countdown timer', () => {
    it('should show remaining time', () => {
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/剩余时间/)).toBeInTheDocument();
    });
  });

  describe('UT-GP-005: handles share invite action', () => {
    it('should copy share link when share button clicked', async () => {
      const mockWriteText = jest.fn();
      Object.assign(navigator, {
        clipboard: {
          writeText: mockWriteText,
        },
      });
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      const shareButton = screen.getByRole('button', { name: /分享邀请/ });
      fireEvent.click(shareButton);

      expect(mockWriteText).toHaveBeenCalled();
    });
  });

  describe('UT-GP-006: handles cancel group action', () => {
    it('should call cancelGroup when cancel button clicked and confirmed', async () => {
      const mockCancelGroup = jest.fn().mockResolvedValue(true);

      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: mockCancelGroup,
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      const cancelButton = screen.getByRole('button', { name: /取消拼团/ });
      fireEvent.click(cancelButton);

      await waitFor(() => {
        expect(screen.getByText(/确认取消拼团/)).toBeInTheDocument();
      });

      const confirmButton = screen.getByRole('button', { name: /确认取消/ });
      fireEvent.click(confirmButton);

      await waitFor(() => {
        expect(mockCancelGroup).toHaveBeenCalledWith(1);
      });
    });
  });

  describe('UT-GP-007: displays group status correctly', () => {
    it('should show active status for ongoing group', () => {
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: mockGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/进行中/)).toBeInTheDocument();
    });

    it('should show completed status for finished group', () => {
      const completedGroup = { ...mockGroup, status: 'completed' as const, current_count: 2 };
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: completedGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/拼团成功/)).toBeInTheDocument();
    });

    it('should show failed status for expired group', () => {
      const failedGroup = { ...mockGroup, status: 'failed' as const };
      (useGroupStore as unknown as jest.Mock).mockReturnValue({
        currentGroup: failedGroup,
        isLoading: false,
        error: null,
        getGroupProgress: jest.fn(),
        cancelGroup: jest.fn(),
      });

      render(
        <BrowserRouter>
          <GroupProgressPage />
        </BrowserRouter>
      );

      expect(screen.getByText(/拼团失败/)).toBeInTheDocument();
    });
  });
});
