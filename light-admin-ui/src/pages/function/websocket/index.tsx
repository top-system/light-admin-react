/**
 * /function/websocket — shows the singleton WS client status + a live frame
 * log. Lets you emit arbitrary frames to test server handlers.
 */
import { PageContainer } from '@ant-design/pro-components';
import { Badge, Button, Card, Input, Space, Typography } from 'antd';
import React, { useEffect, useRef, useState } from 'react';
import { ResizableTable } from '@/components/ResizableTable';
import { getWsClient, type WsFrame } from '@/utils/ws/client';

const MAX_LOG = 50;

const WebSocketDemo: React.FC = () => {
  const [log, setLog] = useState<(WsFrame & { _t: number })[]>([]);
  const [type, setType] = useState('ping');
  const [data, setData] = useState('');
  const offRefs = useRef<(() => void)[]>([]);
  const [open, setOpen] = useState(() => getWsClient()?.isOpen() ?? false);

  // Subscribe to every well-known frame type once. Unknown types are ignored
  // because the client dispatches by type; rather than a catch-all we listen
  // to the handful of types the server sends.
  useEffect(() => {
    const client = getWsClient();
    if (!client) return;
    const types = [
      'online-count',
      'dict-change',
      'notice',
      'message',
      'system',
    ];
    for (const t of types) {
      const off = client.on(t, (frame) => {
        setLog((prev) =>
          [{ ...frame, _t: Date.now() }, ...prev].slice(0, MAX_LOG),
        );
      });
      offRefs.current.push(off);
    }
    const timer = setInterval(() => setOpen(client.isOpen()), 1000);
    return () => {
      for (const off of offRefs.current) off();
      offRefs.current = [];
      clearInterval(timer);
    };
  }, []);

  const send = () => {
    const client = getWsClient();
    if (!client) return;
    let payload: unknown;
    try {
      payload = data.trim() ? JSON.parse(data) : undefined;
    } catch {
      payload = data;
    }
    client.emit(type, payload);
  };

  return (
    <PageContainer
      title="WebSocket 调试"
      subTitle="底层走 /utils/ws/client.ts 单例,登录后自动连接"
    >
      <Card
        title={
          <Space>
            连接状态
            <Badge
              status={open ? 'processing' : 'error'}
              text={open ? '已连接' : '未连接'}
            />
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Space>
          <Input
            addonBefore="type"
            value={type}
            onChange={(e) => setType(e.target.value)}
            style={{ width: 240 }}
          />
          <Input
            addonBefore="data"
            value={data}
            onChange={(e) => setData(e.target.value)}
            placeholder='{"foo":"bar"} 或纯字符串'
            style={{ width: 360 }}
          />
          <Button type="primary" onClick={send} disabled={!open}>
            发送
          </Button>
          <Button onClick={() => setLog([])}>清空日志</Button>
        </Space>
      </Card>

      <Card title={`收到的帧(最多保留 ${MAX_LOG} 条)`}>
        <ResizableTable
          size="small"
          rowKey="_t"
          pagination={false}
          dataSource={log}
          locale={{ emptyText: '等待服务端推送…' }}
          columns={[
            {
              title: '时间',
              dataIndex: '_t',
              width: 180,
              render: (v: number) => new Date(v).toLocaleTimeString(),
            },
            { title: 'type', dataIndex: 'type', width: 140 },
            {
              title: 'data',
              dataIndex: 'data',
              render: (v: unknown) => (
                <Typography.Text
                  style={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace' }}
                >
                  {JSON.stringify(v, null, 2)}
                </Typography.Text>
              ),
            },
          ]}
        />
      </Card>
    </PageContainer>
  );
};

export default WebSocketDemo;
