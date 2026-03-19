import { userService } from '../user'
import api from '../api'

// 模拟 api
jest.mock('../api')

const mockApi = api as jest.Mocked<typeof api>

describe('userService', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('getCurrentUser calls api.get with correct parameters', async () => {
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'User retrieved successfully',
    }

    mockApi.get.mockResolvedValue(mockResponse)

    const result = await userService.getCurrentUser()

    expect(mockApi.get).toHaveBeenCalledWith('/users/me')
    expect(result).toEqual(mockResponse)
  })

  test('updateCurrentUser calls api.put with correct parameters', async () => {
    const mockData = {
      name: 'Updated User',
      email: 'updated@example.com',
    }
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        email: 'updated@example.com',
        name: 'Updated User',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'User updated successfully',
    }

    mockApi.put.mockResolvedValue(mockResponse)

    const result = await userService.updateCurrentUser(mockData)

    expect(mockApi.put).toHaveBeenCalledWith('/users/me', mockData)
    expect(result).toEqual(mockResponse)
  })

  test('getUserByID calls api.get with correct parameters', async () => {
    const mockUserId = 1
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'User retrieved successfully',
    }

    mockApi.get.mockResolvedValue(mockResponse)

    const result = await userService.getUserByID(mockUserId)

    expect(mockApi.get).toHaveBeenCalledWith(`/users/${mockUserId}`)
    expect(result).toEqual(mockResponse)
  })

  test('updateUser calls api.put with correct parameters', async () => {
    const mockUserId = 1
    const mockData = {
      name: 'Updated User',
      role: 'admin',
    }
    const mockResponse = {
      success: true,
      data: {
        id: 1,
        email: 'test@example.com',
        name: 'Updated User',
        role: 'admin',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      },
      message: 'User updated successfully',
    }

    mockApi.put.mockResolvedValue(mockResponse)

    const result = await userService.updateUser(mockUserId, mockData)

    expect(mockApi.put).toHaveBeenCalledWith(`/users/${mockUserId}`, mockData)
    expect(result).toEqual(mockResponse)
  })
})
