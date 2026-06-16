/**
 * System / Role — list, CRUD, and menu assignment.
 *
 * Assignment UX: row action "分配权限" opens a Drawer with a Tree of menus
 * (from /menus — all menus, not just the current user's). The tree is
 * prefilled with the role's current menuIds; submit fires assignRoleMenus.
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
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Drawer, Popconfirm, Space, Tree } from 'antd';
import type { DataNode } from 'antd/es/tree';
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import Auth from '@/components/business/Auth';
import { listMenus } from '@/services/light-admin/menu';
import {
  assignRoleMenus,
  createRole,
  deleteRole,
  getRoleForm,
  getRoleMenuIds,
  queryRoles,
  updateRole,
} from '@/services/light-admin/role';
import type { MenuNode, Role, RoleForm } from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

type EditState = { mode: 'create' } | { mode: 'edit'; id: string };

function menusToTreeData(nodes: MenuNode[]): DataNode[] {
  return nodes.map((n) => ({
    key: n.id,
    title: n.name,
    children: n.children?.length ? menusToTreeData(n.children) : undefined,
  }));
}

const RolePage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [edit, setEdit] = useState<EditState | null>(null);
  const [editInitial, setEditInitial] = useState<RoleForm | undefined>();

  const [assignRole, setAssignRole] = useState<Role | null>(null);
  const [menuTree, setMenuTree] = useState<DataNode[]>([]);
  const [checkedMenus, setCheckedMenus] = useState<React.Key[]>([]);
  const [assignLoading, setAssignLoading] = useState(false);

  useEffect(() => {
    if (!assignRole) return;
    setAssignLoading(true);
    Promise.all([listMenus(), getRoleMenuIds(assignRole.id)])
      .then(([nodes, ids]) => {
        setMenuTree(menusToTreeData(nodes));
        setCheckedMenus(ids);
      })
      .finally(() => setAssignLoading(false));
  }, [assignRole]);

  const openCreate = () => {
    setEditInitial({ name: '', code: '', status: 1, sort: 0 });
    setEdit({ mode: 'create' });
  };

  const openEdit = useCallback(async (id: string) => {
    const form = await getRoleForm(id);
    setEditInitial(form);
    setEdit({ mode: 'edit', id });
  }, []);

  const handleSubmit = async (values: RoleForm) => {
    if (!edit) return false;
    const payload: RoleForm = {
      ...values,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (edit.mode === 'create') {
      await createRole(payload);
      message.success('创建成功');
    } else {
      await updateRole(edit.id, payload);
      message.success('更新成功');
    }
    setEdit(null);
    actionRef.current?.reload();
    return true;
  };

  const handleDelete = async (id: string) => {
    await deleteRole(id);
    message.success('已删除');
    actionRef.current?.reload();
  };

  const handleAssign = async () => {
    if (!assignRole) return;
    await assignRoleMenus(assignRole.id, checkedMenus.map(String));
    message.success('权限已保存');
    setAssignRole(null);
  };

  const columns = useMemo<ProColumns<Role>[]>(
    () => [
      { title: '名称', dataIndex: 'name' },
      { title: '代码', dataIndex: 'code' },
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
            <Auth code="sys:role:edit">
              <a onClick={() => void openEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:role:edit">
              <a onClick={() => setAssignRole(row)}>分配权限</a>
            </Auth>
            <Auth code="sys:role:delete">
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
      <ProTable<Role>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryRoles)}
        toolBarRender={() => [
          <Auth key="add" code="sys:role:add">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新增
            </Button>
          </Auth>,
        ]}
        search={{ labelWidth: 'auto' }}
      />

      <ModalForm<RoleForm>
        key={
          edit
            ? edit.mode === 'edit'
              ? `edit-${edit.id}`
              : 'create'
            : 'closed'
        }
        title={edit?.mode === 'edit' ? '编辑角色' : '新增角色'}
        open={edit !== null}
        onOpenChange={(open) => {
          if (!open) setEdit(null);
        }}
        initialValues={editInitial}
        onFinish={handleSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormText name="code" label="代码" rules={[{ required: true }]} />
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

      <Drawer
        open={assignRole !== null}
        title={`分配权限 — ${assignRole?.name ?? ''}`}
        width={420}
        onClose={() => setAssignRole(null)}
        destroyOnClose
        extra={
          <Space>
            <Button onClick={() => setAssignRole(null)}>取消</Button>
            <Button
              type="primary"
              onClick={handleAssign}
              loading={assignLoading}
            >
              保存
            </Button>
          </Space>
        }
      >
        <Tree
          checkable
          selectable={false}
          treeData={menuTree}
          checkedKeys={checkedMenus}
          onCheck={(keys) =>
            setCheckedMenus(Array.isArray(keys) ? keys : keys.checked)
          }
          defaultExpandAll
        />
      </Drawer>
    </PageContainer>
  );
};

export default RolePage;
