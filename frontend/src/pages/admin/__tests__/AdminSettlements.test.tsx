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

/** 与生产 /api/v1/admin/settlements 列表一致：仅有 total_sales，无 total_sales_cny */
const settlementListOnlyTotalSales = {
  settlements: [
    {
      id: 1,
      merchant_id: 4,
      company_name: 'Test Co',
      period_start: '2026-04-01T00:00:00Z',
      period_end: '2026-04-30T00:00:00Z',
      total_sales: 0.001879,
      platform_fee: 0.000094,
      settlement_amount: 0.001785,
      status: 'pending',
      merchant_confirmed: false,
      finance_approved: false,
      created_at: '2026-04-11T00:00:00Z',
      updated_at: '2026-04-11T00:00:00Z',
    },
  ],
};

describe('AdminSettlements', () => {
  let AdminSettlements: React.FC;

  beforeEach(async () => {
    jest.clearAllMocks();
    mockFetch.mockImplementation((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url.includes('/admin/settlements') && !url.includes('/items')) {
        return Promise.resolve({
          ok: true,
          json: async () => settlementListOnlyTotalSales,
        });
      }
      if (url.includes('/admin/merchants')) {
        return Promise.resolve({
          ok: true,
          json: async () => ({ data: [] }),
        });
      }
      return Promise.resolve({
        ok: false,
        json: async () => ({}),
      });
    });
    AdminSettlements = (await import('../AdminSettlements')).default;
  });

  it('FE-ADMIN-SETTLEMENTS-001: 列表仅有 total_sales 时仍渲染销售总额列，不白屏', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <AdminSettlements />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('结算管理')).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByText('¥0.001879')).toBeInTheDocument();
    });
  });
});
