/**
 * Member service (platform admin). Members are tenant-isolated, so status and
 * password operations carry the row's `tenantId` (the backend scopes by it).
 */
import { requestData, requestList } from './request';
import type { MemberRow } from '@/types/light-admin/domain';

export function queryMembers(params: Record<string, unknown> = {}) {
  return requestList<MemberRow>('/members', 'GET', { params });
}

export function setMemberStatus(
  id: string,
  data: { tenantId: string; status: number },
) {
  return requestData<void>(`/members/${id}/status`, 'PUT', { data });
}

export function resetMemberPassword(
  id: string,
  data: { tenantId: string; password: string },
) {
  return requestData<void>(`/members/${id}/password`, 'PUT', { data });
}
