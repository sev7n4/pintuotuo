import { render, screen, act } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { adminService } from '@/services/admin';

jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(() => ({
    user: { id: 1, name: 'Admin User', role: 'admin' },
  })),
}));

jest.mock('@/services/api', () => ({
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

jest.mock('@/services/admin', () => ({
  adminService: {
    getStats: jest.fn(),
  },
}));

describe('AdminDashboard', () => {
  let AdminDashboard: React.FC;

  beforeEach(async () => {
    jest.clearAllMocks();
    (adminService.getStats as jest.Mock).mockResolvedValue({
      data: {
        code: 0,
        message: 'success',
        data: {
          total_users: 1234,
          total_merchants: 56,
          total_orders: 892,
          total_revenue: 125680,
          pending_orders: 10,
          paid_orders: 700,
          cancelled_orders: 100,
          order_conversion_rate: 0.78,
          payment_success_rate: 0.96,
          cancellation_rate: 0.11,
          multi_item_order_ratio: 0.35,
        },
      },
    });
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

  it('displays p1 metrics cards', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminDashboard />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('订单转化率')).toBeInTheDocument();
    expect(screen.getByText('支付成功率')).toBeInTheDocument();
    expect(screen.getByText('订单取消率')).toBeInTheDocument();
    expect(screen.getByText('多明细订单占比')).toBeInTheDocument();
  });
});
