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

function resolvePath(
  routePath: string | undefined,
  parentPath: string,
): string {
  return routePath?.startsWith('/')
    ? routePath
    : `${parentPath.replace(/\/$/, '')}/${routePath ?? ''}`;
}

function walk(
  nodes: RouteItem[],
  parentPath: string,
  permissions: Set<string>,
  routeButtons: RoutePermMap,
): MenuDataItem[] {
  const out: MenuDataItem[] = [];
  for (const node of nodes) {
    if (node.type === 'button') {
      const codes = node.perms ?? [];
      for (const code of codes) permissions.add(code);
      if (codes.length) {
        const existing = routeButtons[parentPath] ?? [];
        routeButtons[parentPath] = [...existing, ...codes];
      }
      continue;
    }
    if (!node.path && !node.redirect) continue;

    for (const code of node.perms ?? []) permissions.add(code);

    const fullPath = resolvePath(node.path, parentPath);
    // The Go API wraps standalone pages in a Layout with an index child.
    // React already supplies that layout at the parent page path.
    const standalone =
      node.component === 'Layout' &&
      node.children?.length === 1 &&
      node.children[0].path === 'index';
    const item: MenuDataItem & { routes?: MenuDataItem[] } = {
      path: fullPath,
      name: node.meta?.title ?? node.name,
      icon: resolveIcon(node.meta?.icon ?? node.icon),
      hideInMenu: node.meta?.hidden ?? node.hideInMenu,
      redirect: standalone ? undefined : node.redirect,
    };
    if (node.children?.length && !standalone) {
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

export function transformMenu(
  routes: RouteItem[] | null | undefined,
): TransformResult {
  const permissions = new Set<string>();
  const routeButtons: RoutePermMap = {};
  const menu = walk(routes ?? [], '/', permissions, routeButtons);
  return {
    menu,
    permissions: Array.from(permissions),
    routeButtons,
  };
}

/** Collect every non-button path present in the backend-returned tree. */
export function collectAllowedPaths(routes: RouteItem[]): Set<string> {
  const out = new Set<string>();
  const walkPaths = (nodes: RouteItem[], parentPath: string) => {
    for (const n of nodes) {
      if (n.type === 'button') continue;
      const fullPath = resolvePath(n.path, parentPath);
      if (n.path) out.add(fullPath);
      if (n.children?.length) walkPaths(n.children, fullPath);
    }
  };
  walkPaths(routes, '/');
  return out;
}
