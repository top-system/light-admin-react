/**
 * System / Menu — tree-table with in-place CRUD.
 *
 * Menu types follow backend codes:
 *   M = directory, C = menu page, B = button/permission.
 * Type determines which fields are visible in the form.
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
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import {
  createMenu,
  deleteMenu,
  getMenuForm,
  getMenuOptions,
  listMenus,
  updateMenu,
} from '@/services/light-admin/menu';
import type {
  MenuForm,
  MenuNode,
  MenuTypeCode,
} from '@/types/light-admin/domain';

type EditState =
  | { mode: 'create'; parentId?: number }
  | { mode: 'edit'; id: number };

const TYPE_LABEL: Record<MenuTypeCode, string> = {
  M: '目录',
  C: '菜单',
  B: '按钮',
};

const MenuPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<MenuForm | undefined>();
  const [selectedType, setSelectedType] = useState<MenuTypeCode>('C');

  const openCreate = (parentId?: number) => {
    setEditInitial({
      name: '',
      parentId: parentId ?? 0,
      type: 'C',
      visible: 1,
      sort: 0,
    });
    setSelectedType('C');
    setEdit({ mode: 'create', parentId });
  };

  const openEdit = useCallback(async (id: number) => {
    const form = await getMenuForm(id);
    setEditInitial(form);
    setSelectedType(
      (typeof form.type === 'string' ? form.type : 'C') as MenuTypeCode,
    );
    setEdit({ mode: 'edit', id });
  }, []);

  const handleSubmit = async (values: MenuForm) => {
    if (!edit) return false;
    const payload: MenuForm = {
      ...values,
      visible: values.visible !== undefined ? Number(values.visible) : undefined,
    };
    if (edit.mode === 'create') {
      await createMenu(payload);
      message.success('创建成功');
    } else {
      await updateMenu(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: number) => {
    await deleteMenu(id);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<MenuNode>[]>(
    () => [
      { title: '名称', dataIndex: 'name', search: false },
      {
        title: '类型',
        dataIndex: 'type',
        search: false,
        width: 80,
        render: (_, row) => TYPE_LABEL[row.type] ?? row.type,
      },
      { title: '路径', dataIndex: 'routePath', search: false },
      { title: '权限码', dataIndex: 'perm', search: false },
      { title: '排序', dataIndex: 'sort', search: false, width: 80 },
      {
        title: '可见',
        dataIndex: 'visible',
        search: false,
        width: 80,
        render: (_, row) => (row.visible ? '是' : '否'),
      },
      {
        title: '操作',
        valueType: 'option',
        width: 220,
        render: (_, row) => (
          <Space size="small">
            <Auth code="sys:menu:add">
              <a onClick={() => openCreate(row.id)}>新增子项</a>
            </Auth>
            <Auth code="sys:menu:edit">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:menu:delete">
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
      <ProTable<MenuNode>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        search={false}
        pagination={false}
        request={async () => ({ data: await listMenus(), success: true })}
        expandable={{
          childrenColumnName: 'children',
          defaultExpandAllRows: true,
        }}
        toolBarRender={() => [
          <Auth key="add" code="sys:menu:add">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => openCreate()}
            >
              新增
            </Button>
          </Auth>,
        ]}
      />

      <ModalForm<MenuForm>
        key={
          edit
            ? edit.mode === 'edit'
              ? `edit-${edit.id}`
              : 'create'
            : 'closed'
        }
        title={edit?.mode === 'edit' ? '编辑菜单' : '新增菜单'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        onValuesChange={(changed) => {
          if (changed.type) setSelectedType(changed.type as MenuTypeCode);
        }}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormSelect
          name="type"
          label="类型"
          valueEnum={TYPE_LABEL}
          rules={[{ required: true }]}
        />
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormTreeSelect
          name="parentId"
          label="上级"
          request={async () =>
            [
              { value: 0, label: '顶级', children: await getMenuOptions() },
            ] as never
          }
          fieldProps={{ treeDefaultExpandAll: true }}
          initialValue={0}
        />
        {selectedType !== 'B' && (
          <>
            <ProFormText name="routeName" label="路由名" />
            <ProFormText name="routePath" label="路由路径" />
            <ProFormText name="icon" label="图标" />
          </>
        )}
        {selectedType === 'C' && (
          <ProFormText
            name="component"
            label="组件路径"
            tooltip="如 system/user/index"
          />
        )}
        {selectedType !== 'M' && <ProFormText name="perm" label="权限码" />}
        <ProFormDigit name="sort" label="排序" min={0} />
        {selectedType !== 'B' && (
          <ProFormSelect
            name="visible"
            label="可见"
            options={[
              { value: 1, label: '显示' },
              { value: 0, label: '隐藏' },
            ]}
            initialValue={1}
          />
        )}
      </ModalForm>
    </PageContainer>
  );
};

export default MenuPage;
