import { requestData, requestList } from './request';
import type {
  Notice,
  NoticeDetail,
  NoticeForm,
  UserNotice,
} from '@/types/light-admin/domain';

export function queryNotices(params: Record<string, unknown> = {}) {
  return requestList<Notice>('/notices', 'GET', { params });
}

export function getNoticeForm(id: number) {
  return requestData<NoticeForm>(`/notices/${id}/form`, 'GET');
}

export function getNoticeDetail(id: number) {
  return requestData<NoticeDetail>(`/notices/${id}/detail`, 'GET');
}

export function createNotice(data: NoticeForm) {
  return requestData<void>('/notices', 'POST', { data });
}

export function updateNotice(id: number, data: NoticeForm) {
  return requestData<void>(`/notices/${id}`, 'PUT', { data });
}

export function deleteNotices(ids: number[]) {
  return requestData<void>(`/notices/${ids.join(',')}`, 'DELETE');
}

export function publishNotice(id: number) {
  return requestData<void>(`/notices/${id}/publish`, 'PUT');
}

export function revokeNotice(id: number) {
  return requestData<void>(`/notices/${id}/revoke`, 'PUT');
}

export function listMyNotices(params: Record<string, unknown> = {}) {
  return requestList<UserNotice>('/notices/my', 'GET', { params });
}

export function readAllNotices() {
  return requestData<void>('/notices/read-all', 'PUT');
}
