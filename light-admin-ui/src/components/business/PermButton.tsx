/**
 * <PermButton code="sys:user:add"> — antd Button wrapped in <Auth>. Silent
 * when the user lacks the permission.
 */
import { Button, type ButtonProps } from 'antd';
import React from 'react';
import Auth from './Auth';

export type PermButtonProps = ButtonProps & {
  code?: string | string[];
  role?: string | string[];
  mode?: 'any' | 'all';
};

export const PermButton: React.FC<PermButtonProps> = ({
  code,
  role,
  mode,
  children,
  ...rest
}) => (
  <Auth code={code} role={role} mode={mode}>
    <Button {...rest}>{children}</Button>
  </Auth>
);

export default PermButton;
