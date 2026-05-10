import { Button, Result } from 'antd';
import { Link, useLocation } from 'react-router-dom';

export default function NotFoundPage() {
  const location = useLocation();

  return (
    <div style={{ padding: '48px 16px', maxWidth: 560, margin: '0 auto' }}>
      <Result
        status="404"
        title="页面不存在"
        subTitle={
          <>
            未找到路径 <code style={{ wordBreak: 'break-all' }}>{location.pathname}</code>
            ，可能链接已失效或地址有误。
          </>
        }
        extra={[
          <Link key="home" to="/">
            <Button type="primary">返回首页</Button>
          </Link>,
          <Link key="catalog" to="/catalog">
            <Button>去卖场</Button>
          </Link>,
        ]}
      />
    </div>
  );
}
