/**
 * Task queue — stats cards (4 counts) + ProTable with type/status filter +
 * detail drawer + delete.
 */
import {
  type ActionType,
  PageContainer,
  type ProColumns,
  ProTable,
} from '@ant-design/pro-components';
import {
  App,
  Badge,
  Card,
  Col,
  Descriptions,
  Drawer,
  Popconfirm,
  Row,
  Statistic,
} from 'antd';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import Auth from '@/components/business/Auth';
import {
  deleteTask,
  getTask,
  getTaskStats,
  getTaskTypes,
  queryTasks,
} from '@/services/light-admin/task';
import type {
  TaskRow,
  TaskStats,
  TaskStatus,
  TaskType,
} from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';

const STATUS_ENUM: Record<TaskStatus, { text: string; status: string }> = {
  queued: { text: '排队', status: 'Default' },
  processing: { text: '执行中', status: 'Processing' },
  completed: { text: '成功', status: 'Success' },
  error: { text: '失败', status: 'Error' },
  canceled: { text: '已取消', status: 'Warning' },
  suspending: { text: '暂停', status: 'Warning' },
};

const TaskPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [stats, setStats] = useState<TaskStats | null>(null);
  const [types, setTypes] = useState<TaskType[]>([]);
  const [detail, setDetail] = useState<TaskRow | null>(null);

  useEffect(() => {
    getTaskStats()
      .then(setStats)
      .catch(() => {});
    getTaskTypes()
      .then(setTypes)
      .catch(() => {});
  }, []);

  const handleDelete = async (id: number) => {
    await deleteTask(id);
    message.success('已删除');
    actionRef.current?.reload();
    getTaskStats()
      .then(setStats)
      .catch(() => {});
  };

  const columns = useMemo<ProColumns<TaskRow>[]>(
    () => [
      { title: 'ID', dataIndex: 'id', search: false, width: 80 },
      {
        title: '类型',
        dataIndex: 'type',
        valueType: 'select',
        valueEnum: types.reduce<Record<string, string>>((acc, t) => {
          acc[t.value] = t.label;
          return acc;
        }, {}),
      },
      {
        title: '状态',
        dataIndex: 'status',
        valueType: 'select',
        valueEnum: STATUS_ENUM,
      },
      { title: '关联ID', dataIndex: 'correlationId', search: false },
      { title: '重试', dataIndex: 'retryCount', search: false, width: 80 },
      {
        title: '耗时',
        dataIndex: 'executedDuration',
        search: false,
        width: 100,
        render: (_, row) => `${row.executedDuration}ms`,
      },
      { title: '创建时间', dataIndex: 'createdAt', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 160,
        render: (_, row) => (
          <>
            <a onClick={() => getTask(row.id).then(setDetail)}>详情</a>
            <Auth code="sys:task:delete">
              <Popconfirm
                title="删除该任务?"
                onConfirm={() => handleDelete(row.id)}
              >
                <a style={{ color: '#ff4d4f', marginLeft: 8 }}>删除</a>
              </Popconfirm>
            </Auth>
          </>
        ),
      },
    ],
    [types],
  );

  return (
    <PageContainer>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title="排队" value={stats?.queuedCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title="执行中" value={stats?.processingCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic title="成功" value={stats?.completedCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={6}>
          <Card>
            <Statistic
              title="失败"
              value={stats?.errorCount ?? 0}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
      </Row>

      <ProTable<TaskRow>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryTasks)}
        search={{ labelWidth: 'auto' }}
      />

      <Drawer
        open={detail !== null}
        title={`任务 #${detail?.id}`}
        width={600}
        onClose={() => setDetail(null)}
        destroyOnClose
      >
        {detail && (
          <Descriptions column={1} size="small">
            <Descriptions.Item label="类型">{detail.type}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Badge
                status={STATUS_ENUM[detail.status]?.status as never}
                text={STATUS_ENUM[detail.status]?.text ?? detail.status}
              />
            </Descriptions.Item>
            <Descriptions.Item label="关联ID">
              {detail.correlationId}
            </Descriptions.Item>
            <Descriptions.Item label="Owner">
              {detail.ownerId}
            </Descriptions.Item>
            <Descriptions.Item label="重试次数">
              {detail.retryCount}
            </Descriptions.Item>
            <Descriptions.Item label="耗时">
              {detail.executedDuration}ms
            </Descriptions.Item>
            <Descriptions.Item label="错误">
              {detail.error || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="错误历史">
              <pre style={{ maxHeight: 200, overflow: 'auto', margin: 0 }}>
                {detail.errorHistory || '-'}
              </pre>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {detail.createdAt}
            </Descriptions.Item>
            <Descriptions.Item label="更新时间">
              {detail.updatedAt}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default TaskPage;
