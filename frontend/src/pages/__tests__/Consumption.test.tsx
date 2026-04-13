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
  const mockFn = jest.fn(() => ({
    format: jest.fn(() => '2024-01-15'),
    subtract: jest.fn(),
  }));
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
  Area: () => null,
  Bar: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
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
});
