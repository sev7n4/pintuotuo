import api from './api';
import type { FuelStationConfig, FuelStationTemplate } from '@/types/fuelStation';

export const fuelStationService = {
  getPublicConfig: () => api.get<{ data: FuelStationConfig }>('/catalog/fuel-station-config'),
  getAdminConfig: () => api.get<{ data: FuelStationConfig }>('/admin/fuel-station-config'),
  updateAdminConfig: (payload: FuelStationConfig) =>
    api.put<{ data: FuelStationConfig }>('/admin/fuel-station-config', payload),
  getAdminTemplates: () =>
    api.get<{ data: FuelStationTemplate[] }>('/admin/fuel-station-templates'),
  updateAdminTemplates: (templates: FuelStationTemplate[]) =>
    api.put<{ data: FuelStationTemplate[] }>('/admin/fuel-station-templates', { templates }),
};
