/**
 * Menu service. Task 2.4/2.5 only needs the user-specific route fetch —
 * `listMenus`, `getMenuForm`, and CRUD land in Task 3.6.
 */
import { requestData } from './request';
import type {
  MenuForm,
  MenuNode,
  MenuOption,
  RouteItem,
} from '@/types/light-admin/domain';

export function getRoutes() {
  return requestData<RouteItem[]>('/menus/routes', 'GET');
}

export function listMenus(params: Record<string, unknown> = {}) {
  return requestData<MenuNode[]>('/menus', 'GET', { params });
}

export function getMenuOptions() {
  return requestData<MenuOption[]>('/menus/options', 'GET');
}

export function getMenuForm(id: number) {
  return requestData<MenuForm>(`/menus/${id}/form`, 'GET');
}

export function createMenu(data: MenuForm) {
  return requestData<void>('/menus', 'POST', { data });
}

export function updateMenu(id: number, data: MenuForm) {
  return requestData<void>(`/menus/${id}`, 'PUT', { data });
}

export function deleteMenu(id: number) {
  return requestData<void>(`/menus/${id}`, 'DELETE');
}
