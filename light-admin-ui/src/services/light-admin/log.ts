import { requestList } from './request';
import type { LogRow } from '@/types/light-admin/domain';

export function queryLogs(params: Record<string, unknown> = {}) {
  return requestList<LogRow>('/logs', 'GET', { params });
}
