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
  return roles.some((r) => ADMIN_ROLES.includes(r));
}

export default function access(initialState: InitialState | undefined) {
  const roles = initialState?.roles ?? [];
  const permissions = initialState?.permissions ?? [];
  const admin = isAdminRoles(roles);

  return {
    canAdmin: admin,
    hasRole: (role: string) => admin || roles.includes(role),
    hasPerm: (perm: string) => admin || permissions.includes(perm),
    hasAnyPerm: (perms: string[]) =>
      admin || perms.some((p) => permissions.includes(p)),
    hasAllPerm: (perms: string[]) =>
      admin || perms.every((p) => permissions.includes(p)),
  };
}
