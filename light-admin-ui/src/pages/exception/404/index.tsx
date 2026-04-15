import { Link } from '@umijs/max';
import { Button, Result } from 'antd';

export default () => (
  <Result
    status="404"
    title="404"
    subTitle="抱歉，你访问的页面不存在。"
    extra={
      <Link to="/" prefetch>
        <Button type="primary">返回首页</Button>
      </Link>
    }
  />
);
