import { render, screen, fireEvent, act, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import Consumption from '../Consumption'

// Mock fetch
const mockFetch = jest.fn()
global.fetch = mockFetch

// Mock localStorage
const mockLocalStorage = {
  getItem: jest.fn(() => 'test-token'),
  setItem: jest.fn(),
  removeItem: jest.fn(),
}
Object.defineProperty(window, 'localStorage', { value: mockLocalStorage })

// Mock dayjs properly
jest.mock('dayjs', () => {
  const dayjs = jest.fn((date?: string) => ({
    format: jest.fn(() => '2024-01-15'),
    subtract: jest.fn(() => dayjs()),
  }))
  dayjs.extend = jest.fn()
  return dayjs
})

describe('Consumption', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    mockFetch.mockReset()
  })

  it('renders consumption page with title', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { total_requests: 0, total_tokens: 0, total_cost: 0, avg_latency_ms: 0 } }),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      )
    })

    expect(screen.getByText('消费记录')).toBeInTheDocument()
  })

  it('shows statistics cards', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { total_requests: 10, total_tokens: 1000, total_cost: 0.5, avg_latency_ms: 1200 } }),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      )
    })

    await waitFor(() => {
      expect(screen.getByText('总请求数')).toBeInTheDocument()
    })
  })

  it('displays refresh button', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { total_requests: 0, total_tokens: 0, total_cost: 0, avg_latency_ms: 0 } }),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      )
    })

    const refreshButtons = screen.getAllByRole('button')
    const refreshButton = refreshButtons.find(btn => btn.textContent?.includes('刷新'))
    expect(refreshButton).toBeInTheDocument()
  })

  it('displays export button', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { total_requests: 0, total_tokens: 0, total_cost: 0, avg_latency_ms: 0 } }),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      )
    })

    const exportButtons = screen.getAllByRole('button')
    const exportButton = exportButtons.find(btn => btn.textContent?.includes('导出'))
    expect(exportButton).toBeInTheDocument()
  })

  it('fetches data on mount', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: [] }),
    })
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ data: { total_requests: 0, total_tokens: 0, total_cost: 0, avg_latency_ms: 0 } }),
    })

    await act(async () => {
      render(
        <MemoryRouter>
          <Consumption />
        </MemoryRouter>
      )
    })

    expect(mockFetch).toHaveBeenCalled()
  })
})
