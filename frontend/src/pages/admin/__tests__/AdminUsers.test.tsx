import { render, screen, act, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const mockFetch = jest.fn();
global.fetch = mockFetch;

const mockLocalStorage = {
  getItem: jest.fn(() => 'mock-token'),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

describe('AdminUsers', () => {
  let AdminUsers: React.FC;

  beforeEach(async () => {
    jest.clearAllMocks();
    mockFetch.mockResolvedValue({
      json: async () => ({
        code: 0,
        data: [
          {
            id: 1,
            email: 'admin@example.com',
            name: 'Admin',
            role: 'admin',
            created_at: '2024-01-01T00:00:00Z',
          },
          {
            id: 2,
            email: 'merchant@example.com',
            name: 'Merchant',
            role: 'merchant',
            created_at: '2024-01-02T00:00:00Z',
          },
          {
            id: 3,
            email: 'user@example.com',
            name: 'User',
            role: 'user',
            created_at: '2024-01-03T00:00:00Z',
          },
        ],
      }),
    });
    AdminUsers = (await import('../AdminUsers')).default;
  });

  it('renders users management page', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminUsers />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('用户管理')).toBeInTheDocument();
  });

  it('displays create admin button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminUsers />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('创建管理员')).toBeInTheDocument();
  });

  it('displays user table columns', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminUsers />
        </MemoryRouter>
      );
    });
    await waitFor(() => {
      expect(screen.getByText('ID')).toBeInTheDocument();
      expect(screen.getByText('邮箱')).toBeInTheDocument();
      expect(screen.getByText('用户名')).toBeInTheDocument();
      expect(screen.getByText('角色')).toBeInTheDocument();
    });
  });

  it('displays role tags', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminUsers />
        </MemoryRouter>
      );
    });
    await waitFor(() => {
      expect(screen.getByText('管理员')).toBeInTheDocument();
      expect(screen.getByText('商户')).toBeInTheDocument();
      expect(screen.getByText('用户')).toBeInTheDocument();
    });
  });
});
