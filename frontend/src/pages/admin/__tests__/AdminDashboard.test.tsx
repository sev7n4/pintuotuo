import { render, screen, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(() => ({
    user: { id: 1, name: 'Admin User', role: 'admin' },
  })),
}));

jest.mock('@/services/api', () => ({
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

describe('AdminDashboard', () => {
  let AdminDashboard: React.FC;

  beforeEach(async () => {
    jest.clearAllMocks();
    AdminDashboard = (await import('../AdminDashboard')).default;
  });

  it('renders dashboard title', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('平台运营概览')).toBeInTheDocument();
  });

  it('displays total users statistic', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('总用户数')).toBeInTheDocument();
  });

  it('displays merchant count statistic', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('商户数量')).toBeInTheDocument();
  });

  it('displays total orders statistic', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('总订单数')).toBeInTheDocument();
  });

  it('displays platform revenue statistic', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('平台总收入')).toBeInTheDocument();
  });
});
