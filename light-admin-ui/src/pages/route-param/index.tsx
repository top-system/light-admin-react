/**
 * /route-param/route-param-type1 and /route-param/route-param-type2 share this
 * single file; the distinguishing `type` param comes from `useSearchParams`
 * (if the menu passes it as query) or from the pathname.
 */
import { PageContainer } from '@ant-design/pro-components';
import { useLocation, useSearchParams } from '@umijs/max';
import { Card, Descriptions } from 'antd';
import React from 'react';

const RouteParam: React.FC = () => {
  const [params] = useSearchParams();
  const { pathname } = useLocation();

  const paramObj: Record<string, string> = {};
  params.forEach((v, k) => {
    paramObj[k] = v;
  });

  return (
    <PageContainer
      title="路由参数"
      subTitle="从 query string / pathname 读取,多菜单复用同一页面"
    >
      <Card>
        <Descriptions column={1} size="small">
          <Descriptions.Item label="pathname">{pathname}</Descriptions.Item>
          <Descriptions.Item label="query params">
            <code>{JSON.stringify(paramObj)}</code>
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </PageContainer>
  );
};

export default RouteParam;
