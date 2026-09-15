import access from './access';

it('recognizes the backend ROOT role as administrator', () => {
  const permissions = access({
    roles: ['ROOT'],
    permissions: ['*:*:*'],
    menuRoutes: [],
  });
  expect(permissions.canAdmin).toBe(true);
  expect(permissions.hasPerm('face-swap:task:add')).toBe(true);
});

it('honors explicit full access while preserving ordinary action restrictions', () => {
  const full = access({ roles: [], permissions: ['*:*:*'], menuRoutes: [] });
  expect(full.hasAllPerm(['face-swap:task:add', 'face-swap:task:stop'])).toBe(
    true,
  );
  const viewer = access({
    roles: ['viewer'],
    permissions: ['face-swap:task:query'],
    menuRoutes: [],
  });
  expect(viewer.canAdmin).toBe(false);
  expect(viewer.hasPerm('face-swap:task:query')).toBe(true);
  expect(viewer.hasPerm('face-swap:task:add')).toBe(false);
  expect(viewer.hasAnyPerm(['face-swap:task:add', 'face-swap:task:stop'])).toBe(
    false,
  );
});
