/**
 * Users service — identity + full CRUD. Identity helpers are used by the
 * bootstrap flow (Task 2.3); the CRUD ones drive the system/user page.
 */
import { requestData, requestList } from './request';
import type {
  ChangePasswordRequest,
  CurrentUser,
  User,
  UserForm,
  UserOption,
  UserProfileUpdate,
  UserQuery,
} from '@/types/light-admin/domain';

// Identity ------------------------------------------------------------------

export function getMe() {
  return requestData<CurrentUser>('/users/me', 'GET');
}

export function getProfile() {
  return requestData<CurrentUser>('/users/profile', 'GET');
}

export function updateProfile(payload: UserProfileUpdate) {
  return requestData<void>('/users/profile', 'PUT', { data: payload });
}

/** 当前用户自助修改密码：需校验旧密码。 */
export function changePassword(payload: ChangePasswordRequest) {
  return requestData<void>('/users/password', 'PUT', { data: payload });
}

export function getUserOptions() {
  return requestData<UserOption[]>('/users/options', 'GET');
}

// CRUD ----------------------------------------------------------------------

export function queryUsers(params: UserQuery) {
  return requestList<User>('/users', 'GET', { params });
}

export function createUser(data: UserForm) {
  return requestData<void>('/users', 'POST', { data });
}

export function getUserForm(id: string) {
  return requestData<UserForm>(`/users/${id}/form`, 'GET');
}

export function updateUser(id: string, data: UserForm) {
  return requestData<void>(`/users/${id}`, 'PUT', { data });
}

export function deleteUser(id: string) {
  return requestData<void>(`/users/${id}`, 'DELETE');
}

/**
 * 管理员重置指定用户密码。后端优先读 query `password`，同时兼容 JSON body，
 * 两处都带上以适配新旧版本服务端。
 */
export function resetUserPassword(id: string, password: string) {
  return requestData<void>(`/users/${id}/password/reset`, 'PUT', {
    params: { password },
    data: { password },
  });
}
