/**
 * NoticeBell — header bell icon. Fetches `/notices/my` on mount, updates
 * unread count on every `notice` ws frame. Click an item to open a detail
 * drawer; opening the dropdown marks all as read.
 */
import { BellOutlined } from '@ant-design/icons';
import { Badge, Drawer, Dropdown, List, Tag } from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import {
  getNoticeDetail,
  listMyNotices,
  readAllNotices,
} from '@/services/light-admin/notice';
import type { NoticeDetail, UserNotice } from '@/types/light-admin/domain';
import { useWsEvent } from '@/utils/ws/hooks';

export const NoticeBell: React.FC = () => {
  const [items, setItems] = useState<UserNotice[]>([]);
  const [detail, setDetail] = useState<NoticeDetail | null>(null);

  const refresh = useCallback(() => {
    listMyNotices({ pageSize: 10 })
      .then((r) => setItems(r.data))
      .catch(() => {
        /* bell is non-critical; swallow */
      });
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  useWsEvent('notice', refresh);

  const unread = items.filter((n) => !n.isRead).length;

  const onOpen = async (open: boolean) => {
    if (!open || unread === 0) return;
    try {
      await readAllNotices();
      setItems((prev) => prev.map((n) => ({ ...n, isRead: 1 })));
    } catch {
      /* ignore */
    }
  };

  const openDetail = async (noticeId: string) => {
    try {
      const d = await getNoticeDetail(noticeId);
      setDetail(d);
    } catch {
      /* ignore */
    }
  };

  return (
    <>
      <Dropdown
        trigger={['click']}
        onOpenChange={onOpen}
        popupRender={() => (
          <div
            style={{
              background: '#fff',
              width: 320,
              padding: 8,
              boxShadow: '0 2px 8px rgba(0,0,0,0.12)',
            }}
          >
            <List
              size="small"
              dataSource={items}
              locale={{ emptyText: '暂无通知' }}
              renderItem={(n) => (
                <List.Item
                  style={{ cursor: 'pointer' }}
                  onClick={() => void openDetail(n.noticeId)}
                >
                  <List.Item.Meta
                    title={
                      <span>
                        {!n.isRead && <Tag color="red">新</Tag>}
                        {n.title}
                      </span>
                    }
                    description={n.publishTime ?? ''}
                  />
                </List.Item>
              )}
            />
          </div>
        )}
      >
        <span style={{ padding: '0 8px', cursor: 'pointer' }}>
          <Badge count={unread} size="small">
            <BellOutlined style={{ fontSize: 16 }} />
          </Badge>
        </span>
      </Dropdown>

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
              {detail.publisherName || '-'} · {detail.publishTime ?? '未发布'}
            </div>
            <div
              style={{ whiteSpace: 'pre-wrap' }}
              // Notice content is already HTML-escaped server-side when
              // stored via the rich-text editor, so rendering as HTML is safe.
              dangerouslySetInnerHTML={{ __html: detail.content }}
            />
          </>
        )}
      </Drawer>
    </>
  );
};

export default NoticeBell;
