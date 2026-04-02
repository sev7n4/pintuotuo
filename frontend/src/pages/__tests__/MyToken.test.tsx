import { render, screen, act, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import MyToken from '../MyToken';
import { useTokenStore } from '@/stores/tokenStore';

jest.mock('@/stores/tokenStore');

jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
    warning: jest.fn(),
  },
}));

const mockFetchBalance = jest.fn();
const mockFetchTransactions = jest.fn();
const mockFetchAPIKeys = jest.fn();
const mockCreateAPIKey = jest.fn();
const mockDeleteAPIKey = jest.fn();
const mockTransfer = jest.fn();
const mockFetchRechargeOrders = jest.fn();
const mockCreateRechargeOrder = jest.fn();
const mockMockCompleteRechargeOrder = jest.fn();
const mockClearError = jest.fn();

describe('MyToken', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (useTokenStore as unknown as jest.Mock).mockReturnValue({
      balance: 1000,
      transactions: [],
      apiKeys: [],
      rechargeOrders: [],
      fetchBalance: mockFetchBalance,
      fetchTransactions: mockFetchTransactions,
      fetchAPIKeys: mockFetchAPIKeys,
      fetchRechargeOrders: mockFetchRechargeOrders,
      createAPIKey: mockCreateAPIKey,
      deleteAPIKey: mockDeleteAPIKey,
      createRechargeOrder: mockCreateRechargeOrder,
      mockCompleteRechargeOrder: mockMockCompleteRechargeOrder,
      transfer: mockTransfer,
      isLoading: false,
      error: null,
      clearError: mockClearError,
    });
  });

  it('renders my token page with title', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('我的Token')).toBeInTheDocument();
  });

  it('displays balance statistic', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('当前余额')).toBeInTheDocument();
  });

  it('shows tabs for different sections', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('交易记录')).toBeInTheDocument();
    expect(screen.getByText('API密钥')).toBeInTheDocument();
  });

  it('fetches data on mount', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    expect(mockFetchBalance).toHaveBeenCalled();
    expect(mockFetchTransactions).toHaveBeenCalled();
    expect(mockFetchAPIKeys).toHaveBeenCalled();
  });

  it('shows transfer button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    const transferButtons = screen.getAllByRole('button');
    const transferButton = transferButtons.find((btn) => btn.textContent?.includes('转账'));
    expect(transferButton).toBeInTheDocument();
  });

  it('shows create API key button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    await act(async () => {
      fireEvent.click(screen.getByText('API密钥'));
    });

    const createButtons = screen.getAllByRole('button');
    const createButton = createButtons.find((btn) => btn.textContent?.includes('创建密钥'));
    expect(createButton).toBeInTheDocument();
  });

  it('displays empty state for transactions', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    expect(screen.getByText('暂无交易记录')).toBeInTheDocument();
  });

  it('displays empty state for API keys', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MyToken />
        </MemoryRouter>
      );
    });

    await act(async () => {
      fireEvent.click(screen.getByText('API密钥'));
    });

    expect(screen.getByText('暂无API密钥')).toBeInTheDocument();
  });
});
