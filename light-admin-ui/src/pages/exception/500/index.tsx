import { Link } from '@umijs/max';
import { Button, Result } from 'antd';

export default () => (
  <Result
    status="500"
    title="500"
    subTitle="服务器内部错误，请稍后重试。"
    extra={
      <Link to="/" prefetch>
        <Button type="primary">返回首页</Button>
      </Link>
    }
  />
);
