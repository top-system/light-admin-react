import {
  type ActionType,
  PageContainer,
  type ProColumns,
} from '@ant-design/pro-components';
import { Descriptions, Drawer } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import { ResizableProTable } from '@/components/ResizableTable';
import { queryLogs } from '@/services/light-admin/log';
import type { LogRow } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

const LogPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const [detail, setDetail] = useState<LogRow | null>(null);

  const columns = useMemo<ProColumns<LogRow>[]>(
    () => [
      { title: '模块', dataIndex: 'module' },
      { title: '内容', dataIndex: 'content', search: false, ellipsis: true },
      { title: '方法', dataIndex: 'requestMethod', search: false, width: 80 },
      { title: '路径', dataIndex: 'requestUri', search: false, ellipsis: true },
      { title: 'IP', dataIndex: 'ip', search: false, width: 140 },
      {
        title: '耗时',
        dataIndex: 'executionTime',
        search: false,
        width: 80,
        render: (_, row) => `${row.executionTime}ms`,
      },
      { title: '创建时间', dataIndex: 'createTime', search: false },
      {
        title: '关键词',
        dataIndex: 'keywords',
        hideInTable: true,
        valueType: 'text',
      },
      {
        title: '时间范围',
        dataIndex: 'createTime',
        hideInTable: true,
        valueType: 'dateRange',
        search: { transform: (v: [string, string]) => ({ createTime: v }) },
      },
      {
        title: '操作',
        valueType: 'option',
        width: 80,
        render: (_, row) => <a onClick={() => setDetail(row)}>详情</a>,
      },
    ],
    [],
  );

  return (
    <PageContainer>
      <ResizableProTable<LogRow>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryLogs)}
        search={{ labelWidth: 'auto' }}
      />

      <Drawer
        open={detail !== null}
        title="日志详情"
        width={560}
        onClose={() => setDetail(null)}
        destroyOnClose
      >
        {detail && (
          <Descriptions column={1} size="small">
            <Descriptions.Item label="模块">{detail.module}</Descriptions.Item>
            <Descriptions.Item label="内容">{detail.content}</Descriptions.Item>
            <Descriptions.Item label="方法">
              {detail.requestMethod}
            </Descriptions.Item>
            <Descriptions.Item label="路径">
              {detail.requestUri}
            </Descriptions.Item>
            <Descriptions.Item label="IP">
              {detail.ip}
              {detail.province && ` (${detail.province} ${detail.city})`}
            </Descriptions.Item>
            <Descriptions.Item label="耗时">
              {detail.executionTime}ms
            </Descriptions.Item>
            <Descriptions.Item label="浏览器">
              {detail.browser} {detail.browserVersion}
            </Descriptions.Item>
            <Descriptions.Item label="系统">{detail.os}</Descriptions.Item>
            <Descriptions.Item label="时间">
              {detail.createTime}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default LogPage;
