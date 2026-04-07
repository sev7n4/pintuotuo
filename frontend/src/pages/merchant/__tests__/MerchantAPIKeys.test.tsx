import { render, screen, act, within, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

jest.mock('@/stores/merchantStore', () => ({
  useMerchantStore: jest.fn(() => ({
    apiKeys: [
      {
        id: 1,
        name: 'Production Key',
        provider: 'openai',
        quota_limit: 100,
        status: 'active',
        last_used_at: '2024-01-15T10:30:00Z',
        created_at: '2024-01-01T00:00:00Z',
      },
      {
        id: 2,
        name: 'Test Key',
        provider: 'anthropic',
        quota_limit: 0,
        status: 'inactive',
        last_used_at: null,
        created_at: '2024-01-10T00:00:00Z',
      },
    ],
    apiKeyUsage: [
      {
        id: 1,
        quota_used: 45.5,
        quota_limit: 100,
        usage_percentage: 45.5,
      },
    ],
    fetchAPIKeys: jest.fn(),
    fetchAPIKeyUsage: jest.fn(),
    createAPIKey: jest.fn(),
    updateAPIKey: jest.fn(),
    deleteAPIKey: jest.fn(),
    isLoading: false,
  })),
}));

jest.mock('@/stores/authStore', () => ({
  useAuthStore: jest.fn(() => ({
    user: { id: 1, name: 'Test Merchant', role: 'merchant' },
  })),
}));

jest.mock('@/services/api', () => ({
  default: { get: jest.fn(), post: jest.fn(), put: jest.fn(), delete: jest.fn() },
}));

const mockProvidersPayload = {
  data: {
    data: [
      {
        id: 1,
        code: 'openai',
        name: 'OpenAI',
        api_format: 'openai',
        status: 'active',
        sort_order: 0,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        cache_enabled: false,
      },
      {
        id: 2,
        code: 'anthropic',
        name: 'Anthropic',
        api_format: 'anthropic',
        status: 'active',
        sort_order: 1,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        cache_enabled: false,
      },
    ],
  },
};

jest.mock('@/services/merchant', () => ({
  merchantService: {
    getMerchantModelProviders: jest.fn(),
  },
}));

describe('MerchantAPIKeys', () => {
  let MerchantAPIKeys: React.FC;

  jest.setTimeout(20_000);

  beforeEach(async () => {
    jest.clearAllMocks();
    const { merchantService } = require('@/services/merchant');
    merchantService.getMerchantModelProviders.mockResolvedValue(mockProvidersPayload);
    MerchantAPIKeys = (await import('../MerchantAPIKeys')).default;
  });

  it('renders API keys page with title', async () => {
    render(
      <MemoryRouter>
        <MerchantAPIKeys />
      </MemoryRouter>
    );
    await waitFor(() => {
      expect(screen.getByText('API密钥管理')).toBeInTheDocument();
    });
  });

  it('displays add key button', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantAPIKeys />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('添加密钥')).toBeInTheDocument();
  });

  it('displays API key table with data', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantAPIKeys />
        </MemoryRouter>
      );
    });
    const table = screen.getByRole('table');
    expect(within(table).getAllByText('名称').length).toBeGreaterThan(0);
    expect(within(table).getAllByText('提供商').length).toBeGreaterThan(0);
    expect(screen.getByText('Production Key')).toBeInTheDocument();
    expect(screen.getByText('Test Key')).toBeInTheDocument();
  });

  it('displays provider tags', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantAPIKeys />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('OPENAI')).toBeInTheDocument();
    expect(screen.getByText('ANTHROPIC')).toBeInTheDocument();
  });

  it('displays status tags', async () => {
    await act(async () => {
      render(
        <MemoryRouter>
          <MerchantAPIKeys />
        </MemoryRouter>
      );
    });
    expect(screen.getByText('启用')).toBeInTheDocument();
    expect(screen.getByText('禁用')).toBeInTheDocument();
  });
});
