/**
 * System / User — list + CRUD page.
 *
 * Uses ProTable with ModalForm for create / edit, Popconfirm for delete +
 * password reset. Status / gender columns render via <DictTag>; role multi-
 * select + dept tree-select drive the edit form.
 */
import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormSelect,
  ProFormText,
  ProFormTreeSelect,
  ProTable,
} from '@ant-design/pro-components';
import { App, Avatar, Button, Popconfirm, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import DictTag from '@/components/business/DictTag';
import { getDeptOptions } from '@/services/light-admin/dept';
import { getRoleOptions } from '@/services/light-admin/role';
import {
  createUser,
  deleteUser,
  getUserForm,
  queryUsers,
  resetUserPassword,
  updateUser,
} from '@/services/light-admin/user';
import type { User, UserForm, UserQuery } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

type EditState = { mode: 'create' } | { mode: 'edit'; id: string };

const UserPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<UserForm | undefined>();

  const openCreate = () => {
    setEditInitial({ username: '', nickname: '', status: 1, gender: 0 });
    setEdit({ mode: 'create' });
  };

  const openEdit = useCallback(async (id: string) => {
    const form = await getUserForm(id);
    setEditInitial(form);
    setEdit({ mode: 'edit', id });
  }, []);

  const handleSubmit = async (values: UserForm) => {
    if (!edit) return false;
    const payload: UserForm = {
      ...values,
      gender: values.gender !== undefined ? Number(values.gender) : undefined,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (edit.mode === 'create') {
      await createUser(payload);
      message.success('创建成功');
    } else {
      await updateUser(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteUser(id);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const handleResetPwd = async (id: string) => {
    await resetUserPassword(id);
    message.success('密码已重置');
  };

  const columns = useMemo<ProColumns<User>[]>(
    () => [
      {
        title: '头像',
        dataIndex: 'avatar',
        width: 60,
        search: false,
        render: (_, row) => (
          <Avatar src={row.avatar}>{row.nickname?.[0]}</Avatar>
        ),
      },
      { title: '账号', dataIndex: 'username' },
      { title: '昵称', dataIndex: 'nickname' },
      {
        title: '性别',
        dataIndex: 'gender',
        search: false,
        render: (_, row) => <DictTag code="gender" value={row.gender} />,
      },
      {
        title: '部门',
        dataIndex: 'deptId',
        search: false,
        render: (_, row) => row.deptName || '-',
      },
      { title: '手机号', dataIndex: 'mobile', search: false },
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
            <Auth code="sys:user:edit">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:user:reset-password">
              <Popconfirm
                title="重置为默认密码?"
                onConfirm={() => handleResetPwd(row.id)}
              >
                <a>重置密码</a>
              </Popconfirm>
            </Auth>
            <Auth code="sys:user:delete">
              <Popconfirm
                title="确认删除该用户?"
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
      <ProTable<User, UserQuery>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryUsers)}
        toolBarRender={() => [
          <Auth key="add" code="sys:user:add">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<UserForm>
        key={
          edit
            ? edit.mode === 'edit'
              ? `edit-${edit.id}`
              : 'create'
            : 'closed'
        }
        title={edit?.mode === 'edit' ? '编辑用户' : '新增用户'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText
          name="username"
          label="账号"
          rules={[{ required: true }]}
          disabled={edit?.mode === 'edit'}
        />
        <ProFormText
          name="nickname"
          label="昵称"
          rules={[{ required: true }]}
        />
        {edit?.mode === 'create' && (
          <ProFormText.Password
            name="password"
            label="初始密码"
            rules={[{ required: true, min: 6 }]}
          />
        )}
        <ProFormText name="mobile" label="手机号" />
        <ProFormText name="email" label="邮箱" />
        <ProFormSelect
          name="gender"
          label="性别"
          options={[
            { value: 0, label: '未知' },
            { value: 1, label: '男' },
            { value: 2, label: '女' },
          ]}
        />
        <ProFormTreeSelect
          name="deptId"
          label="部门"
          request={async () => (await getDeptOptions()) as never}
          fieldProps={{ treeDefaultExpandAll: true }}
        />
        <ProFormSelect
          name="roleIds"
          label="角色"
          mode="multiple"
          request={async () => (await getRoleOptions()) as never}
        />
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

export default UserPage;
