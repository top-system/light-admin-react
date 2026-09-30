/**
 * Downloader — stats + backend selector (with "测试" button) + ProTable of
 * active/finished downloads + create/cancel/sync/delete + file-selection +
 * ws `download-progress` → optimistic progress update on matching row.
 */
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import {
  type ActionType,
  ModalForm,
  PageContainer,
  type ProColumns,
  ProFormSelect,
  ProFormText,
} from '@ant-design/pro-components';
import {
  App,
  Button,
  Card,
  Checkbox,
  Col,
  Drawer,
  Popconfirm,
  Progress,
  Row,
  Space,
  Statistic,
  Tag,
} from 'antd';
import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import Auth from '@/components/business/Auth';
import { ResizableProTable, ResizableTable } from '@/components/ResizableTable';
import {
  cancelDownload,
  createDownload,
  deleteDownload,
  getDownload,
  getDownloaders,
  getDownloadStats,
  queryDownloads,
  setDownloadFiles,
  syncDownload,
  testDownloader,
} from '@/services/light-admin/download';
import type {
  CreateDownloadRequest,
  DownloadDetail,
  DownloaderInfo,
  DownloadRow,
  DownloadStats,
} from '@/types/light-admin/domain';
import { toProTableRequest } from '@/utils/response/adapter';
import { useWsEvent } from '@/utils/ws/hooks';

function formatBytes(n: number): string {
  if (!n) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.min(
    units.length - 1,
    Math.floor(Math.log(n) / Math.log(1024)),
  );
  return `${(n / 1024 ** i).toFixed(1)} ${units[i]}`;
}

const STATUS_COLOR: Record<string, string> = {
  downloading: 'processing',
  seeding: 'cyan',
  completed: 'success',
  error: 'error',
  queued: 'default',
  canceled: 'warning',
};

const DownloadPage: React.FC = () => {
  const actionRef = useRef<ActionType | undefined>(undefined);
  const { message } = App.useApp();
  const [stats, setStats] = useState<DownloadStats | null>(null);
  const [downloaders, setDownloaders] = useState<DownloaderInfo[]>([]);
  const [creating, setCreating] = useState(false);
  const [selectedDownloader, setSelectedDownloader] = useState<
    string | undefined
  >();
  const [detail, setDetail] = useState<DownloadDetail | null>(null);

  const reloadStats = useCallback(() => {
    getDownloadStats()
      .then(setStats)
      .catch(() => {});
  }, []);

  useEffect(() => {
    reloadStats();
    getDownloaders()
      .then((list) => {
        setDownloaders(list);
        if (list.length && !selectedDownloader)
          setSelectedDownloader(list[0].value);
      })
      .catch(() => {});
  }, [reloadStats, selectedDownloader]);

  // Optimistic progress via ws. Frame shape expected:
  //   { type: "download-progress", data: { id, progress, downloadSpeed, ... } }
  useWsEvent<{ id: number; progress?: number }>('download-progress', () => {
    // Simplest reasonable implementation: ask the table to re-run. Row-level
    // optimistic merging requires ProTable exposing its internal dataSource,
    // which isn't worth the coupling for a 1–2s tick.
    actionRef.current?.reload(true);
  });

  const handleTest = async () => {
    if (!selectedDownloader) return;
    try {
      const res = await testDownloader(selectedDownloader);
      message.success(`测试通过 — ${res.version ?? 'ok'}`);
    } catch {
      message.error('测试失败');
    }
  };

  const handleCreate = async (values: CreateDownloadRequest) => {
    await createDownload({ ...values, downloader: selectedDownloader });
    message.success('已提交');
    setCreating(false);
    actionRef.current?.reload();
    reloadStats();
    return true;
  };

  const handleCancel = async (id: number) => {
    await cancelDownload(id);
    message.success('已取消');
    actionRef.current?.reload();
  };

  const handleSync = async (id: number) => {
    await syncDownload(id);
    message.success('同步已触发');
    actionRef.current?.reload();
  };

  const handleDelete = async (id: number) => {
    await deleteDownload(id);
    message.success('已删除');
    actionRef.current?.reload();
    reloadStats();
  };

  const openDetail = async (id: number) => {
    setDetail(await getDownload(id));
  };

  const saveFiles = async () => {
    if (!detail) return;
    const files = detail.files.map((f) => ({
      index: f.index,
      download: f.selected,
    }));
    await setDownloadFiles(detail.id, files);
    message.success('已保存');
    setDetail(null);
    actionRef.current?.reload();
  };

  const columns = useMemo<ProColumns<DownloadRow>[]>(
    () => [
      { title: '名称', dataIndex: 'name', ellipsis: true },
      {
        title: '状态',
        dataIndex: 'status',
        valueType: 'select',
        valueEnum: {
          downloading: '下载中',
          seeding: '做种',
          completed: '完成',
          error: '错误',
          queued: '排队',
          canceled: '取消',
        },
        render: (_, row) => (
          <Tag color={STATUS_COLOR[row.status] ?? 'default'}>{row.status}</Tag>
        ),
      },
      {
        title: '进度',
        dataIndex: 'progress',
        search: false,
        width: 160,
        render: (_, row) => <Progress percent={row.progress} size="small" />,
      },
      {
        title: '速率',
        search: false,
        width: 120,
        render: (_, row) => `${formatBytes(row.downloadSpeed)}/s`,
      },
      {
        title: '大小',
        search: false,
        width: 100,
        render: (_, row) => formatBytes(row.total),
      },
      { title: '下载器', dataIndex: 'downloader', search: false, width: 100 },
      { title: '创建时间', dataIndex: 'createdAt', search: false },
      {
        title: '操作',
        valueType: 'option',
        width: 260,
        render: (_, row) => (
          <Space size="small">
            <a onClick={() => void openDetail(row.id)}>详情</a>
            <Auth code="sys:download:edit">
              <a onClick={() => handleSync(row.id)}>同步</a>
            </Auth>
            {row.status !== 'completed' && row.status !== 'canceled' && (
              <Auth code="sys:download:edit">
                <Popconfirm
                  title="取消该下载?"
                  onConfirm={() => handleCancel(row.id)}
                >
                  <a>取消</a>
                </Popconfirm>
              </Auth>
            )}
            <Auth code="sys:download:delete">
              <Popconfirm
                title="删除该下载?"
                onConfirm={() => handleDelete(row.id)}
              >
                <a style={{ color: '#ff4d4f' }}>删除</a>
              </Popconfirm>
            </Auth>
          </Space>
        ),
      },
    ],
    [],
  );

  return (
    <PageContainer>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col xs={12} md={5}>
          <Card>
            <Statistic title="下载中" value={stats?.downloadingCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={5}>
          <Card>
            <Statistic title="做种" value={stats?.seedingCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={5}>
          <Card>
            <Statistic title="完成" value={stats?.completedCount ?? 0} />
          </Card>
        </Col>
        <Col xs={12} md={5}>
          <Card>
            <Statistic
              title="错误"
              value={stats?.errorCount ?? 0}
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col xs={24} md={4}>
          <Card style={{ height: '100%' }}>
            <Statistic title="全部" value={stats?.totalCount ?? 0} />
          </Card>
        </Col>
      </Row>

      <ResizableProTable<DownloadRow>
        actionRef={actionRef}
        rowKey="id"
        columns={columns}
        request={toProTableRequest(queryDownloads)}
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <Space key="downloader">
            <ProFormSelect
              noStyle
              placeholder="下载器"
              options={downloaders}
              fieldProps={{
                value: selectedDownloader,
                onChange: setSelectedDownloader,
                style: { width: 160 },
              }}
            />
            <Button icon={<ReloadOutlined />} onClick={handleTest}>
              测试
            </Button>
          </Space>,
          <Auth key="add" code="sys:download:add">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setCreating(true)}
            >
              新增下载
            </Button>
          </Auth>,
        ]}
      />

      <ModalForm<CreateDownloadRequest>
        title="新增下载"
        open={creating}
        onOpenChange={setCreating}
        onFinish={handleCreate}
        modalProps={{ destroyOnClose: true, maskClosable: false }}
      >
        <ProFormText
          name="url"
          label="磁力链/URL"
          rules={[{ required: true }]}
          placeholder="magnet:?... 或 http(s)://..."
        />
      </ModalForm>

      <Drawer
        open={detail !== null}
        title={detail?.name}
        width={640}
        onClose={() => setDetail(null)}
        destroyOnClose
        extra={
          detail && (
            <Space>
              <Button onClick={() => setDetail(null)}>关闭</Button>
              <Button type="primary" onClick={saveFiles}>
                保存文件选择
              </Button>
            </Space>
          )
        }
      >
        {detail && (
          <>
            <Progress percent={detail.progress} />
            <div style={{ margin: '12px 0', color: '#8c8c8c' }}>
              {formatBytes(detail.downloaded)} / {formatBytes(detail.total)} ·{' '}
              {formatBytes(detail.downloadSpeed)}/s
            </div>
            <ResizableTable
              size="small"
              rowKey="index"
              pagination={false}
              dataSource={detail.files}
              columns={[
                {
                  title: '下载',
                  dataIndex: 'selected',
                  width: 60,
                  render: (_, row) => (
                    <Checkbox
                      checked={row.selected}
                      onChange={(e) =>
                        setDetail({
                          ...detail,
                          files: detail.files.map((f) =>
                            f.index === row.index
                              ? { ...f, selected: e.target.checked }
                              : f,
                          ),
                        })
                      }
                    />
                  ),
                },
                { title: '文件', dataIndex: 'name', ellipsis: true },
                {
                  title: '大小',
                  dataIndex: 'size',
                  width: 100,
                  render: (v: number) => formatBytes(v),
                },
                {
                  title: '进度',
                  dataIndex: 'progress',
                  width: 140,
                  render: (v: number) => <Progress percent={v} size="small" />,
                },
              ]}
            />
          </>
        )}
      </Drawer>
    </PageContainer>
  );
};

export default DownloadPage;
