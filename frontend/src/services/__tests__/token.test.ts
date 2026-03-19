import { tokenService } from '../token'
import api from '../api'
import type { AxiosResponse } from 'axios'

jest.mock('../api')

const mockApi = api as jest.Mocked<typeof api>

const createMockResponse = <T>(data: T): AxiosResponse<T> => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
})

describe('tokenService', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('getBalance calls api.get with correct parameters', async () => {
    const mockResponse = {
      id: 1,
      user_id: 1,
      balance: 1000,
      frozen_balance: 100,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.getBalance()

    expect(mockApi.get).toHaveBeenCalledWith('/tokens/balance')
    expect(result.data).toEqual(mockResponse)
  })

  test('getConsumption calls api.get with correct parameters', async () => {
    const mockResponse = [
      {
        id: 1,
        user_id: 1,
        amount: -100,
        type: 'consumption',
        description: 'Purchase product',
        created_at: '2024-01-01T00:00:00Z',
      },
    ]

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.getConsumption()

    expect(mockApi.get).toHaveBeenCalledWith('/tokens/consumption')
    expect(result.data).toEqual(mockResponse)
  })

  test('transfer calls api.post with correct parameters', async () => {
    const mockRecipientId = 2
    const mockAmount = 100
    const mockResponse = {
      message: 'Transfer successful',
    }

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.transfer(mockRecipientId, mockAmount)

    expect(mockApi.post).toHaveBeenCalledWith('/tokens/transfer', { recipient_id: mockRecipientId, amount: mockAmount })
    expect(result.data).toEqual(mockResponse)
  })

  test('getAPIKeys calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: [
        {
          id: 1,
          user_id: 1,
          name: 'Test API Key',
          api_key: 'test-api-key',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      ],
      message: 'API keys retrieved successfully',
    }

    mockApi.get.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.getAPIKeys()

    expect(mockApi.get).toHaveBeenCalledWith('/tokens/keys')
    expect(result.data).toEqual(mockResponse)
  })

  test('createAPIKey calls api.post with correct parameters', async () => {
    const mockName = 'Test API Key'
    const mockResponse = {
      id: 1,
      user_id: 1,
      name: 'Test API Key',
      api_key: 'test-api-key',
      status: 'active',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    mockApi.post.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.createAPIKey(mockName)

    expect(mockApi.post).toHaveBeenCalledWith('/tokens/keys', { name: mockName })
    expect(result.data).toEqual(mockResponse)
  })

  test('updateAPIKey calls api.put with correct parameters', async () => {
    const mockApiKeyId = 1
    const mockData = {
      name: 'Updated API Key',
      status: 'inactive' as const,
    }
    const mockResponse = {
      id: 1,
      user_id: 1,
      name: 'Updated API Key',
      api_key: 'test-api-key',
      status: 'inactive',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    }

    mockApi.put.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.updateAPIKey(mockApiKeyId, mockData)

    expect(mockApi.put).toHaveBeenCalledWith(`/tokens/keys/${mockApiKeyId}`, mockData)
    expect(result.data).toEqual(mockResponse)
  })

  test('deleteAPIKey calls api.delete with correct parameters', async () => {
    const mockApiKeyId = 1
    const mockResponse = {
      success: true,
      message: 'API key deleted successfully',
    }

    mockApi.delete.mockResolvedValue(createMockResponse(mockResponse))

    const result = await tokenService.deleteAPIKey(mockApiKeyId)

    expect(mockApi.delete).toHaveBeenCalledWith(`/tokens/keys/${mockApiKeyId}`)
    expect(result.data).toEqual(mockResponse)
  })
})
