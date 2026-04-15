import {
  LogoutOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { history, useModel } from '@umijs/max';
import type { MenuProps } from 'antd';
import { Spin } from 'antd';
import { createStyles } from 'antd-style';
import React from 'react';
import { flushSync } from 'react-dom';
import HeaderDropdown from '../HeaderDropdown';

export type GlobalHeaderRightProps = {
  menu?: boolean;
  children?: React.ReactNode;
};

export const AvatarName = () => {
  const { initialState } = useModel('@@initialState');
  const { currentUser } = initialState || {};
  return <span className="anticon">{currentUser?.nickname}</span>;
};

const useStyles = createStyles(({ token }) => {
  return {
    action: {
      display: 'flex',
      height: '48px',
      marginLeft: 'auto',
      overflow: 'hidden',
      alignItems: 'center',
      padding: '0 8px',
      cursor: 'pointer',
      borderRadius: token.borderRadius,
      '&:hover': {
        backgroundColor: token.colorBgTextHover,
      },
    },
  };
});

export const AvatarDropdown: React.FC<GlobalHeaderRightProps> = ({
  menu,
  children,
}) => {
  const { styles } = useStyles();
  const { initialState, setInitialState } = useModel('@@initialState');

  // Real logout wiring lands in Task 2.3 (auth service). For now this is a
  // local session reset that matches the eventual behavior: clear token,
  // clear initialState, and bounce back to the login page.
  const loginOut = async () => {
    try {
      window.localStorage.removeItem('light-admin-token');
    } catch {
      /* ignore storage errors */
    }
    const { search, pathname } = window.location;
    const searchParams = new URLSearchParams({ redirect: pathname + search });
    if (window.location.pathname !== '/user/login') {
      history.replace({
        pathname: '/user/login',
        search: searchParams.toString(),
      });
    }
  };

  const onMenuClick: MenuProps['onClick'] = (event) => {
    const { key } = event;
    if (key === 'logout') {
      flushSync(() => {
        setInitialState((s) => ({
          ...s,
          currentUser: undefined,
          permissions: [],
          roles: [],
          menuRoutes: [],
        }));
      });
      void loginOut();
      return;
    }
    if (key === 'center' || key === 'settings') {
      history.push('/profile');
    }
  };

  const loading = (
    <span className={styles.action}>
      <Spin size="small" style={{ marginLeft: 8, marginRight: 8 }} />
    </span>
  );

  if (!initialState) {
    return loading;
  }

  const { currentUser } = initialState;
  if (!currentUser?.nickname) {
    return loading;
  }

  const menuItems = [
    ...(menu
      ? [
          { key: 'center', icon: <UserOutlined />, label: '个人中心' },
          { key: 'settings', icon: <SettingOutlined />, label: '个人设置' },
          { type: 'divider' as const },
        ]
      : []),
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录' },
  ];

  return (
    <HeaderDropdown
      menu={{
        selectedKeys: [],
        onClick: onMenuClick,
        items: menuItems,
      }}
    >
      {children}
    </HeaderDropdown>
  );
};
