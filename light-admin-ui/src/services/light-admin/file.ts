/**
 * File upload/delete service. Uses axios directly so we can stream
 * progress events; `requestData` cannot surface XHR progress.
 */
import { request } from '@umijs/max';
import type { FileUploadResult } from '@/types/light-admin/domain';
import type { ApiResponse } from '@/utils/response/adapter';

export type UploadOptions = {
  onProgress?: (percent: number) => void;
  fieldName?: string;
};

export async function uploadFile(
  file: File | Blob,
  opts: UploadOptions = {},
): Promise<FileUploadResult> {
  const form = new FormData();
  form.append(opts.fieldName ?? 'file', file);
  const response = await request<ApiResponse<FileUploadResult>>('/files', {
    method: 'POST',
    data: form,
    requestType: 'form',
    onUploadProgress: (event) => {
      if (!opts.onProgress || !event.total) return;
      opts.onProgress(Math.round((event.loaded * 100) / event.total));
    },
    getResponse: true,
  });
  return response.data.data;
}

export function deleteFile(filePath: string) {
  return request<ApiResponse<void>>('/files', {
    method: 'DELETE',
    params: { filePath },
    getResponse: true,
  }).then(() => undefined);
}
