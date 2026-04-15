/**
 * Response envelope + ProTable adapters.
 *
 * Backend envelope (echox.Response) — note code is a string server-side even
 * though the design doc writes it as a number. We accept both.
 */
import type { AxiosError, AxiosResponse } from 'axios';

export type PageInfo = {
  total: number;
  pageNum: number;
  pageSize: number;
};

export type ApiResponse<T = unknown> = {
  code: number | string;
  data: T;
  message?: string;
  page?: PageInfo;
};

/** Any of these values is considered a successful biz code. Backend (echox) uses the "00000" convention. */
const SUCCESS_CODES = new Set<number | string>([
  0,
  '0',
  200,
  '200',
  '00000',
  'SUCCESS',
]);

export function isSuccess(code: number | string | undefined): boolean {
  if (code === undefined || code === null) return false;
  return SUCCESS_CODES.has(code);
}

export class ApiError extends Error {
  readonly code: number | string;
  readonly httpStatus?: number;
  readonly requestId?: string;
  readonly response?: ApiResponse<unknown>;
  readonly kind: 'biz' | 'http' | 'network';

  constructor(init: {
    message: string;
    code: number | string;
    httpStatus?: number;
    requestId?: string;
    response?: ApiResponse<unknown>;
    kind: 'biz' | 'http' | 'network';
  }) {
    super(init.message);
    this.name = 'ApiError';
    this.code = init.code;
    this.httpStatus = init.httpStatus;
    this.requestId = init.requestId;
    this.response = init.response;
    this.kind = init.kind;
  }
}

export function fromAxiosError(error: AxiosError): ApiError {
  const requestId =
    (error.response?.headers?.['x-request-id'] as string) ?? undefined;
  if (!error.response) {
    return new ApiError({
      message: '网络异常，请稍后重试',
      code: 'NETWORK',
      requestId,
      kind: 'network',
    });
  }
  const body = (error.response.data ?? {}) as Partial<ApiResponse<unknown>>;
  const message =
    body.message ?? error.message ?? `HTTP ${error.response.status}`;
  return new ApiError({
    message,
    code: body.code ?? error.response.status,
    httpStatus: error.response.status,
    requestId,
    response: body as ApiResponse<unknown>,
    kind: 'http',
  });
}

/**
 * ProTable request adapter. Given a function that queries a list endpoint in
 * Light-Admin shape, return a function shaped for ProTable.
 *
 * Light Admin pagination: query params `pageNum`, `pageSize`; response
 * `{ data, page: { total, pageNum, pageSize } }`.
 */
export type ProTableParams = {
  current?: number;
  pageSize?: number;
  [key: string]: unknown;
};

export type ListQuery = {
  pageNum?: number;
  pageSize?: number;
  [key: string]: unknown;
};

export type ListResult<T> = {
  data: T[];
  page?: PageInfo;
};

export function toProTableRequest<T>(
  service: (params: ListQuery) => Promise<ListResult<T>>,
) {
  return async (params: ProTableParams) => {
    const { current, pageSize, ...rest } = params;
    const resp = await service({
      pageNum: current,
      pageSize,
      ...rest,
    });
    return {
      data: resp.data,
      total: resp.page?.total ?? resp.data.length,
      success: true,
    };
  };
}

/** Raw-axios type re-export so service modules don't have to import plugin-request themselves. */
export type RawResponse<T = unknown> = AxiosResponse<ApiResponse<T>>;
