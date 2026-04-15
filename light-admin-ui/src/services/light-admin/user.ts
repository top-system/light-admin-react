/**
 * Users service — identity + full CRUD. Identity helpers are used by the
 * bootstrap flow (Task 2.3); the CRUD ones drive the system/user page.
 */
import { requestData, requestList } from './request';
import type {
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

export function getUserForm(id: number) {
  return requestData<UserForm>(`/users/${id}/form`, 'GET');
}

export function updateUser(id: number, data: UserForm) {
  return requestData<void>(`/users/${id}`, 'PUT', { data });
}

export function deleteUser(id: number) {
  return requestData<void>(`/users/${id}`, 'DELETE');
}

export function resetUserPassword(id: number, password?: string) {
  return requestData<void>(`/users/${id}/password/reset`, 'PUT', {
    data: password ? { password } : undefined,
  });
}
