/**
 * Dept service. Backend returns a tree; `/depts/options` is the dropdown
 * shape used by user forms and anywhere a dept tree-select is needed.
 */
import { requestData } from './request';
import type { Dept, DeptForm, DeptOption } from '@/types/light-admin/domain';

export function queryDepts(params: Record<string, unknown> = {}) {
  return requestData<Dept[]>('/depts', 'GET', { params });
}

export function getDeptOptions() {
  return requestData<DeptOption[]>('/depts/options', 'GET');
}

export function getDeptForm(id: string) {
  return requestData<DeptForm>(`/depts/${id}/form`, 'GET');
}

export function createDept(data: DeptForm) {
  return requestData<void>('/depts', 'POST', { data });
}

export function updateDept(id: string, data: DeptForm) {
  return requestData<void>(`/depts/${id}`, 'PUT', { data });
}

export function deleteDepts(ids: string[]) {
  return requestData<void>(`/depts/${ids.join(',')}`, 'DELETE');
}
