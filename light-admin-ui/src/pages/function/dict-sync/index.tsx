/**
 * /function/dict-sync — live dict-cache demo:
 *   - enter a dict code, show the cached options
 *   - "失效" clears the cache (next DictSelect render refetches)
 *   - server-side `dict-change` ws frames automatically invalidate too
 */
import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Input, Space, Table, Typography } from 'antd';
import React, { useEffect, useState } from 'react';
import DictSelect from '@/components/business/DictSelect';
import type { DictItemOption } from '@/types/light-admin/domain';
import {
  invalidateDict,
  loadDictOptions,
  peekDictOptions,
  subscribeDict,
} from '@/utils/dict/cache';

const DictSync: React.FC = () => {
  const [code, setCode] = useState('gender');
  const [items, setItems] = useState<DictItemOption[] | undefined>();

  useEffect(() => {
    setItems(peekDictOptions(code));
    const off = subscribeDict(code, () => setItems(peekDictOptions(code)));
    loadDictOptions(code).then((v) => setItems(v)).catch(() => setItems([]));
    return off;
  }, [code]);

  return (
    <PageContainer
      title="字典实时同步"
      subTitle="展示 utils/dict/cache.ts 的缓存 + 订阅 + ws 失效链路"
    >
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Input
            addonBefore="dictCode"
            value={code}
            onChange={(e) => setCode(e.target.value)}
            style={{ width: 260 }}
          />
          <Button
            onClick={() => {
              invalidateDict(code);
            }}
          >
            手动失效
          </Button>
          <Button
            onClick={() => {
              loadDictOptions(code, true).then((v) => setItems(v));
            }}
          >
            强制拉取
          </Button>
        </Space>
      </Card>

      <Card title="DictSelect 实时绑定" style={{ marginBottom: 16 }}>
        <DictSelect code={code} style={{ width: 240 }} placeholder="选一项看看" />
      </Card>

      <Card title={`当前缓存内容 — ${code}`}>
        <Typography.Paragraph type="secondary">
          {items === undefined
            ? '加载中…'
            : items.length === 0
              ? '(空,或该字典不存在)'
              : `共 ${items.length} 项`}
        </Typography.Paragraph>
        <Table<DictItemOption>
          size="small"
          rowKey="value"
          pagination={false}
          dataSource={items ?? []}
          columns={[
            { title: 'value', dataIndex: 'value' },
            { title: 'label', dataIndex: 'label' },
            { title: 'tagType', dataIndex: 'tagType' },
          ]}
        />
      </Card>
    </PageContainer>
  );
};

export default DictSync;
