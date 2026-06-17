/**
 * Tenant service (platform admin). Paginated list via `requestList`;
 * CRUD mirrors the system resources (dept/role) convention.
 */
import { requestData, requestList } from './request';
import type { Tenant, TenantForm } from '@/types/light-admin/domain';

export function queryTenants(params: Record<string, unknown> = {}) {
  return requestList<Tenant>('/tenants', 'GET', { params });
}

export function createTenant(data: TenantForm) {
  return requestData<{ id: string }>('/tenants', 'POST', { data });
}

export function updateTenant(id: string, data: TenantForm) {
  return requestData<void>(`/tenants/${id}`, 'PUT', { data });
}

export function deleteTenant(id: string) {
  return requestData<void>(`/tenants/${id}`, 'DELETE');
}
