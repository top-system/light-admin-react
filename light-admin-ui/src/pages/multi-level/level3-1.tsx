import { PageContainer } from '@ant-design/pro-components';
import { Card } from 'antd';
import React from 'react';

const Level3A: React.FC = () => (
  <PageContainer title="三级菜单 - 1" subTitle="多级菜单嵌套演示的叶子节点 A">
    <Card>
      <p>这是路由 <code>/multi-level/multi-level1/multi-level2/multi-level3-1</code> 对应的页面。</p>
      <p>ProLayout 会把上级目录渲染为可折叠的侧栏分组。</p>
    </Card>
  </PageContainer>
);

export default Level3A;
