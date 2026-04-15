/**
 * /component/curd — self-contained CRUD pattern demo.
 *
 * Shows the shape you'll reach for when building a new business page: ProTable
 * driven by a typed query function, ModalForm for create/edit, Popconfirm for
 * row delete, and toolbar batch-delete via `rowSelection`.
 *
 * Data lives in module state so the page works with no backend.
 */
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormDigit,
  ProFormSelect,
  ProFormText,
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space, Tag } from 'antd';
import React, { useMemo, useRef, useState } from 'react';

type Row = {
  id: number;
  name: string;
  category: 'A' | 'B' | 'C';
  score: number;
  status: 1 | 0;
};

type Query = {
  name?: string;
  category?: Row['category'];
  status?: Row['status'];
  pageNum?: number;
  pageSize?: number;
};

// --- In-memory dataset -----------------------------------------------------

let SEQ = 10;
const DATA: Row[] = Array.from({ length: 27 }, (_, i) => ({
  id: i + 1,
  name: `示例数据 ${String(i + 1).padStart(2, '0')}`,
  category: (['A', 'B', 'C'] as const)[i % 3],
  score: Math.floor(Math.random() * 100),
  status: i % 5 === 0 ? 0 : 1,
}));

function matchQuery(row: Row, q: Query): boolean {
  if (q.name && !row.name.includes(q.name)) return false;
  if (q.category && row.category !== q.category) return false;
  if (q.status !== undefined && row.status !== Number(q.status)) return false;
  return true;
}

async function fakeQuery(q: Query) {
  await new Promise((r) => setTimeout(r, 120));
  const pageNum = q.pageNum ?? 1;
  const pageSize = q.pageSize ?? 10;
  const filtered = DATA.filter((row) => matchQuery(row, q));
  const start = (pageNum - 1) * pageSize;
  return { data: filtered.slice(start, start + pageSize), total: filtered.length };
}

async function fakeCreate(data: Omit<Row, 'id'>) {
  await new Promise((r) => setTimeout(r, 150));
  SEQ += 1;
  DATA.unshift({ id: SEQ, ...data });
}

async function fakeUpdate(id: number, data: Omit<Row, 'id'>) {
  await new Promise((r) => setTimeout(r, 150));
  const idx = DATA.findIndex((r) => r.id === id);
  if (idx >= 0) DATA[idx] = { id, ...data };
}

async function fakeDelete(ids: number[]) {
  await new Promise((r) => setTimeout(r, 120));
  for (let i = DATA.length - 1; i >= 0; i--) {
    if (ids.includes(DATA[i].id)) DATA.splice(i, 1);
  }
}

// --- Component -------------------------------------------------------------

type EditState = { mode: 'create' } | { mode: 'edit'; id: number };

const CurdDemo: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<Partial<Row> | undefined>();
  const [selected, setSelected] = useState<number[]>([]);

  const openCreate = () => {
    setEditInitial({ name: '', category: 'A', score: 0, status: 1 });
    setEdit({ mode: 'create' });
  };

  const openEdit = (row: Row) => {
    setEditInitial(row);
    setEdit({ mode: 'edit', id: row.id });
  };

  const handleSubmit = async (values: Omit<Row, 'id'>) => {
    if (!edit) return false;
    const payload = { ...values, score: Number(values.score), status: Number(values.status) as 0 | 1 };
    if (edit.mode === 'create') {
      await fakeCreate(payload);
      message.success('创建成功');
    } else {
      await fakeUpdate(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (ids: number[]) => {
    await fakeDelete(ids);
    message.success(`已删除 ${ids.length} 条`);
    setSelected([]);
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<Row>[]>(
    () => [
      { title: 'ID', dataIndex: 'id', search: false, width: 80 },
      { title: '名称', dataIndex: 'name' },
      {
        title: '分类',
        dataIndex: 'category',
        valueEnum: { A: 'A 类', B: 'B 类', C: 'C 类' },
      },
      { title: '分数', dataIndex: 'score', search: false, width: 80 },
      {
        title: '状态',
        dataIndex: 'status',
        valueEnum: {
          1: { text: '启用', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
        render: (_, row) => (
          <Tag color={row.status === 1 ? 'success' : 'default'}>
            {row.status === 1 ? '启用' : '禁用'}
          </Tag>
        ),
      },
      {
        title: '操作',
        valueType: 'option',
        width: 140,
        render: (_, row) => (
          <Space size="small">
            <a onClick={() => openEdit(row)}>编辑</a>
            <Popconfirm title="确认删除?" onConfirm={() => handleDelete([row.id])}>
              <a style={{ color: '#ff4d4f' }}>删除</a>
            </Popconfirm>
          </Space>
        ),
      },
    ],
    [],
  );

  return (
    <PageContainer
      title="CRUD 模式示例"
      subTitle="ProTable + ModalForm 组合,适用于 90% 的后台增删改查页面"
    >
      <ProTable<Row, Query>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async (params) => {
          const { data, total } = await fakeQuery(params as Query);
          return { data, total, success: true };
        }}
        rowSelection={{
          selectedRowKeys: selected,
          onChange: (keys) => setSelected(keys.map(Number)),
        }}
        tableAlertRender={({ selectedRowKeys, onCleanSelected }) => (
          <Space>
            已选 {selectedRowKeys.length} 项
            <a onClick={onCleanSelected}>取消选择</a>
          </Space>
        )}
        tableAlertOptionRender={() => (
          <Popconfirm
            title={`批量删除 ${selected.length} 条?`}
            onConfirm={() => handleDelete(selected)}
          >
            <Button size="small" icon={<DeleteOutlined />} danger>
              批量删除
            </Button>
          </Popconfirm>
        )}
        toolBarRender={() => [
          <Button
            key="add"
            type="primary"
            icon={<PlusOutlined />}
            onClick={openCreate}
          >
            新增
          </Button>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<Omit<Row, 'id'>>
        key={edit ? (edit.mode === 'edit' ? `e-${edit.id}` : 'c') : 'closed'}
        title={edit?.mode === 'edit' ? '编辑' : '新增'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect
          name="category"
          label="分类"
          valueEnum={{ A: 'A 类', B: 'B 类', C: 'C 类' }}
          rules={[{ required: true }]}
        />
        <ProFormDigit name="score" label="分数" min={0} max={100} />
        <ProFormSelect
          name="status"
          label="状态"
          options={[
            { value: 1, label: '启用' },
            { value: 0, label: '禁用' },
          ]}
          initialValue={1}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default CurdDemo;
