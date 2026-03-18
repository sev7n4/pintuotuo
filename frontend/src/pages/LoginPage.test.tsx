import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import LoginPage from './LoginPage'
import { useAuthStore } from '../stores/authStore'
import { message } from 'antd'

// Mock dependencies
jest.mock('../stores/authStore')
const mockNavigate = jest.fn()
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}))
jest.mock('antd', () => ({
  ...jest.requireActual('antd'),
  message: {
    success: jest.fn(),
    error: jest.fn(),
  },
}))

describe('LoginPage', () => {
  const mockLogin = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useAuthStore as unknown as jest.Mock).mockReturnValue({
      login: mockLogin,
      isLoading: false,
      error: null,
    })
  })

  it('renders login form', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    expect(screen.getByLabelText(/邮箱/)).toBeInTheDocument()
    expect(screen.getByLabelText(/密码/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /登 录/ })).toBeInTheDocument()
  })

  it('shows error messages for empty fields', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    fireEvent.click(screen.getByRole('button', { name: /登 录/ }))

    await waitFor(() => {
      expect(screen.getByText('请输入邮箱')).toBeInTheDocument()
      expect(screen.getByText('请输入密码')).toBeInTheDocument()
    })
  })

  it('submits login form with valid data', async () => {
    mockLogin.mockResolvedValueOnce(undefined)

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    fireEvent.change(screen.getByLabelText(/邮箱/), { target: { value: 'test@example.com' } })
    fireEvent.change(screen.getByLabelText(/密码/), { target: { value: 'password' } })
    fireEvent.click(screen.getByRole('button', { name: /登 录/ }))

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password')
      expect(message.success).toHaveBeenCalledWith('登录成功')
      expect(mockNavigate).toHaveBeenCalledWith('/products')
    })
  })

  it('handles login failure', async () => {
    mockLogin.mockRejectedValueOnce(new Error('Login failed'))

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    )

    fireEvent.change(screen.getByLabelText(/邮箱/), { target: { value: 'test@example.com' } })
    fireEvent.change(screen.getByLabelText(/密码/), { target: { value: 'password' } })
    fireEvent.click(screen.getByRole('button', { name: /登 录/ }))

    await waitFor(() => {
      expect(message.error).toHaveBeenCalled()
    })
  })
})
