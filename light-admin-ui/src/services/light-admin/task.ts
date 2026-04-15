import { requestData, requestList } from './request';
import type { TaskRow, TaskStats, TaskType } from '@/types/light-admin/domain';

export function getTaskStats() {
  return requestData<TaskStats>('/tasks/stats', 'GET');
}

export function getTaskTypes() {
  return requestData<TaskType[]>('/tasks/types', 'GET');
}

export function queryTasks(params: Record<string, unknown> = {}) {
  return requestList<TaskRow>('/tasks', 'GET', { params });
}

export function getTask(id: number) {
  return requestData<TaskRow>(`/tasks/${id}`, 'GET');
}

export function deleteTask(id: number) {
  return requestData<void>(`/tasks/${id}`, 'DELETE');
}
