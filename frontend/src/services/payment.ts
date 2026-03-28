import api from './api';
import { Payment, APIResponse } from '@/types';

interface InitiatePaymentRequest {
  order_id: number;
  pay_method: 'alipay' | 'wechat';
  amount: number;
}

interface PaymentCallback {
  payment_id: number;
  status: string;
  amount: number;
}

export const paymentService = {
  // Initiate payment
  initiatePayment: (data: InitiatePaymentRequest) =>
    api.post<APIResponse<Payment>>('/payments', data),

  // Get payment by ID
  getPaymentByID: (id: number) => api.get<APIResponse<Payment>>(`/payments/${id}`),

  // Refund payment
  refundPayment: (id: number) => api.post<APIResponse<Payment>>(`/payments/${id}/refund`, {}),

  // Handle Alipay callback
  handleAlipayCallback: (data: PaymentCallback) =>
    api.post<APIResponse<void>>('/payments/webhooks/alipay', data),

  // Handle WeChat callback
  handleWechatCallback: (data: PaymentCallback) =>
    api.post<APIResponse<void>>('/payments/webhooks/wechat', data),
};
