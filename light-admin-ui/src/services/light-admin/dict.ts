/**
 * Dict service — dict + dict-item CRUD plus the hot-path option fetch used by
 * <DictSelect>. The cache layer lives in `utils/dict/cache.ts`.
 */
import { requestData, requestList } from './request';
import type {
  Dict,
  DictForm,
  DictItem,
  DictItemForm,
  DictItemOption,
} from '@/types/light-admin/domain';

export function queryDicts(params: Record<string, unknown> = {}) {
  return requestList<Dict>('/dicts', 'GET', { params });
}

export function getDictForm(id: string) {
  return requestData<DictForm>(`/dicts/${id}/form`, 'GET');
}

export function createDict(data: DictForm) {
  return requestData<void>('/dicts', 'POST', { data });
}

export function updateDict(id: string, data: DictForm) {
  return requestData<void>(`/dicts/${id}`, 'PUT', { data });
}

export function deleteDicts(ids: string[]) {
  return requestData<void>(`/dicts/${ids.join(',')}`, 'DELETE');
}

export function queryDictItems(dictCode: string, params: Record<string, unknown> = {}) {
  return requestList<DictItem>(`/dicts/${dictCode}/items`, 'GET', { params });
}

export function getDictItemOptions(dictCode: string) {
  return requestData<DictItemOption[]>(`/dicts/${dictCode}/items/options`, 'GET');
}

export function getDictItemForm(dictCode: string, itemId: string) {
  return requestData<DictItemForm>(`/dicts/${dictCode}/items/${itemId}/form`, 'GET');
}

export function createDictItem(dictCode: string, data: DictItemForm) {
  return requestData<void>(`/dicts/${dictCode}/items`, 'POST', { data });
}

export function updateDictItem(dictCode: string, itemId: string, data: DictItemForm) {
  return requestData<void>(`/dicts/${dictCode}/items/${itemId}`, 'PUT', { data });
}

export function deleteDictItems(dictCode: string, itemIds: string[]) {
  return requestData<void>(`/dicts/${dictCode}/items/${itemIds.join(',')}`, 'DELETE');
}
