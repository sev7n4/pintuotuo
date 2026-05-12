import { useState } from 'react';
import { Button, Card, Form, Input, Typography, message } from 'antd';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import api from '@/services/api';
import { getApiErrorMessage } from '@/utils/apiError';

const { Title, Paragraph } = Typography;

export default function ResetPasswordPage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const tokenFromUrl = (searchParams.get('token') || '').trim();

  const onFinish = async (values: { token: string; password: string }) => {
    setLoading(true);
    try {
      await api.post('/users/password/reset', {
        token: values.token.trim(),
        password: values.password,
      });
      message.success('密码已重置，请使用新密码登录');
      navigate('/login', { replace: true });
    } catch (e) {
      message.error(getApiErrorMessage(e));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 440, margin: '48px auto', padding: '0 16px' }}>
      <Card>
        <Title level={3} style={{ marginTop: 0 }}>
          重置密码
        </Title>
        <Paragraph type="secondary">
          请输入邮件中的重置令牌（开发环境可在「忘记密码」弹窗中复制），并设置新密码。
        </Paragraph>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{ token: tokenFromUrl }}
        >
          <Form.Item
            label="重置令牌"
            name="token"
            rules={[{ required: true, message: '请输入重置令牌' }]}
          >
            <Input.TextArea rows={2} placeholder="粘贴令牌" autoComplete="off" />
          </Form.Item>
          <Form.Item
            label="新密码"
            name="password"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少 6 位' },
            ]}
          >
            <Input.Password placeholder="至少 6 位" autoComplete="new-password" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              确认重置
            </Button>
          </Form.Item>
        </Form>
        <Paragraph style={{ marginBottom: 0, textAlign: 'center' }}>
          <Link to="/forgot-password">没有令牌？申请重置</Link>
          {' · '}
          <Link to="/login">返回登录</Link>
        </Paragraph>
      </Card>
    </div>
  );
}
