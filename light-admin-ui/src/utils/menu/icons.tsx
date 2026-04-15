/**
 * Icon name → React element, resolved dynamically against `@ant-design/icons`.
 *
 * Store the exact Ant Design icon component name in the backend menu (e.g.
 * `UserOutlined`, `SettingOutlined`, `CloudDownloadOutlined`). Case-insensitive
 * lookup + the `Outlined` suffix are tolerated so `user`, `User`, `user-outlined`
 * all resolve to `UserOutlined`.
 *
 * Trade-off: `import * as Icons` can't tree-shake (dynamic key access), so the
 * full `@ant-design/icons` set ships. On an admin the bundle impact is
 * acceptable; in exchange we avoid a hand-maintained mapping table.
 */
import * as Icons from '@ant-design/icons';
import React from 'react';

type IconComponent = React.ComponentType<{ className?: string }>;

const table = Icons as unknown as Record<string, IconComponent>;

function toPascal(s: string): string {
  return s
    .replace(/^el-icon-/i, '')
    .replace(/[-_\s]+(.)/g, (_, c: string) => c.toUpperCase())
    .replace(/^./, (c) => c.toUpperCase());
}

export function resolveIcon(raw?: string): React.ReactNode {
  if (!raw) return undefined;
  const name = raw.trim();
  if (!name) return undefined;

  const candidates = [
    name,
    toPascal(name),
    `${toPascal(name)}Outlined`,
  ];
  for (const key of candidates) {
    const Cmp = table[key];
    if (Cmp && typeof Cmp === 'object') {
      return React.createElement(Cmp);
    }
  }
  return undefined;
}
