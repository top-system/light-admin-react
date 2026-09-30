/**
 * Member / Tenant — platform-admin CRUD for tenants. Paginated list mirrors
 * the system/role page; tenant `code` is immutable in spirit but editable here
 * for correction. Row-level member isolation keys off the tenant id.
 */
import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import { ResizableProTable } from '@/components/ResizableTable';
import {
  createTenant,
  deleteTenant,
  queryTenants,
  updateTenant,
} from '@/services/light-admin/tenant';
import type { Tenant, TenantForm } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

type EditState = { mode: 'create' } | { mode: 'edit'; id: string };

const TenantPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<TenantForm | undefined>();

  const openCreate = () => {
    setEditInitial({ code: '', name: '', status: 1 });
    setEdit({ mode: 'create' });
  };

  const openEdit = useCallback((row: Tenant) => {
    setEditInitial({
      id: row.id,
      code: row.code,
      name: row.name,
      status: row.status,
    });
    setEdit({ mode: 'edit', id: row.id });
  }, []);

  const handleSubmit = async (values: TenantForm) => {
    if (!edit) return false;
    const payload: TenantForm = {
      ...values,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (edit.mode === 'create') {
      await createTenant(payload);
      message.success('创建成功');
    } else {
      await updateTenant(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteTenant(id);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<Tenant>[]>(
    () => [
      { title: '租户编码', dataIndex: 'code' },
      { title: '租户名称', dataIndex: 'name' },
      {
        title: '状态',
        dataIndex: 'status',
        valueType: 'select',
        valueEnum: {
          1: { text: '正常', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
      },
      { title: '创建时间', dataIndex: 'createTime', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 160,
        render: (_, row) => (
          <Space size="small">
            <Auth code="member:tenant:edit">
              <a onClick={() => openEdit(row)}>编辑</a>
            </Auth>
            <Auth code="member:tenant:delete">
              <Popconfirm
                title="确认删除?"
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
      <ResizableProTable<Tenant>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryTenants)}
        toolBarRender={() => [
          <Auth key="add" code="member:tenant:add">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<TenantForm>
        key={
          edit
            ? edit.mode === 'edit'
              ? `edit-${edit.id}`
              : 'create'
            : 'closed'
        }
        title={edit?.mode === 'edit' ? '编辑租户' : '新增租户'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText
          name="code"
          label="租户编码"
          rules={[{ required: true }]}
        />
        <ProFormText
          name="name"
          label="租户名称"
          rules={[{ required: true }]}
        />
        <ProFormSelect
          name="status"
          label="状态"
          options={[
            { value: 1, label: '正常' },
            { value: 0, label: '禁用' },
          ]}
          initialValue={1}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default TenantPage;
