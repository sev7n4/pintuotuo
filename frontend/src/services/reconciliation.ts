import api from './api';

export interface LedgerReconciliation {
  usage_log_total: number;
  usage_tx_total: number;
  delta: number;
  matched: boolean;
  checked_at: string;
}

export interface UsageDriftRow {
  user_id: number;
  log_sum: number;
  tx_sum: number;
  delta: number;
}

export interface DriftListResponse {
  data: UsageDriftRow[];
  total: number;
  page: number;
  page_size: number;
  checked_at: string;
}

export interface GMVReport {
  order_count: number;
  gmv_cny: number;
  currency: string;
  start_date?: string;
  end_date?: string;
}

export const reconciliationService = {
  getLedger: () => api.get<LedgerReconciliation>('/admin/reconciliation/ledger'),

  getDrift: (params: { page?: number; page_size?: number }) =>
    api.get<DriftListResponse>('/admin/reconciliation/ledger/drift', { params }),

  exportDriftCSV: () =>
    api.get<Blob>('/admin/reconciliation/ledger/drift/export', { responseType: 'blob' }),

  postLedgerCheck: () => api.post<LedgerReconciliation>('/admin/reconciliation/ledger/check'),

  getGMV: (params?: { start_date?: string; end_date?: string }) =>
    api.get<GMVReport>('/admin/reconciliation/gmv', { params }),
};
