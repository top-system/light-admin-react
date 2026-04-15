/**
 * Auth service — captcha + login + logout. Sits on top of `requestData` so
 * callers always receive the unwrapped `data` payload or throw ApiError.
 */
import { requestData } from './request';
import type {
  Captcha,
  LoginRequest,
  LoginResponse,
} from '@/types/light-admin/domain';

export function getCaptcha() {
  return requestData<Captcha>('/auth/captcha', 'GET', { skipErrorHandler: true });
}

export function login(payload: LoginRequest) {
  return requestData<LoginResponse>('/auth/login', 'POST', {
    data: payload,
    skipErrorHandler: true,
  });
}

export function logout() {
  return requestData<void>('/auth/logout', 'DELETE', { skipErrorHandler: true });
}
