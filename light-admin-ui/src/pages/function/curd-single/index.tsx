/**
 * /function/curd-single — the entire CRUD pattern inlined in one file with
 * no helpers. A reference for when you want to spike a quick page.
 */
import { PlusOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormText,
} from '@ant-design/pro-components';
import { App, Button, Popconfirm, Space } from 'antd';
import React, { useRef, useState } from 'react';
import { ResizableProTable } from '@/components/ResizableTable';

type Row = { id: number; title: string; tag: string };

let SEQ = 3;
const DATA: Row[] = [
  { id: 1, title: '示例一', tag: 'A' },
  { id: 2, title: '示例二', tag: 'B' },
  { id: 3, title: '示例三', tag: 'C' },
];

const CurdSingle: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Row | null>(null);

  return (
    <PageContainer title="CRUD 单文件" subTitle="全部逻辑 <100 行内完成">
      <ResizableProTable<Row>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        pagination={false}
        request={async () => ({ data: [...DATA], success: true })}
        toolBarRender={() => [
          <Button
            key="add"
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditing(null);
              setOpen(true);
            }}
          >
            新增
          </Button>,
        ]}
        columns={
          [
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '标题', dataIndex: 'title' },
            { title: '标签', dataIndex: 'tag' },
            {
              title: '操作',
              valueType: 'option',
              width: 140,
              render: (_, row) => (
                <Space>
                  <a
                    onClick={() => {
                      setEditing(row);
                      setOpen(true);
                    }}
                  >
                    编辑
                  </a>
                  <Popconfirm
                    title="删除?"
                    onConfirm={() => {
                      const idx = DATA.findIndex((r) => r.id === row.id);
                      if (idx >= 0) DATA.splice(idx, 1);
                      message.success('已删除');
                      actionRef.current?.reload();
                    }}
                  >
                    <a style={{ color: '#ff4d4f' }}>删除</a>
                  </Popconfirm>
                </Space>
              ),
            },
          ] as ProColumns<Row>[]
        }
      />

      <ModalForm<Row>
        key={editing?.id ?? 'new'}
        title={editing ? '编辑' : '新增'}
        open={open}
        onOpenChange={setOpen}
        initialValues={editing ?? { title: '', tag: '' }}
        modalProps={{ destroyOnClose: true }}
        onFinish={async (values) => {
          if (editing) {
            const idx = DATA.findIndex((r) => r.id === editing.id);
            if (idx >= 0) {
              DATA[idx] = {
                id: editing.id,
                title: values.title,
                tag: values.tag,
              };
            }
            message.success('已更新');
          } else {
            SEQ += 1;
            DATA.push({ id: SEQ, title: values.title, tag: values.tag });
            message.success('已新增');
          }
          setOpen(false);
          actionRef.current?.reload();
          return true;
        }}
      >
        <ProFormText name="title" label="标题" rules={[{ required: true }]} />
        <ProFormText name="tag" label="标签" />
      </ModalForm>
    </PageContainer>
  );
};

export default CurdSingle;
