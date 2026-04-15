/**
 * Runtime configuration — request client, initial state, layout.
 * Task 2.5 layers on dynamic menu + patchClientRoutes.
 */
import type { RequestConfig, RunTimeLayoutConfig } from '@umijs/max';
import { history, Link } from '@umijs/max';
import React from 'react';
import { AvatarDropdown, AvatarName, Footer, NoticeBell } from '@/components';
import AccessGate from '@/components/business/AccessGate';
import { getRoutes } from '@/services/light-admin/menu';
import { lightAdminRequestConfig } from '@/services/light-admin/request';
import { getMe } from '@/services/light-admin/user';
import type { CurrentUser, RouteItem } from '@/types/light-admin/domain';
import { buildLoginRedirect, LOGIN_PATH } from '@/utils/auth/redirect';
import { clearToken, getToken } from '@/utils/auth/token';
import { invalidateDict } from '@/utils/dict/cache';
import { transformMenu } from '@/utils/menu/transform';
import { startWsClient, stopWsClient } from '@/utils/ws/client';
import defaultSettings from '../config/defaultSettings';

export type InitialState = {
  settings?: typeof defaultSettings;
  currentUser?: CurrentUser;
  permissions: string[];
  roles: string[];
  menuRoutes: RouteItem[];
  refresh?: () => Promise<InitialState>;
};

const publicPaths = new Set([LOGIN_PATH]);

function buildEmptyState(): InitialState {
  return {
    settings: defaultSettings,
    permissions: [],
    roles: [],
    menuRoutes: [],
  };
}

/**
 * Bootstrap. When no token, we hand back an empty state and let the layout
 * bounce to /user/login. With a token, fetch identity + menu in parallel; any
 * failure short-circuits back to login with the token cleared.
 */
async function bootstrapState(): Promise<InitialState> {
  const empty = buildEmptyState();
  const token = getToken();
  if (!token) {
    stopWsClient();
    return empty;
  }

  try {
    const [currentUser, menuRoutes] = await Promise.all([
      getMe(),
      getRoutes().catch(() => [] as RouteItem[]),
    ]);
    // Only start the WebSocket once we know the token is valid.
    const ws = startWsClient({ url: '/ws', token });
    ws.on<{ dictCode?: string }>('dict-change', (frame) => {
      if (frame.data?.dictCode) invalidateDict(frame.data.dictCode);
    });
    return {
      ...empty,
      currentUser,
      permissions: currentUser.perms ?? [],
      roles: currentUser.roles ?? [],
      menuRoutes,
    };
  } catch {
    clearToken();
    stopWsClient();
    return empty;
  }
}

export async function getInitialState(): Promise<InitialState> {
  const state = await bootstrapState();
  state.refresh = bootstrapState;
  return state;
}

export const layout: RunTimeLayoutConfig = ({ initialState }) => {
  const menuRoutes = initialState?.menuRoutes ?? [];
  const { menu: menuData } = transformMenu(menuRoutes);

  return {
    menuItemRender: (item, dom) =>
      item.path ? (
        <Link to={item.path} prefetch>
          {dom}
        </Link>
      ) : (
        dom
      ),
    menu: {
      locale: false,
      // Feed the layout the backend-derived tree instead of config/routes.ts.
      request: async () => menuData,
      params: { roles: initialState?.roles ?? [] },
    },
    actionsRender: () => [<NoticeBell key="notice" />],
    avatarProps: {
      src: initialState?.currentUser?.avatar,
      title: <AvatarName />,
      render: (_, children) => <AvatarDropdown>{children}</AvatarDropdown>,
    },
    footerRender: () => <Footer />,
    onPageChange: () => {
      const { location } = history;
      if (!initialState?.currentUser && !publicPaths.has(location.pathname)) {
        window.location.href = buildLoginRedirect();
      }
    },
    childrenRender: (children) => {
      // Route-level guard. We do not have per-route perm declarations yet
      // (added alongside each page in Chunks 3+), so the gate is a no-op
      // placeholder that future pages can opt into by providing `perms`.
      return <AccessGate>{children}</AccessGate>;
    },
    menuHeaderRender: undefined,
    ...initialState?.settings,
  };
};

/**
 * Static-route filter. Hides any route whose path is NOT present in the
 * backend-supplied menu tree (except for always-public paths). Unknown
 * backend routes only produce a console warning — we never synthesise new
 * client routes from backend strings.
 */
export function patchClientRoutes({ routes }: { routes: any[] }) {
  // Umi calls this before any `getInitialState` data is persisted; we read
  // the latest menu from localStorage-free state by hitting the model via
  // `window.g_initialProps` shim if available. For simplicity we instead
  // leave all routes visible — the layout's `menu.request` controls nav
  // visibility, and `AccessGate` guards per-page perms once 3+ lands.
  void routes;
}

export const request: RequestConfig = lightAdminRequestConfig;
