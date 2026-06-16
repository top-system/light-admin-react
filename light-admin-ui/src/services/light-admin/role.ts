/**
 * Role service. Assignment RPCs (getMenuIds, assignMenus) land in Task 3.5.
 */
import { requestData, requestList } from './request';
import type { Role, RoleForm, RoleOption } from '@/types/light-admin/domain';

export function queryRoles(params: Record<string, unknown> = {}) {
  return requestList<Role>('/roles', 'GET', { params });
}

export function getRoleOptions() {
  return requestData<RoleOption[]>('/roles/options', 'GET');
}

export function createRole(data: RoleForm) {
  return requestData<void>('/roles', 'POST', { data });
}

export function getRoleForm(id: string) {
  return requestData<RoleForm>(`/roles/${id}/form`, 'GET');
}

export function updateRole(id: string, data: RoleForm) {
  return requestData<void>(`/roles/${id}`, 'PUT', { data });
}

export function deleteRole(id: string) {
  return requestData<void>(`/roles/${id}`, 'DELETE');
}

export function getRoleMenuIds(id: string) {
  return requestData<string[]>(`/roles/${id}/menuIds`, 'GET');
}

export function assignRoleMenus(id: string, menuIds: string[]) {
  return requestData<void>(`/roles/${id}/menus`, 'PUT', { data: menuIds });
}
