/**
 * Placeholder landing page. Replaced in Chunk 3+ once real dashboards exist.
 * Serves as the post-login `/` target so the router has somewhere to land.
 */
import { PageContainer } from '@ant-design/pro-components';
import { useModel } from '@umijs/max';
import { Card, Descriptions } from 'antd';
import React from 'react';

const Home: React.FC = () => {
  const { initialState } = useModel('@@initialState');
  const user = initialState?.currentUser;

  return (
    <PageContainer title="欢迎">
      <Card>
        <p>登录成功。业务页面将在 Chunk 3 及之后逐步上线。</p>
        {user && (
          <Descriptions column={2} size="small" style={{ marginTop: 16 }}>
            <Descriptions.Item label="用户">{user.nickname}</Descriptions.Item>
            <Descriptions.Item label="账号">{user.username}</Descriptions.Item>
            <Descriptions.Item label="部门">
              {user.deptName || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="角色">
              {(user.roles ?? []).join(', ') || '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>
    </PageContainer>
  );
};

export default Home;
