/**
 * /component/drag — reorderable list using HTML5 drag-and-drop. No external lib.
 */
import { HolderOutlined } from '@ant-design/icons';
import { PageContainer } from '@ant-design/pro-components';
import { Card, List, Typography } from 'antd';
import React, { useRef, useState } from 'react';

type Item = { id: number; label: string };

const INITIAL: Item[] = [
  { id: 1, label: '任务 A' },
  { id: 2, label: '任务 B' },
  { id: 3, label: '任务 C' },
  { id: 4, label: '任务 D' },
  { id: 5, label: '任务 E' },
];

const Drag: React.FC = () => {
  const [items, setItems] = useState(INITIAL);
  const dragFrom = useRef<number | null>(null);

  const onDragStart = (idx: number) => () => {
    dragFrom.current = idx;
  };
  const onDragOver = (idx: number) => (e: React.DragEvent) => {
    e.preventDefault();
    const from = dragFrom.current;
    if (from === null || from === idx) return;
    const next = [...items];
    const [moved] = next.splice(from, 1);
    next.splice(idx, 0, moved);
    dragFrom.current = idx;
    setItems(next);
  };

  return (
    <PageContainer title="拖拽排序" subTitle="HTML5 原生 drag API,零依赖">
      <Card>
        <Typography.Paragraph type="secondary">
          按住左侧手柄,上下拖动调整顺序。松手即保存(示例未持久化)。
        </Typography.Paragraph>
        <List
          bordered
          dataSource={items}
          renderItem={(item, idx) => (
            <List.Item
              draggable
              onDragStart={onDragStart(idx)}
              onDragOver={onDragOver(idx)}
              style={{ cursor: 'move', background: '#fff' }}
            >
              <HolderOutlined style={{ color: '#aaa', marginRight: 12 }} />
              {item.label}
            </List.Item>
          )}
        />
        <Typography.Paragraph style={{ marginTop: 16 }} copyable>
          {JSON.stringify(items.map((i) => i.id))}
        </Typography.Paragraph>
      </Card>
    </PageContainer>
  );
};

export default Drag;
