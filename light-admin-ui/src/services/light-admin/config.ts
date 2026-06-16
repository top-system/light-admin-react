import { requestData, requestList } from './request';
import type { Config, ConfigForm } from '@/types/light-admin/domain';

export function queryConfigs(params: Record<string, unknown> = {}) {
  return requestList<Config>('/configs', 'GET', { params });
}

export function getConfigForm(id: string) {
  return requestData<ConfigForm>(`/configs/${id}/form`, 'GET');
}

export function createConfig(data: ConfigForm) {
  return requestData<void>('/configs', 'POST', { data });
}

export function updateConfig(id: string, data: ConfigForm) {
  return requestData<void>(`/configs/${id}`, 'PUT', { data });
}

export function deleteConfig(id: string) {
  return requestData<void>(`/configs/${id}`, 'DELETE');
}

export function refreshConfigCache() {
  return requestData<void>('/configs/refresh', 'PUT');
}
