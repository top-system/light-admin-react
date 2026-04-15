/**
 * Route-level permission guard. Wrapped around page content inside the layout
 * `childrenRender` pipeline. If the current user lacks any of the required
 * perms, the user is redirected to `/exception/403`.
 */
import { history, useAccess } from '@umijs/max';
import React, { useEffect } from 'react';

export type AccessGateProps = {
  perms?: string[];
  children: React.ReactNode;
};

export const AccessGate: React.FC<AccessGateProps> = ({ perms, children }) => {
  const access = useAccess() as {
    canAdmin?: boolean;
    hasAnyPerm?: (p: string[]) => boolean;
  };

  const allowed =
    !perms ||
    perms.length === 0 ||
    access.canAdmin ||
    access.hasAnyPerm?.(perms);

  useEffect(() => {
    if (!allowed) {
      history.replace('/exception/403');
    }
  }, [allowed]);

  if (!allowed) return null;
  return <>{children}</>;
};

export default AccessGate;
