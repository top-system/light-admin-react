/**
 * System / Dict — split layout:
 *   left: dict list (ProTable) — select a row to show its items on the right
 *   right: dict items of the selected dict (ProTable)
 *
 * Both tables support create / edit / batch-delete via ModalForm + Popconfirm.
 * When a dict item mutates, we fire an optional `dict-change` invalidation on
 * the shared client cache so open <DictSelect> instances refresh.
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
  ProFormTextArea,
  ProTable,
} from '@ant-design/pro-components';
import { App, Button, Col, Popconfirm, Row, Space } from 'antd';
import React, { useCallback, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import {
  createDict,
  createDictItem,
  deleteDictItems,
  deleteDicts,
  getDictForm,
  getDictItemForm,
  queryDictItems,
  queryDicts,
  updateDict,
  updateDictItem,
} from '@/services/light-admin/dict';
import type {
  Dict,
  DictForm,
  DictItem,
  DictItemForm,
} from '@/types/light-admin/domain';
import { invalidateDict } from '@/utils/dict/cache';
import { toProTableRequest } from '@/utils/response/adapter';

type DictEdit = { mode: 'create' } | { mode: 'edit'; id: string };
type ItemEdit =
  | { mode: 'create'; dictCode: string }
  | { mode: 'edit'; dictCode: string; id: string };

const DictPage: React.FC = () => {
  const { message } = App.useApp();
  const dictActionRef = useRef<ActionType | undefined>(undefined);
  const itemActionRef = useRef<ActionType | undefined>(undefined);

  const [selected, setSelected] = useState<Dict | null>(null);

  const [dictEdit, setDictEdit] = useState<DictEdit | null>(null);
  const [dictInitial, setDictInitial] = useState<DictForm | undefined>();

  const [itemEdit, setItemEdit] = useState<ItemEdit | null>(null);
  const [itemInitial, setItemInitial] = useState<DictItemForm | undefined>();

  // --- Dict list --------------------------------------------------------

  const openDictCreate = () => {
    setDictInitial({ dictCode: '', name: '', status: 1 });
    setDictEdit({ mode: 'create' });
  };
  const openDictEdit = useCallback(async (id: string) => {
    const form = await getDictForm(id);
    setDictInitial(form);
    setDictEdit({ mode: 'edit', id });
  }, []);

  const handleDictSubmit = async (values: DictForm) => {
    if (!dictEdit) return false;
    const payload: DictForm = {
      ...values,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (dictEdit.mode === 'create') {
      await createDict(payload);
      message.success('创建成功');
    } else {
      await updateDict(dictEdit.id, payload);
      message.success('更新成功');
    }
    setDictEdit(null);
    dictActionRef.current?.reload();
    return true;
  };

  const handleDictDelete = async (id: string) => {
    await deleteDicts([id]);
    if (selected?.id === id) setSelected(null);
    message.success('已删除');
    dictActionRef.current?.reload();
  };

  const dictColumns = useMemo<ProColumns<Dict>[]>(
    () => [
      { title: '编码', dataIndex: 'dictCode' },
      { title: '名称', dataIndex: 'name' },
      {
        title: '状态',
        dataIndex: 'status',
        search: false,
        width: 80,
        valueEnum: {
          1: { text: '启用', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
      },
      {
        title: '操作',
        valueType: 'option',
        width: 180,
        render: (_, row) => (
          <Space size="small">
            <Auth code="sys:dict:edit">
              <a onClick={() => void openDictEdit(row.id)}>编辑</a>
            </Auth>
            <Auth code="sys:dict:delete">
              <Popconfirm
                title="删除该字典及其全部字典项?"
                onConfirm={() => handleDictDelete(row.id)}
              >
                <a style={{ color: '#ff4d4f' }}>删除</a>
              </Popconfirm>
            </Auth>
          </Space>
        ),
      },
    ],
    [openDictEdit, selected],
  );

  // --- Dict items -------------------------------------------------------

  const openItemCreate = () => {
    if (!selected) return;
    setItemInitial({
      dictCode: selected.dictCode,
      label: '',
      value: '',
      status: 1,
      sort: 0,
    });
    setItemEdit({ mode: 'create', dictCode: selected.dictCode });
  };

  const openItemEdit = useCallback(async (dictCode: string, id: string) => {
    const form = await getDictItemForm(dictCode, id);
    setItemInitial(form);
    setItemEdit({ mode: 'edit', dictCode, id });
  }, []);

  const handleItemSubmit = async (values: DictItemForm) => {
    if (!itemEdit) return false;
    const payload: DictItemForm = {
      ...values,
      status: values.status !== undefined ? Number(values.status) : undefined,
    };
    if (itemEdit.mode === 'create') {
      await createDictItem(itemEdit.dictCode, payload);
      message.success('创建成功');
    } else {
      await updateDictItem(itemEdit.dictCode, itemEdit.id, payload);
      message.success('更新成功');
    }
    invalidateDict(itemEdit.dictCode);
    setItemEdit(null);
    itemActionRef.current?.reload();
    return true;
  };

  const handleItemDelete = async (dictCode: string, ids: string[]) => {
    await deleteDictItems(dictCode, ids);
    invalidateDict(dictCode);
    message.success('已删除');
    itemActionRef.current?.reload();
  };

  const itemColumns = useMemo<ProColumns<DictItem>[]>(
    () => [
      { title: '标签', dataIndex: 'label' },
      { title: '值', dataIndex: 'value' },
      { title: '标签样式', dataIndex: 'tagType', search: false, width: 100 },
      { title: '排序', dataIndex: 'sort', search: false, width: 80 },
      {
        title: '状态',
        dataIndex: 'status',
        search: false,
        width: 80,
        valueEnum: {
          1: { text: '启用', status: 'Success' },
          0: { text: '禁用', status: 'Default' },
        },
      },
      {
        title: '操作',
        valueType: 'option',
        width: 160,
        render: (_, row) => (
          <Space size="small">
            <Auth code="sys:dict-item:edit">
              <a onClick={() => void openItemEdit(row.dictCode, row.id)}>
                编辑
              </a>
            </Auth>
            <Auth code="sys:dict-item:delete">
              <Popconfirm
                title="删除该字典项?"
                onConfirm={() => handleItemDelete(row.dictCode, [row.id])}
              >
                <a style={{ color: '#ff4d4f' }}>删除</a>
              </Popconfirm>
            </Auth>
          </Space>
        ),
      },
    ],
    [openItemEdit],
  );

  // --- Render -----------------------------------------------------------

  return (
    <PageContainer>
      <Row gutter={16}>
        <Col xs={24} md={10}>
          <ProTable<Dict>
            actionRef={dictActionRef}
            rowKey="id"
            columns={dictColumns}
            request={toProTableRequest(queryDicts)}
            toolBarRender={() => [
              <Auth key="add" code="sys:dict:add">
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  onClick={openDictCreate}
                >
                  新增字典
                </Button>
              </Auth>,
            ]}
            rowClassName={(row) =>
              selected?.id === row.id ? 'ant-table-row-selected' : ''
            }
            onRow={(row) => ({
              onClick: () => setSelected(row),
              style: { cursor: 'pointer' },
            })}
            search={{ labelWidth: 'auto' }}
            headerTitle="字典"
          />
        </Col>
        <Col xs={24} md={14}>
          <ProTable<DictItem>
            actionRef={itemActionRef}
            rowKey="id"
            columns={itemColumns}
            params={{ dictCode: selected?.dictCode }}
            request={
              selected
                ? toProTableRequest((p) => queryDictItems(selected.dictCode, p))
                : async () => ({ data: [], success: true, total: 0 })
            }
            headerTitle={
              selected ? `字典项 — ${selected.name}` : '请先选择字典'
            }
            toolBarRender={() => [
              <Auth key="add" code="sys:dict-item:add">
                <Button
                  type="primary"
                  icon={<PlusOutlined />}
                  disabled={!selected}
                  onClick={openItemCreate}
                >
                  新增项
                </Button>
              </Auth>,
            ]}
            search={{ labelWidth: 'auto' }}
          />
        </Col>
      </Row>

      <ModalForm<DictForm>
        key={
          dictEdit
            ? dictEdit.mode === 'edit'
              ? `de-${dictEdit.id}`
              : 'dc'
            : 'd-closed'
        }
        title={dictEdit?.mode === 'edit' ? '编辑字典' : '新增字典'}
        open={dictEdit !== null}
        onOpenChange={(open) => {
          if (!open) setDictEdit(null);
        }}
        initialValues={dictInitial}
        onFinish={handleDictSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText
          name="dictCode"
          label="编码"
          rules={[{ required: true }]}
          disabled={dictEdit?.mode === 'edit'}
        />
        <ProFormText name="name" label="名称" rules={[{ required: true }]} />
        <ProFormSelect
          name="status"
          label="状态"
          options={[
            { value: 1, label: '启用' },
            { value: 0, label: '禁用' },
          ]}
          initialValue={1}
        />
        <ProFormTextArea name="remark" label="备注" />
      </ModalForm>

      <ModalForm<DictItemForm>
        key={
          itemEdit
            ? itemEdit.mode === 'edit'
              ? `ie-${itemEdit.id}`
              : 'ic'
            : 'i-closed'
        }
        title={itemEdit?.mode === 'edit' ? '编辑字典项' : '新增字典项'}
        open={itemEdit !== null}
        onOpenChange={(open) => {
          if (!open) setItemEdit(null);
        }}
        initialValues={itemInitial}
        onFinish={handleItemSubmit}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText name="label" label="标签" rules={[{ required: true }]} />
        <ProFormText name="value" label="值" rules={[{ required: true }]} />
        <ProFormSelect
          name="tagType"
          label="标签样式"
          valueEnum={{
            '': '默认',
            blue: '蓝色',
            green: '绿色',
            orange: '橙色',
            red: '红色',
            purple: '紫色',
          }}
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
        <ProFormTextArea name="remark" label="备注" />
      </ModalForm>
    </PageContainer>
  );
};

export default DictPage;
