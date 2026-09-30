/**
 * /component/table-select — "picker" pattern: a read-only input that opens a
 * modal containing a table; selecting a row returns the picked value to the
 * outer form. Useful for FK-style fields where the option list is too large
 * for a normal Select.
 */
import { PageContainer } from '@ant-design/pro-components';
import { Button, Card, Form, Input, Modal, Space, Typography } from 'antd';
import React, { useState } from 'react';
import { ResizableTable } from '@/components/ResizableTable';

type Row = { id: number; name: string; org: string };

const DATA: Row[] = Array.from({ length: 20 }, (_, i) => ({
  id: i + 1,
  name: `成员 ${String(i + 1).padStart(2, '0')}`,
  org: i % 2 ? '技术部' : '运营部',
}));

type PickerProps = {
  value?: Row | null;
  onChange?: (v: Row | null) => void;
};

const RowPicker: React.FC<PickerProps> = ({ value, onChange }) => {
  const [open, setOpen] = useState(false);
  const [temp, setTemp] = useState<React.Key[]>(value ? [value.id] : []);

  const confirm = () => {
    const picked = DATA.find((r) => r.id === temp[0]) ?? null;
    onChange?.(picked);
    setOpen(false);
  };

  return (
    <>
      <Space>
        <Input
          readOnly
          placeholder="未选择"
          value={value?.name ?? ''}
          style={{ width: 240 }}
        />
        <Button onClick={() => setOpen(true)}>选择</Button>
        {value && <Button onClick={() => onChange?.(null)}>清空</Button>}
      </Space>
      <Modal
        open={open}
        title="选择成员"
        width={640}
        onCancel={() => setOpen(false)}
        onOk={confirm}
      >
        <ResizableTable<Row>
          rowKey="id"
          size="small"
          pagination={{ pageSize: 8 }}
          dataSource={DATA}
          rowSelection={{
            type: 'radio',
            selectedRowKeys: temp,
            onChange: setTemp,
          }}
          columns={[
            { title: 'ID', dataIndex: 'id', width: 80 },
            { title: '姓名', dataIndex: 'name' },
            { title: '部门', dataIndex: 'org' },
          ]}
          onRow={(row) => ({ onClick: () => setTemp([row.id]) })}
        />
      </Modal>
    </>
  );
};

const TableSelect: React.FC = () => {
  const [form] = Form.useForm();

  return (
    <PageContainer
      title="列表选择器"
      subTitle="输入框 + Modal 表格:FK 字段的常见替代方案"
    >
      <Card>
        <Form form={form} layout="vertical" style={{ maxWidth: 520 }}>
          <Form.Item label="负责人" name="owner">
            <RowPicker />
          </Form.Item>
          <Form.Item shouldUpdate label="当前取值" style={{ marginBottom: 0 }}>
            {() => (
              <Typography.Paragraph copyable>
                {JSON.stringify(form.getFieldValue('owner') ?? null)}
              </Typography.Paragraph>
            )}
          </Form.Item>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default TableSelect;
