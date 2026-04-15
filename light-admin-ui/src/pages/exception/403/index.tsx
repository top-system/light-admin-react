import { Link } from '@umijs/max';
import { Button, Result } from 'antd';

export default () => (
  <Result
    status="403"
    title="403"
    subTitle="抱歉，你没有访问该页面的权限。"
    extra={
      <Link to="/" prefetch>
        <Button type="primary">返回首页</Button>
      </Link>
    }
  />
);
