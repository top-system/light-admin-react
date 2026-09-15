import { collectAllowedPaths, transformMenu } from './transform';

describe('backend route menus', () => {
  it('resolves relative children and retains pathless button permissions', () => {
    const routes = [
      {
        path: '/face-swap',
        children: [
          {
            path: 'tasks',
            children: [
              {
                path: '',
                type: 'button' as const,
                perms: ['face-swap:task:add'],
              },
            ],
          },
        ],
      },
    ];
    expect(collectAllowedPaths(routes).has('/face-swap/tasks')).toBe(true);
    expect(transformMenu(routes).routeButtons['/face-swap/tasks']).toEqual([
      'face-swap:task:add',
    ]);
  });
  it('uses route metadata for titles, icons and visibility', () => {
    const { menu } = transformMenu([
      {
        path: '/face-swap',
        name: '',
        component: 'Layout',
        meta: { title: '换脸管理', icon: 'VideoCameraOutlined' },
        children: [
          {
            path: '/face-swap/tasks',
            meta: { title: '换脸任务', hidden: true },
          },
        ],
      },
    ] as any);
    expect(menu[0].name).toBe('换脸管理');
    expect(menu[0].icon).toBeTruthy();
    expect(menu[0].routes?.[0]).toMatchObject({
      name: '换脸任务',
      hideInMenu: true,
    });
  });

  it('maps a synthetic Layout/index wrapper to the existing page route', () => {
    const { menu } = transformMenu([
      {
        path: '/home',
        component: 'Layout',
        redirect: '/home/index',
        meta: { title: '首页' },
        children: [
          { path: 'index', component: 'home/index', meta: { title: '首页' } },
        ],
      },
    ] as any);
    expect(menu).toHaveLength(1);
    expect(menu[0]).toMatchObject({ path: '/home', name: '首页' });
    expect(menu[0].routes).toBeUndefined();
    expect(menu[0].redirect).toBeUndefined();
  });

  it('accepts an empty backend response', () => {
    expect(transformMenu(null as any).menu).toEqual([]);
  });
});
