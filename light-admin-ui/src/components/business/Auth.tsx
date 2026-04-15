/**
 * <Auth code="sys:user:add"> — hides its children unless the current user has
 * the given perm. Accepts either a single code or an array (any-of). Intended
 * wrapper for toolbar buttons, row actions, etc.
 */
import { useAccess } from '@umijs/max';
import React from 'react';

export type AuthProps = {
  code?: string | string[];
  role?: string | string[];
  mode?: 'any' | 'all';
  fallback?: React.ReactNode;
  children: React.ReactNode;
};

export const Auth: React.FC<AuthProps> = ({
  code,
  role,
  mode = 'any',
  fallback = null,
  children,
}) => {
  const access = useAccess() as {
    canAdmin?: boolean;
    hasPerm?: (p: string) => boolean;
    hasAnyPerm?: (p: string[]) => boolean;
    hasAllPerm?: (p: string[]) => boolean;
    hasRole?: (r: string) => boolean;
  };

  if (access.canAdmin) return <>{children}</>;

  if (code) {
    const codes = Array.isArray(code) ? code : [code];
    const pass =
      mode === 'all' ? access.hasAllPerm?.(codes) : access.hasAnyPerm?.(codes);
    if (!pass) return <>{fallback}</>;
  }

  if (role) {
    const roles = Array.isArray(role) ? role : [role];
    const pass =
      mode === 'all'
        ? roles.every((r) => access.hasRole?.(r))
        : roles.some((r) => access.hasRole?.(r));
    if (!pass) return <>{fallback}</>;
  }

  return <>{children}</>;
};

export default Auth;
