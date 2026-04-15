/**
 * /component/text-scroll — horizontal marquee via CSS animation.
 */
import { PageContainer } from '@ant-design/pro-components';
import { Card, InputNumber, Slider, Space, Switch } from 'antd';
import React, { useState } from 'react';

const TextScroll: React.FC = () => {
  const [speed, setSpeed] = useState(12);
  const [paused, setPaused] = useState(false);
  const [content, setContent] = useState(
    '这是一段滚动文本,可以用来在页面顶部显示公告、警告、财经行情等。',
  );

  return (
    <PageContainer title="滚动文本" subTitle="纯 CSS 实现,零 JS 开销">
      <Card>
        <Space wrap style={{ marginBottom: 16 }}>
          <span>速度(秒/圈):</span>
          <Slider
            style={{ width: 200 }}
            min={3}
            max={30}
            value={speed}
            onChange={setSpeed}
          />
          <Switch
            checked={paused}
            onChange={setPaused}
            checkedChildren="暂停"
            unCheckedChildren="运行"
          />
        </Space>
        <div
          style={{
            overflow: 'hidden',
            whiteSpace: 'nowrap',
            background: '#fafafa',
            border: '1px solid #eee',
            padding: '12px 0',
          }}
        >
          <span
            style={{
              display: 'inline-block',
              paddingLeft: '100%',
              animation: `ls-marquee ${speed}s linear infinite`,
              animationPlayState: paused ? 'paused' : 'running',
            }}
          >
            {content}
          </span>
        </div>
        <style>{`
          @keyframes ls-marquee {
            from { transform: translateX(0); }
            to   { transform: translateX(-100%); }
          }
        `}</style>
        <Space style={{ marginTop: 16 }}>
          <span>文本:</span>
          <InputNumber
            style={{ visibility: 'hidden', width: 0 }}
          />
          <input
            style={{ width: 400, padding: 6, border: '1px solid #d9d9d9' }}
            value={content}
            onChange={(e) => setContent(e.target.value)}
          />
        </Space>
      </Card>
    </PageContainer>
  );
};

export default TextScroll;
