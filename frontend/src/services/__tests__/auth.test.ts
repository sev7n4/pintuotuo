import api from '../api'
import { authService } from '../auth'

jest.mock('../api')

const mockedApi = api as jest.Mocked<typeof api>

const createMockResponse = <T>(data: T) => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: { headers: {} },
})

describe('AuthService', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('register', () => {
    it('should call POST /users/register with correct data', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          user: {
            id: 1,
            email: 'test@example.com',
            name: 'Test User',
            role: 'user' as const,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          token: 'test-token',
        },
      })

      mockedApi.post.mockResolvedValueOnce(mockResponse as any)

      const result = await authService.register({
        email: 'test@example.com',
        name: 'Test User',
        password: 'password123',
      })

      expect(mockedApi.post).toHaveBeenCalledWith('/users/register', {
        email: 'test@example.com',
        name: 'Test User',
        password: 'password123',
      })
      expect(result.data.data?.token).toBe('test-token')
    })
  })

  describe('login', () => {
    it('should call POST /users/login with correct data', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
        data: {
          user: {
            id: 1,
            email: 'test@example.com',
            name: 'Test User',
            role: 'user' as const,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          token: 'test-token',
        },
      })

      mockedApi.post.mockResolvedValueOnce(mockResponse as any)

      const result = await authService.login({
        email: 'test@example.com',
        password: 'password123',
      })

      expect(mockedApi.post).toHaveBeenCalledWith('/users/login', {
        email: 'test@example.com',
        password: 'password123',
      })
      expect(result.data.data?.token).toBe('test-token')
    })
  })

  describe('logout', () => {
    it('should call POST /users/logout', async () => {
      const mockResponse = createMockResponse({
        code: 0,
        message: 'success',
      })

      mockedApi.post.mockResolvedValueOnce(mockResponse as any)

      await authService.logout()

      expect(mockedApi.post).toHaveBeenCalledWith('/users/logout')
    })
  })
})
