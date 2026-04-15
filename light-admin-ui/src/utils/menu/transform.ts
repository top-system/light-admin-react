/**
 * Backend menu tree → Umi / Pro-Layout MenuDataItem tree + flat perm map.
 *
 * Button nodes (`type: 'button'`) never become menu entries — their `perms`
 * are collected into a map keyed by the nearest ancestor route path.
 */
import type { MenuDataItem } from '@ant-design/pro-components';
import type { RouteItem } from '@/types/light-admin/domain';
import { resolveIcon } from './icons';

export type RoutePermMap = Record<string, string[]>;

export type TransformResult = {
  menu: MenuDataItem[];
  /** Flat set of every permission code encountered (button-level + route-level). */
  permissions: string[];
  /** Route path → list of button perm codes declared under that route. */
  routeButtons: RoutePermMap;
};

function walk(
  nodes: RouteItem[],
  parentPath: string,
  permissions: Set<string>,
  routeButtons: RoutePermMap,
): MenuDataItem[] {
  const out: MenuDataItem[] = [];
  for (const node of nodes) {
    if (!node.path && !node.redirect) continue;

    if (node.type === 'button') {
      const codes = node.perms ?? [];
      for (const code of codes) permissions.add(code);
      if (codes.length) {
        const existing = routeButtons[parentPath] ?? [];
        routeButtons[parentPath] = [...existing, ...codes];
      }
      continue;
    }

    for (const code of node.perms ?? []) permissions.add(code);

    const fullPath = node.path ?? parentPath;
    const item: MenuDataItem & { routes?: MenuDataItem[] } = {
      path: fullPath,
      name: node.name,
      icon: resolveIcon(node.icon),
      hideInMenu: node.hideInMenu,
      redirect: node.redirect,
    };
    if (node.children?.length) {
      const kids = walk(node.children, fullPath, permissions, routeButtons);
      (item as { routes?: MenuDataItem[] }).routes = kids;
      if (kids.length === 0 && node.type === 'dir') {
        item.hideInMenu = true;
      }
    }
    out.push(item);
  }
  return out;
}

export function transformMenu(routes: RouteItem[]): TransformResult {
  const permissions = new Set<string>();
  const routeButtons: RoutePermMap = {};
  const menu = walk(routes, '/', permissions, routeButtons);
  return {
    menu,
    permissions: Array.from(permissions),
    routeButtons,
  };
}

/** Collect every non-button path present in the backend-returned tree. */
export function collectAllowedPaths(routes: RouteItem[]): Set<string> {
  const out = new Set<string>();
  const walkPaths = (nodes: RouteItem[]) => {
    for (const n of nodes) {
      if (n.type === 'button') continue;
      if (n.path) out.add(n.path);
      if (n.children?.length) walkPaths(n.children);
    }
  };
  walkPaths(routes);
  return out;
}
