import { render, screen, act, waitFor } from '@testing-library/react';
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

describe('Consumption', () => {
  it('renders consumption page with title', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('消费明细')).toBeInTheDocument();
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

  it('displays export button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      );
    });
    const exportButtons = screen.getAllByRole('button');
    const exportButton = exportButtons.find((btn) => btn.textContent?.includes('导出'));
    expect(exportButton).toBeInTheDocument();
  });
});
