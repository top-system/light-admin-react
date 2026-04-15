/**
 * /component/upload — <FileUpload> showcase. Three common shapes:
 *   1. Uncontrolled demo (local useState)
 *   2. Inside ProForm with other fields
 *   3. Image-only with preview
 *
 * The underlying wire call is always POST /files (see services/light-admin/file.ts).
 */
import {
  PageContainer,
  ProForm,
  ProFormText,
} from '@ant-design/pro-components';
import { App, Avatar, Card, Col, Row, Typography } from 'antd';
import React, { useState } from 'react';
import FileUpload from '@/components/business/FileUpload';

const { Paragraph, Title } = Typography;

const UploadDemo: React.FC = () => {
  const { message } = App.useApp();
  const [anyFile, setAnyFile] = useState<string | undefined>();
  const [avatar, setAvatar] = useState<string | undefined>();

  return (
    <PageContainer
      title="文件上传组件"
      subTitle="统一走 POST /files,本地存储或 OSS 由后端 config.OSS 决定"
    >
      <Row gutter={16}>
        <Col xs={24} md={12}>
          <Card title="1. 任意文件">
            <Paragraph type="secondary">
              最常见的用法。<code>value</code> 是一个 url,<code>onChange</code> 传回新 url
              或 <code>undefined</code>(移除时)。
            </Paragraph>
            <FileUpload
              value={anyFile}
              onChange={setAnyFile}
              maxSizeMb={20}
              buttonText="选择文件"
            />
            <Paragraph copyable={!!anyFile} style={{ marginTop: 8 }}>
              当前值: {anyFile ?? '(空)'}
            </Paragraph>
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card title="2. 头像 / 图片">
            <Paragraph type="secondary">
              通过 <code>accept="image/*"</code> 限定图片,
              <code>maxSizeMb</code> 兜底校验。
            </Paragraph>
            <Avatar size={64} src={avatar} style={{ marginBottom: 12 }}>
              头像
            </Avatar>
            <FileUpload
              value={avatar}
              onChange={setAvatar}
              accept="image/*"
              maxSizeMb={5}
              buttonText="上传头像"
            />
          </Card>
        </Col>

        <Col xs={24}>
          <Card title="3. 嵌入 ProForm">
            <Paragraph type="secondary">
              放进 <code>ProForm</code> 当成一个普通字段,<code>value/onChange</code>
              由 Form 自动注入。
            </Paragraph>
            <ProForm
              onFinish={async (values) => {
                message.success('提交成功');
                // biome-ignore lint/suspicious/noConsole: demo
                console.log('submit', values);
                return true;
              }}
              layout="vertical"
              submitter={{
                searchConfig: { submitText: '提交', resetText: '清空' },
              }}
            >
              <ProFormText name="title" label="标题" rules={[{ required: true }]} />
              <ProForm.Item name="attachment" label="附件">
                <FileUpload buttonText="上传附件" />
              </ProForm.Item>
            </ProForm>
          </Card>
        </Col>

        <Col xs={24}>
          <Card title="接入指南">
            <Title level={5}>1) 后端</Title>
            <Paragraph>
              <code>POST /api/v1/files</code>(multipart,字段名 <code>file</code>),
              返回 <code>{`{ name, url }`}</code>。
              后端 OSS 类型(local / minio / aliyun)在
              <code>config.yaml</code> 的 <code>OSS</code> 段切换。
            </Paragraph>
            <Title level={5}>2) 前端</Title>
            <Paragraph>
              <code>{"import FileUpload from '@/components/business/FileUpload';"}</code>
              。不需要 token 注入:<code>request</code> 拦截器已经统一处理
              <code>Authorization</code>。
            </Paragraph>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default UploadDemo;
