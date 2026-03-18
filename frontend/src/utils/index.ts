// Common utility functions

export const formatPrice = (price: number): string => {
  return `¥${price.toFixed(2)}`
}

export const formatDate = (date: string): string => {
  return new Date(date).toLocaleDateString('zh-CN')
}

export const formatDateTime = (date: string): string => {
  return new Date(date).toLocaleString('zh-CN')
}

export const getStatusLabel = (status: string): string => {
  const labels: Record<string, string> = {
    'pending': '待支付',
    'paid': '已支付',
    'completed': '已完成',
    'failed': '失败',
    'active': '上架',
    'inactive': '下架',
    'archived': '归档',
  }
  return labels[status] || status
}

export const delay = (ms: number): Promise<void> => {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
