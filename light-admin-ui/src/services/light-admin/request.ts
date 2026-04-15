/**
 * Configured @umijs/max `request` for the Light Admin backend.
 *
 * Shape contract: every service function returns the `data` payload directly;
 * business errors surface as `ApiError` thrown from response interception.
 * Paginated services return `{ data, page }` via `requestList`.
 */
import type { RequestConfig } from '@umijs/max';
import type { AxiosError, AxiosRequestConfig } from 'axios';
import { message } from 'antd';
import {
  ApiError,
  type ApiResponse,
  fromAxiosError,
  isSuccess,
  type ListResult,
  type PageInfo,
} from '@/utils/response/adapter';
import { clearToken, getToken } from '@/utils/auth/token';
import { buildLoginRedirect, LOGIN_PATH } from '@/utils/auth/redirect';

export const API_PREFIX = '/api/v1';

/** Build the `request` config Umi Max picks up from app.tsx. */
export const lightAdminRequestConfig: RequestConfig = {
  baseURL: API_PREFIX,
  timeout: 20_000,
  errorConfig: {
    errorHandler(error: unknown, opts: { skipErrorHandler?: boolean } = {}) {
      if (opts.skipErrorHandler) throw error;
      if (error instanceof ApiError) {
        if (error.kind === 'biz') {
          message.error(error.message);
        } else if (error.kind === 'http' && error.httpStatus && error.httpStatus >= 500) {
          const tag = error.requestId ? ` (reqId: ${error.requestId})` : '';
          message.error(`服务器错误${tag}`);
        } else if (error.kind === 'http' && error.httpStatus === 404) {
          message.error('资源不存在');
        } else if (error.kind === 'network') {
          message.error(error.message);
        } else if (error.kind === 'http') {
          message.error(error.message);
        }
      }
      throw error;
    },
  },
  requestInterceptors: [
    (config: AxiosRequestConfig) => {
      const token = getToken();
      if (token) {
        config.headers = {
          ...(config.headers ?? {}),
          Authorization: `Bearer ${token}`,
        };
      }
      return config;
    },
  ],
  responseInterceptors: [
    [
      (response) => {
        const body = response.data as ApiResponse<unknown> | undefined;
        if (!body || typeof body !== 'object') return response;

        if (isSuccess(body.code)) {
          return response;
        }

        // HTTP 2xx with a business error — convert to ApiError so `errorHandler`
        // produces a user-facing toast and the awaiting caller rejects.
        throw new ApiError({
          message: body.message ?? '请求失败',
          code: body.code,
          httpStatus: response.status,
          requestId: (response.headers?.['x-request-id'] as string) ?? undefined,
          response: body,
          kind: 'biz',
        });
      },
      (error: Error) => {
        const apiError = fromAxiosError(error as AxiosError);
        if (apiError.httpStatus === 401) {
          clearToken();
          if (typeof window !== 'undefined' && window.location.pathname !== LOGIN_PATH) {
            window.location.href = buildLoginRedirect();
          }
        } else if (apiError.httpStatus === 403) {
          if (typeof window !== 'undefined' && window.location.pathname !== '/exception/403') {
            window.location.href = '/exception/403';
          }
        }
        return Promise.reject(apiError);
      },
    ],
  ],
};

// --- Service helpers -------------------------------------------------------

import { request } from '@umijs/max';

type RequestOptions = {
  params?: Record<string, unknown>;
  data?: unknown;
  headers?: Record<string, string>;
  skipErrorHandler?: boolean;
  responseType?: 'json' | 'blob' | 'arraybuffer';
};

/** Envelope-unwrapping request that returns `data` directly. */
export async function requestData<T>(url: string, method: string, opts: RequestOptions = {}): Promise<T> {
  const response = await request<ApiResponse<T>>(url, {
    method,
    params: opts.params,
    data: opts.data,
    headers: opts.headers,
    skipErrorHandler: opts.skipErrorHandler,
    responseType: opts.responseType,
    getResponse: true,
  });
  return response.data.data;
}

/** Same as requestData but also preserves pagination metadata. */
export async function requestList<T>(
  url: string,
  method: string,
  opts: RequestOptions = {},
): Promise<ListResult<T>> {
  const response = await request<ApiResponse<T[]>>(url, {
    method,
    params: opts.params,
    data: opts.data,
    headers: opts.headers,
    skipErrorHandler: opts.skipErrorHandler,
    responseType: opts.responseType,
    getResponse: true,
  });
  const body = response.data;
  return {
    data: body.data ?? [],
    page: body.page as PageInfo | undefined,
  };
}

export { ApiError };
