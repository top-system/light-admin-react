/**
 * System / Notice — admin page: CRUD + publish/revoke + detail drawer.
 * Rich-text content is kept as plain textarea for phase 1; a TinyMCE or
 * similar editor can be lazy-loaded later without touching this wiring.
 */
import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormDependency,
  ProFormSelect,
  ProFormText,
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Drawer, Popconfirm, Space, Tag } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import {
  createNotice,
  deleteNotices,
  getNoticeDetail,
  getNoticeForm,
  publishNotice,
  queryNotices,
  revokeNotice,
  updateNotice,
} from '@/services/light-admin/notice';
import { getUserOptions } from '@/services/light-admin/user';
import type {
  Notice,
  NoticeDetail,
  NoticeForm,
} from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

type EditState = { mode: 'create' } | { mode: 'edit'; id: string };

const STATUS_ENUM: Record<number, { text: string; status: string }> = {
  0: { text: '草稿', status: 'Default' },
  1: { text: '已发布', status: 'Success' },
  2: { text: '已撤销', status: 'Warning' },
};

const LEVEL_LABEL: Record<string, { text: string; color: string }> = {
  L: { text: '低', color: 'default' },
  M: { text: '中', color: 'blue' },
  H: { text: '高', color: 'red' },
};

const NoticePage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<NoticeForm | undefined>();
  const [detail, setDetail] = useState<NoticeDetail | null>(null);

  const openCreate = () => {
    setEditInitial({ title: '', type: 1, level: 'M', targetType: 1 });
    setEdit({ mode: 'create' });
  };

  const openEdit = useCallback(async (id: string) => {
    const form = await getNoticeForm(id);
    setEditInitial(form);
    setEdit({ mode: 'edit', id });
  }, []);

  const openDetail = async (id: string) => {
    setDetail(await getNoticeDetail(id));
  };

  const handleSubmit = async (values: NoticeForm) => {
    if (!edit) return false;
    // Defensive: even with numeric-valued options, some antd versions stringify
    // enum values when round-tripped via initialValues. Coerce before POST.
    const payload: NoticeForm = {
      ...values,
      type: Number(values.type),
      targetType: Number(values.targetType),
    };
    if (edit.mode === 'create') {
      await createNotice(payload);
      message.success('创建成功');
    } else {
      await updateNotice(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteNotices([id]);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const handlePublish = async (id: string) => {
    await publishNotice(id);
    message.success('已发布');
    actionRef.current?.reload();
  };

  const handleRevoke = async (id: string) => {
    await revokeNotice(id);
    message.success('已撤销');
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<Notice>[]>(
    () => [
      { title: '标题', dataIndex: 'title' },
      {
        title: '类型',
        dataIndex: 'type',
        valueEnum: {
          1: '公告',
          2: '通知',
          3: '系统',
        },
      },
      {
        title: '级别',
        dataIndex: 'level',
        search: false,
        width: 80,
        render: (_, row) => {
          const m = LEVEL_LABEL[row.level];
          return m ? (
            <Tag color={m.color}>{m.text}</Tag>
          ) : (
            <Tag>{row.level}</Tag>
          );
        },
      },
      {
        title: '状态',
        dataIndex: 'publishStatus',
        valueEnum: STATUS_ENUM,
      },
      {
        title: '发布人',
        dataIndex: 'publisherName',
        search: false,
        render: (_, row) => row.publisherName || '-',
      },
      { title: '发布时间', dataIndex: 'publishTime', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 260,
        render: (_, row) => (
          <Space size="small">
            <a onClick={() => void openDetail(row.id)}>详情</a>
            {row.publishStatus !== 1 ? (
              <Auth code="sys:notice:publish">
                <Popconfirm
                  title="发布此通知?"
                  onConfirm={() => handlePublish(row.id)}
                >
                  <a>发布</a>
                </Popconfirm>
              </Auth>
            ) : (
              <Auth code="sys:notice:revoke">
                <Popconfirm
                  title="撤销此通知?"
                  onConfirm={() => handleRevoke(row.id)}
                >
                  <a>撤销</a>
                </Popconfirm>
              </Auth>
            )}
            <Auth code="sys:notice:edit">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:notice:delete">
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
      <ProTable<Notice>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryNotices)}
        toolBarRender={() => [
          <Auth key="add" code="sys:notice:add">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<NoticeForm>
        key={edit ? (edit.mode === 'edit' ? `e-${edit.id}` : 'c') : 'closed'}
        title={edit?.mode === 'edit' ? '编辑通知' : '新增通知'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
        width={720}
      >
        <ProFormText name="title" label="标题" rules={[{ required: true }]} />
        <ProFormSelect
          name="type"
          label="类型"
          options={[
            { value: 1, label: '公告' },
            { value: 2, label: '通知' },
            { value: 3, label: '系统' },
          ]}
          rules={[{ required: true }]}
        />
        <ProFormSelect
          name="level"
          label="级别"
          options={[
            { value: 'L', label: '低' },
            { value: 'M', label: '中' },
            { value: 'H', label: '高' },
          ]}
          rules={[{ required: true }]}
        />
        <ProFormSelect
          name="targetType"
          label="目标"
          options={[
            { value: 1, label: '全体' },
            { value: 2, label: '指定用户' },
          ]}
          rules={[{ required: true }]}
        />
        <ProFormDependency name={['targetType']}>
          {({ targetType }) =>
            Number(targetType) === 2 ? (
              <ProFormSelect
                name="targetUserIds"
                label="接收用户"
                mode="multiple"
                request={async () => {
                  const opts = await getUserOptions();
                  return opts.map((o) => ({
                    value: String(o.value),
                    label: o.label,
                  }));
                }}
                rules={[{ required: true, message: '请选择至少一个用户' }]}
                fieldProps={{ showSearch: true, optionFilterProp: 'label' }}
              />
            ) : null
          }
        </ProFormDependency>
        <ProFormTextArea name="content" label="内容" fieldProps={{ rows: 8 }} />
      </ModalForm>

      <Drawer
        open={detail !== null}
        title={detail?.title}
        width={560}
        onClose={() => setDetail(null)}
        destroyOnClose
      >
        {detail && (
          <>
            <div style={{ marginBottom: 12, color: '#8c8c8c' }}>
              {detail.publisherName} · {detail.publishTime ?? '未发布'}
            </div>
            <div style={{ whiteSpace: 'pre-wrap' }}>{detail.content}</div>
          </>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default NoticePage;
