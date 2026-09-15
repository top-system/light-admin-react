import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import {
  createConfig,
  deleteConfig,
  getConfigForm,
  queryConfigs,
  refreshConfigCache,
  updateConfig,
} from '@/services/light-admin/config';
import type { Config, ConfigForm } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

type EditState = { mode: 'create' } | { mode: 'edit'; id: string };

const ConfigPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<ConfigForm | undefined>();

  const openCreate = () => {
    setEditInitial({ configName: '', configKey: '', configValue: '' });
    setEdit({ mode: 'create' });
  };

  const openEdit = useCallback(async (id: string) => {
    const form = await getConfigForm(id);
    setEditInitial(form);
    setEdit({ mode: 'edit', id });
  }, []);

  const handleSubmit = async (values: ConfigForm) => {
    if (!edit) return false;
    if (edit.mode === 'create') {
      await createConfig(values);
      message.success('创建成功');
    } else {
      await updateConfig(edit.id, values);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteConfig(id);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const handleRefresh = async () => {
    await refreshConfigCache();
    message.success('配置缓存已刷新');
  };

  const columns = useMemo<ProColumns<Config>[]>(
    () => [
      { title: '名称', dataIndex: 'configName' },
      { title: 'Key', dataIndex: 'configKey' },
      {
        title: 'Value',
        dataIndex: 'configValue',
        search: false,
        ellipsis: true,
      },
      { title: '备注', dataIndex: 'remark', search: false },
      { title: '创建时间', dataIndex: 'createTime', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 160,
        render: (_, row) => (
          <Space size="small">
            <Auth code="sys:config:update">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:config:delete">
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
      <ProTable<Config>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryConfigs)}
        toolBarRender={() => [
          <Auth key="refresh" code="sys:config:refresh">
            <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
              刷新缓存
            </Button>
          </Auth>,
          <Auth key="add" code="sys:config:add">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<ConfigForm>
        key={edit ? (edit.mode === 'edit' ? `e-${edit.id}` : 'c') : 'closed'}
        title={edit?.mode === 'edit' ? '编辑配置' : '新增配置'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText
          name="configName"
          label="名称"
          rules={[{ required: true }]}
        />
        <ProFormText
          name="configKey"
          label="Key"
          rules={[{ required: true }]}
        />
        <ProFormTextArea
          name="configValue"
          label="Value"
          rules={[{ required: true }]}
        />
        <ProFormTextArea name="remark" label="备注" />
      </ModalForm>
    </PageContainer>
  );
};

export default ConfigPage;
