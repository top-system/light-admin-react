/**
 * Access helpers. Three-layer permission: menu (driven by `/menus/routes`),
 * route (layout `childrenRender` wraps in `AccessGate`), and
 * button/action (`<Auth code="...">` + `useAccess()`).
 *
 * `admin` and `root` short-circuit all checks to mirror backend Casbin.
 */
import type { InitialState } from './app';

const ADMIN_ROLES = ['admin', 'root'];

function isAdminRoles(roles: string[] | undefined): boolean {
  if (!roles) return false;
  return roles.some((r) => ADMIN_ROLES.includes(r.toLowerCase()));
}

export default function access(initialState: InitialState | undefined) {
  const roles = initialState?.roles ?? [];
  const permissions = initialState?.permissions ?? [];
  const admin = isAdminRoles(roles);
  const hasPerm = (perm: string) =>
    admin || permissions.includes('*:*:*') || permissions.includes(perm);

  return {
    canAdmin: admin,
    hasRole: (role: string) => admin || roles.includes(role),
    hasPerm,
    hasAnyPerm: (perms: string[]) => admin || perms.some(hasPerm),
    hasAllPerm: (perms: string[]) => admin || perms.every(hasPerm),
  };
}
