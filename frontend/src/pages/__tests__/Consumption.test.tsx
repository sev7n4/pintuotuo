import React from 'react';
import { render, screen, act, waitFor, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Consumption from '../Consumption';

const mockLocalStorage = {
  getItem: jest.fn(() => 'test-token'),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage });

jest.mock('dayjs', () => {
  const create = (): Record<string, unknown> => ({
    format: jest.fn(() => '2024-01-15'),
    subtract: jest.fn(() => create()),
    diff: jest.fn(() => 43200),
  });
  const mockFn = jest.fn(() => create());
  return Object.assign(mockFn, { extend: jest.fn() });
});

jest.mock('antd', () => {
  const originalModule = jest.requireActual('antd');
  return {
    ...originalModule,
    DatePicker: {
      ...originalModule.DatePicker,
      RangePicker: () => <div data-testid="range-picker">RangePicker</div>,
    },
  };
});

jest.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  AreaChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  BarChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  ScatterChart: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Area: () => null,
  Bar: () => null,
  Scatter: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Cell: () => null,
  XAxis: () => null,
  YAxis: () => null,
  ZAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
  Legend: () => null,
}));

describe('Consumption', () => {
  it('renders consumption page with view mode toggle', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('明细列表')).toBeInTheDocument();
    expect(screen.getByText('图表视图')).toBeInTheDocument();
  });

  it('shows statistics cards', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    await waitFor(() => {
      expect(screen.getByText('总请求数')).toBeInTheDocument();
    });
  });

  it('displays refresh button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    const refreshButtons = screen.getAllByRole('button');
    const refreshButton = refreshButtons.find((btn) => btn.textContent?.includes('刷新'));
    expect(refreshButton).toBeInTheDocument();
  });

  it('displays export button in detail view', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    await act(async () => {
      fireEvent.click(screen.getByText('明细列表'));
    });
    await waitFor(() => {
      const exportButtons = screen.getAllByRole('button');
      const exportButton = exportButtons.find((btn) => btn.textContent?.includes('导出'));
      expect(exportButton).toBeInTheDocument();
    });
  });

  describe('chart view', () => {
    const origFetch = global.fetch;

    beforeEach(() => {
      global.fetch = jest.fn((url: string | Request) => {
        const u = typeof url === 'string' ? url : url.url;
        if (u.includes('/consumption/stats')) {
          return Promise.resolve({
            ok: true,
            json: async () => ({
              stats: { total_requests: 2, total_token_deduction: 200, avg_latency_ms: 80 },
              by_provider: [{ provider: 'openai', count: 2, tokens: 200 }],
              models_in_range: ['gpt-4o'],
              model_comparison: [
                {
                  provider: 'openai',
                  model: 'gpt-4o',
                  request_count: 2,
                  total_token_deduction: 200,
                  avg_token_deduction: 100,
                  latency_p50_ms: 150,
                  latency_p95_ms: 300,
                  success_rate: 1,
                },
              ],
            }),
          } as Response);
        }
        return Promise.resolve({
          ok: true,
          json: async () => ({ data: [], total: 0 }),
        } as Response);
      }) as typeof fetch;
    });

    afterEach(() => {
      global.fetch = origFetch;
    });

    it('shows model comparison card when switching to charts', async () => {
      await act(async () => {
        render(
          <MemoryRouter>
            <Consumption />
          </MemoryRouter>
        );
      });
      await waitFor(() => expect(global.fetch).toHaveBeenCalled());
      await act(async () => {
        fireEvent.click(screen.getByText('图表视图'));
      });
      expect(screen.getByText('模型对比（选购参考）')).toBeInTheDocument();
    });
  });
});
