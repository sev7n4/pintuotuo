import axios, { AxiosInstance, AxiosError } from 'axios'

const BASE_URL = '/api/v1'

const instance: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
})

// Request interceptor
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor
instance.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Clear auth on unauthorized
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default instance
