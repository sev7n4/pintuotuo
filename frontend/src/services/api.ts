import axios, { AxiosInstance, AxiosError, AxiosResponse } from 'axios'

const BASE_URL = '/api/v1'

const instance: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
})

instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

instance.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      localStorage.removeItem('remember_me')
      sessionStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default instance as {
  get<T = unknown>(url: string, config?: object): Promise<AxiosResponse<T>>
  post<T = unknown>(url: string, data?: unknown, config?: object): Promise<AxiosResponse<T>>
  put<T = unknown>(url: string, data?: unknown, config?: object): Promise<AxiosResponse<T>>
  delete<T = unknown>(url: string, config?: object): Promise<AxiosResponse<T>>
}
