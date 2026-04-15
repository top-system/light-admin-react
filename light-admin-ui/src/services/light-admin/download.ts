import { requestData, requestList } from './request';
import type {
  CreateDownloadRequest,
  DownloadDetail,
  DownloadRow,
  DownloadStats,
  DownloaderInfo,
} from '@/types/light-admin/domain';

export function getDownloadStats() {
  return requestData<DownloadStats>('/downloads/stats', 'GET');
}

export function getDownloaders() {
  return requestData<DownloaderInfo[]>('/downloads/downloaders', 'GET');
}

export function testDownloader(name: string) {
  return requestData<{ version: string }>(`/downloads/test/${name}`, 'GET');
}

export function queryDownloads(params: Record<string, unknown> = {}) {
  return requestList<DownloadRow>('/downloads', 'GET', { params });
}

export function getDownload(id: number) {
  return requestData<DownloadDetail>(`/downloads/${id}`, 'GET');
}

export function createDownload(data: CreateDownloadRequest) {
  return requestData<DownloadRow>('/downloads', 'POST', { data });
}

export function cancelDownload(id: number) {
  return requestData<void>(`/downloads/${id}/cancel`, 'POST');
}

export function setDownloadFiles(id: number, files: { index: number; download: boolean }[]) {
  return requestData<void>(`/downloads/${id}/files`, 'PUT', { data: { files } });
}

export function syncDownload(id: number) {
  return requestData<void>(`/downloads/${id}/sync`, 'POST');
}

export function deleteDownload(id: number) {
  return requestData<void>(`/downloads/${id}`, 'DELETE');
}
