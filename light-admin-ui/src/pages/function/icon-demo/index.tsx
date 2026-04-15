/**
 * /function/icon-demo — icon gallery. Same data as `/component/icon-select`
 * but without the copy-to-clipboard interaction — useful as a quick visual
 * reference for designers and menu configurators.
 */
import * as Icons from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Card, Tabs } from 'antd';
import React, { useMemo } from 'react';

function pickByPattern(re: RegExp) {
  return Object.entries(Icons)
    .filter(([k, v]) => re.test(k) && typeof v === 'object')
    .map(([k]) => k);
}

const IconDemo: React.FC = () => {
  const outlined = useMemo(() => pickByPattern(/Outlined$/), []);
  const filled = useMemo(() => pickByPattern(/Filled$/), []);
  const twoTone = useMemo(() => pickByPattern(/TwoTone$/), []);

  const renderGrid = (names: string[]) => (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(110px, 1fr))',
        gap: 8,
        maxHeight: 540,
        overflowY: 'auto',
      }}
    >
      {names.map((name) => {
        const Cmp = (Icons as unknown as Record<string, React.ComponentType<{ style?: React.CSSProperties }>>)[name];
        return (
          <div
            key={name}
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              gap: 4,
              padding: '10px 4px',
              border: '1px solid #f0f0f0',
              borderRadius: 4,
            }}
          >
            <Cmp style={{ fontSize: 24 }} />
            <span style={{ fontSize: 11, color: '#666' }}>{name}</span>
          </div>
        );
      })}
    </div>
  );

  return (
    <PageContainer
      title="图标展示"
      subTitle={`${outlined.length + filled.length + twoTone.length} 个 Ant Design 图标`}
    >
      <Card>
        <Tabs
          items={[
            { key: 'outlined', label: `Outlined (${outlined.length})`, children: renderGrid(outlined) },
            { key: 'filled', label: `Filled (${filled.length})`, children: renderGrid(filled) },
            { key: 'two-tone', label: `TwoTone (${twoTone.length})`, children: renderGrid(twoTone) },
          ]}
        />
      </Card>
    </PageContainer>
  );
};

export default IconDemo;
