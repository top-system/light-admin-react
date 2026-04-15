/**
 * Light Admin routes.
 *
 * The static tree is deliberately narrow. Business routes are added in later
 * chunks; `patchClientRoutes` (src/app.tsx) filters them at runtime based on
 * the user-specific menu returned by `/api/v1/menus/routes`.
 */
export default [
  {
    path: '/user',
    layout: false,
    routes: [
      {
        name: 'login',
        path: '/user/login',
        component: './user/login',
      },
      { path: '/user', redirect: '/user/login' },
    ],
  },
  {
    name: 'exception',
    icon: 'warning',
    path: '/exception',
    layout: false,
    routes: [
      { path: '/exception', redirect: '/exception/404' },
      { name: '403', path: '/exception/403', component: './exception/403' },
      { name: '404', path: '/exception/404', component: './exception/404' },
      { name: '500', path: '/exception/500', component: './exception/500' },
    ],
  },
  { name: 'home', path: '/home', icon: 'home', component: './home' },
  {
    path: '/component',
    name: 'component',
    icon: 'appstore',
    routes: [
      { path: '/component', redirect: '/component/curd' },
      { name: 'curd', path: '/component/curd', component: './component/curd' },
      { name: 'upload', path: '/component/upload', component: './component/upload' },
      { name: 'text-scroll', path: '/component/text-scroll', component: './component/text-scroll' },
      { name: 'drag', path: '/component/drag', component: './component/drag' },
      { name: 'icon-select', path: '/component/icon-select', component: './component/icon-select' },
      { name: 'operation-column', path: '/component/operation-column', component: './component/operation-column' },
      { name: 'table-select', path: '/component/table-select', component: './component/table-select' },
      { name: 'tinymce', path: '/component/tinymce', component: './component/tinymce' },
    ],
  },
  {
    path: '/function',
    name: 'function',
    icon: 'thunderbolt',
    routes: [
      { path: '/function', redirect: '/function/websocket' },
      { name: 'websocket', path: '/function/websocket', component: './function/websocket' },
      { name: 'icon-demo', path: '/function/icon-demo', component: './function/icon-demo' },
      { name: 'dict-sync', path: '/function/dict-sync', component: './function/dict-sync' },
      { name: 'curd-single', path: '/function/curd-single', component: './function/curd-single' },
    ],
  },
  {
    path: '/multi-level',
    name: 'multi-level',
    icon: 'apartment',
    routes: [
      { path: '/multi-level', redirect: '/multi-level/multi-level1/multi-level2/multi-level3-1' },
      {
        path: '/multi-level/multi-level1',
        name: 'multi-level1',
        routes: [
          {
            path: '/multi-level/multi-level1/multi-level2',
            name: 'multi-level2',
            routes: [
              {
                name: 'multi-level3-1',
                path: '/multi-level/multi-level1/multi-level2/multi-level3-1',
                component: './multi-level/level3-1',
              },
              {
                name: 'multi-level3-2',
                path: '/multi-level/multi-level1/multi-level2/multi-level3-2',
                component: './multi-level/level3-2',
              },
            ],
          },
        ],
      },
    ],
  },
  {
    path: '/route-param',
    name: 'route-param',
    icon: 'star',
    routes: [
      { path: '/route-param', redirect: '/route-param/route-param-type1?type=1' },
      {
        name: 'route-param-type1',
        path: '/route-param/route-param-type1',
        component: './route-param',
      },
      {
        name: 'route-param-type2',
        path: '/route-param/route-param-type2',
        component: './route-param',
      },
    ],
  },
  {
    path: '/system',
    name: 'system',
    icon: 'setting',
    routes: [
      { path: '/system', redirect: '/system/user' },
      { name: 'user', path: '/system/user', component: './system/user' },
      { name: 'role', path: '/system/role', component: './system/role' },
      { name: 'menu', path: '/system/menu', component: './system/menu' },
      { name: 'dept', path: '/system/dept', component: './system/dept' },
      { name: 'dict', path: '/system/dict', component: './system/dict' },
      { name: 'config', path: '/system/config', component: './system/config' },
      { name: 'log', path: '/system/log', component: './system/log' },
      { name: 'notice', path: '/system/notice', component: './system/notice' },
      { name: 'queue', path: '/system/queue', component: './system/queue' },
      {
        name: 'downloader',
        path: '/system/downloader',
        component: './system/downloader',
      },
    ],
  },
  {
    name: 'profile',
    path: '/profile',
    icon: 'user',
    component: './profile',
    hideInMenu: true,
  },
  { path: '/', redirect: '/home' },
  { component: './404', path: '/*' },
];
