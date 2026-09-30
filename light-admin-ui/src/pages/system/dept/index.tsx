/**
 * System / Dept — tree-table + CRUD. Backend returns a nested tree so we
 * skip pagination and search and use ProTable's `expandable`.
 */
import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormDigit,
  ProFormSelect,
  ProFormText,
  ProFormTreeSelect,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import { ResizableProTable } from '@/components/ResizableTable';
import {
  createDept,
  deleteDepts,
  getDeptForm,
  getDeptOptions,
  queryDepts,
  updateDept,
} from '@/services/light-admin/dept';
import type { Dept, DeptForm } from '@/types/light-admin/domain';

type EditState =
  | { mode: 'create'; parentId?: string }
  | { mode: 'edit'; id: string };

const DeptPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<DeptForm | undefined>();

  const openCreate = (parentId?: string) => {
    setEditInitial({
      name: '',
      code: '',
      parentId: parentId ?? '',
      status: 1,
      sort: 0,
    });
    setEdit({ mode: 'create', parentId });
  };

  const openEdit = useCallback(async (id: string) => {
    const form = await getDeptForm(id);
    setEditInitial(form);
    setEdit({ mode: 'edit', id });
  }, []);

  const handleSubmit = async (values: DeptForm) => {
    if (!edit) return false;
    const payload: DeptForm = {
      ...values,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (edit.mode === 'create') {
      await createDept(payload);
      message.success('创建成功');
    } else {
      await updateDept(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteDepts([id]);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<Dept>[]>(
    () => [
      { title: '名称', dataIndex: 'name' },
      { title: '编码', dataIndex: 'code' },
      { title: '排序', dataIndex: 'sort', search: false, width: 80 },
      {
        title: '状态',
        dataIndex: 'status',
        valueType: 'select',
        valueEnum: {
          1: { text: '启用', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
      },
      { title: '创建时间', dataIndex: 'createTime', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 220,
        render: (_, row) => (
          <Space size="small">
            <Auth code="sys:dept:add">
              <a onClick={() => openCreate(row.id)}>新增子项</a>
            </Auth>
            <Auth code="sys:dept:edit">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:dept:delete">
              <Popconfirm
                title="删除此项及其所有子项?"
                onConfirm={() => handleDelete(row.id)}
              >
                <a style={{ color: '#ff4d4f' }}>删除</a>
              </Popconfirm>
            </Auth>
          </Space>
        ),
      },
    ],
    [openEdit],
  );

  return (
    <PageContainer>
      <ResizableProTable<Dept>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={async (params) => {
          const data = await queryDepts(params);
          return { data, success: true };
        }}
        pagination={false}
        expandable={{
          childrenColumnName: 'children',
          defaultExpandAllRows: true,
        }}
        toolBarRender={() => [
          <Auth key="add" code="sys:dept:add">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => openCreate()}
            >
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<DeptForm>
        key={
          edit
            ? edit.mode === 'edit'
              ? `edit-${edit.id}`
              : 'create'
            : 'closed'
        }
        title={edit?.mode === 'edit' ? '编辑部门' : '新增部门'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormText name="code" label="编码" rules={[{ required: true }]} />
        <ProFormTreeSelect
          name="parentId"
          label="上级部门"
          request={async () =>
            [
              { value: '', label: '顶级', children: await getDeptOptions() },
            ] as never
          }
          fieldProps={{ treeDefaultExpandAll: true }}
          initialValue=""
        />
        <ProFormDigit name="sort" label="排序" min={0} />
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

export default DeptPage;
