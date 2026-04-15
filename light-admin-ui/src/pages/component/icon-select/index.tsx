/**
 * /component/icon-select — searchable icon picker over @ant-design/icons.
 * Reusable pattern: pick an icon, copy its name, use in menu.yaml.
 */
import * as Icons from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { App, Card, Input, Tooltip, Typography } from 'antd';
import React, { useMemo, useState } from 'react';

const ALL = Object.entries(Icons)
  .filter(([k, v]) => /Outlined$/.test(k) && typeof v === 'object')
  .map(([k]) => k);

const IconSelect: React.FC = () => {
  const { message } = App.useApp();
  const [keyword, setKeyword] = useState('');

  const list = useMemo(() => {
    if (!keyword.trim()) return ALL;
    const kw = keyword.toLowerCase();
    return ALL.filter((n) => n.toLowerCase().includes(kw));
  }, [keyword]);

  const copy = (name: string) => {
    navigator.clipboard?.writeText(name).catch(() => {});
    message.success(`已复制: ${name}`);
  };

  return (
    <PageContainer
      title="图标选择器"
      subTitle={`${ALL.length} 个 Outlined 风格图标,点击复制名称`}
    >
      <Card>
        <Input.Search
          placeholder="搜索图标(如 user / setting)"
          allowClear
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <Typography.Paragraph type="secondary">
          共 {list.length} 个匹配项
        </Typography.Paragraph>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(120px, 1fr))',
            gap: 8,
            maxHeight: 540,
            overflowY: 'auto',
          }}
        >
          {list.map((name) => {
            const Cmp = (Icons as unknown as Record<string, React.ComponentType<{ style?: React.CSSProperties }>>)[name];
            return (
              <Tooltip key={name} title={name}>
                <button
                  type="button"
                  onClick={() => copy(name)}
                  style={{
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    gap: 4,
                    padding: '10px 4px',
                    border: '1px solid #eee',
                    background: '#fff',
                    borderRadius: 4,
                    cursor: 'pointer',
                  }}
                >
                  <Cmp style={{ fontSize: 20 }} />
                  <span
                    style={{
                      fontSize: 11,
                      color: '#888',
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                      maxWidth: '100%',
                    }}
                  >
                    {name.replace(/Outlined$/, '')}
                  </span>
                </button>
              </Tooltip>
            );
          })}
        </div>
      </Card>
    </PageContainer>
  );
};

export default IconSelect;
