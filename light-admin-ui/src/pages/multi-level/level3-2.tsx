import { PageContainer } from '@ant-design/pro-components';
import { Card } from 'antd';
import React from 'react';

const Level3B: React.FC = () => (
  <PageContainer title="三级菜单 - 2" subTitle="多级菜单嵌套演示的叶子节点 B">
    <Card>
      <p>这是路由 <code>/multi-level/multi-level1/multi-level2/multi-level3-2</code> 对应的页面。</p>
    </Card>
  </PageContainer>
);

export default Level3B;
