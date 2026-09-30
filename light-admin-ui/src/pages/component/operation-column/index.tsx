/**
 * /component/operation-column — responsive action column. When the window is
 * narrow (or the action set is long), excess actions collapse into a "更多"
 * dropdown. Pattern:  primary actions always visible + overflow menu.
 */
import { DownOutlined } from '@ant-design/icons';
import { PageContainer, type ProColumns } from '@ant-design/pro-components';
import { Dropdown, Space } from 'antd';
import React, { useMemo } from 'react';
import { ResizableProTable } from '@/components/ResizableTable';

type Row = { id: number; name: string; status: string };

const DATA: Row[] = Array.from({ length: 6 }, (_, i) => ({
  id: i + 1,
  name: `项目 ${i + 1}`,
  status: i % 2 ? '启用' : '禁用',
}));

const ALL_ACTIONS = [
  { key: 'view', label: '查看' },
  { key: 'edit', label: '编辑' },
  { key: 'copy', label: '复制' },
  { key: 'export', label: '导出' },
  { key: 'audit', label: '审计' },
  { key: 'delete', label: '删除', danger: true },
];

const VISIBLE = 2; // first N actions rendered inline; rest go into dropdown

const OperationColumn: React.FC = () => {
  const columns = useMemo<ProColumns<Row>[]>(
    () => [
      { title: 'ID', dataIndex: 'id', width: 60 },
      { title: '名称', dataIndex: 'name' },
      { title: '状态', dataIndex: 'status' },
      {
        title: '操作',
        valueType: 'option',
        width: 260,
        render: () => {
          const inline = ALL_ACTIONS.slice(0, VISIBLE);
          const overflow = ALL_ACTIONS.slice(VISIBLE);
          return (
            <Space size="small">
              {inline.map((a) => (
                <a
                  key={a.key}
                  style={{ color: a.danger ? '#ff4d4f' : undefined }}
                >
                  {a.label}
                </a>
              ))}
              {overflow.length > 0 && (
                <Dropdown
                  menu={{
                    items: overflow.map((a) => ({
                      key: a.key,
                      label: a.label,
                      danger: a.danger,
                    })),
                  }}
                >
                  <a>
                    更多 <DownOutlined />
                  </a>
                </Dropdown>
              )}
            </Space>
          );
        },
      },
    ],
    [],
  );

  return (
    <PageContainer
      title="自适应操作列"
      subTitle="前 N 个常用操作常驻,其余收进「更多」下拉,避免操作列过宽"
    >
      <ResizableProTable<Row>
        rowKey="id"
        columns={columns}
        dataSource={DATA}
        pagination={false}
        search={false}
        toolBarRender={false}
      />
    </PageContainer>
  );
};

export default OperationColumn;
