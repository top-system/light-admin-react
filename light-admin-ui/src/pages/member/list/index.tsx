/**
 * Member / List — platform-admin view of tenant members. Status toggle and
 * password reset carry the row's tenantId because the backend scopes every
 * member write by tenant. Read-only on identity fields (username is immutable).
 */
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormText,
} from '@ant-design/pro-components';
import { App, Popconfirm, Space } from 'antd';
import React, { useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import { ResizableProTable } from '@/components/ResizableTable';
import {
  queryMembers,
  resetMemberPassword,
  setMemberStatus,
} from '@/services/light-admin/member';
import type { MemberRow } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

const MemberListPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [resetTarget, setResetTarget] = useState<MemberRow | null>(null);

  const handleToggleStatus = async (row: MemberRow) => {
    const next = row.status === 1 ? 0 : 1;
    await setMemberStatus(row.id, { tenantId: row.tenantId, status: next });
    message.success(next === 1 ? '已启用' : '已禁用');
    actionRef.current?.reload();
  };

  const handleReset = async (values: { password: string }) => {
    if (!resetTarget) return false;
    await resetMemberPassword(resetTarget.id, {
      tenantId: resetTarget.tenantId,
      password: values.password,
    });
    message.success('密码已重置');
    setResetTarget(null);
    return true;
  };

  const columns = useMemo<ProColumns<MemberRow>[]>(
    () => [
      { title: '租户', dataIndex: 'tenantId' },
      { title: '用户名', dataIndex: 'username' },
      { title: '昵称', dataIndex: 'nickname', search: false },
      { title: '邮箱', dataIndex: 'email', search: false },
      { title: '手机号', dataIndex: 'mobile', search: false },
      {
        title: '状态',
        dataIndex: 'status',
        valueType: 'select',
        valueEnum: {
          1: { text: '正常', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
      },
      { title: '注册时间', dataIndex: 'createTime', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 200,
        render: (_, row) => (
          <Space size="small">
            <Auth code="member:member:edit">
              <Popconfirm
                title={row.status === 1 ? '确认禁用该会员?' : '确认启用该会员?'}
                onConfirm={() => handleToggleStatus(row)}
              >
                <a>{row.status === 1 ? '禁用' : '启用'}</a>
              </Popconfirm>
            </Auth>
            <Auth code="member:member:edit">
              <a onClick={() => setResetTarget(row)}>重置密码</a>
            </Auth>
          </Space>
        ),
      },
    ],
    [],
  );

  return (
    <PageContainer>
      <ResizableProTable<MemberRow>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryMembers)}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<{ password: string }>
        key={resetTarget ? `reset-${resetTarget.id}` : 'closed'}
        title={`重置密码 — ${resetTarget?.username ?? ''}`}
        open={resetTarget !== null}
        onOpenChange={(open) => {
          if (!open) setResetTarget(null);
        }}
        onFinish={handleReset}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText.Password
          name="password"
          label="新密码"
          rules={[{ required: true, min: 6, message: '密码至少 6 位' }]}
        />
      </ModalForm>
    </PageContainer>
  );
};

export default MemberListPage;
