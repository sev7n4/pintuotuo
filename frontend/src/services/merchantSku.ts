import api from './api'
import type { APIResponse } from '@/types'
import type { MerchantSKUDetail, AvailableSKU, MerchantSKUCreateRequest, MerchantSKUUpdateRequest } from '@/types/merchantSku'

export const merchantSkuService = {
  async getMerchantSKUs(status?: string): Promise<MerchantSKUDetail[]> {
    const params = new URLSearchParams()
    if (status && status !== 'all') {
      params.append('status', status)
    }
    const response = await api.get<APIResponse<MerchantSKUDetail[]>>(`/merchants/skus?${params.toString()}`)
    return response.data.data || []
  },

  async getAvailableSKUs(provider?: string, type?: string): Promise<AvailableSKU[]> {
    const params = new URLSearchParams()
    if (provider) {
      params.append('provider', provider)
    }
    if (type) {
      params.append('type', type)
    }
    const response = await api.get<APIResponse<AvailableSKU[]>>(`/merchants/skus/available?${params.toString()}`)
    return response.data.data || []
  },

  async createMerchantSKU(data: MerchantSKUCreateRequest): Promise<MerchantSKUDetail> {
    const response = await api.post<APIResponse<MerchantSKUDetail>>('/merchants/skus', data)
    if (!response.data.data) {
      throw new Error('创建失败')
    }
    return response.data.data
  },

  async updateMerchantSKU(id: number, data: MerchantSKUUpdateRequest): Promise<MerchantSKUDetail> {
    const response = await api.put<APIResponse<MerchantSKUDetail>>(`/merchants/skus/${id}`, data)
    if (!response.data.data) {
      throw new Error('更新失败')
    }
    return response.data.data
  },

  async deleteMerchantSKU(id: number): Promise<void> {
    await api.delete(`/merchants/skus/${id}`)
  }
}
