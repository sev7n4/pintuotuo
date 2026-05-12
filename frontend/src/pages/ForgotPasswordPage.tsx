import { useState } from 'react';
import { Button, Card, Form, Input, Typography, message, Modal } from 'antd';
import { Link } from 'react-router-dom';
import api from '@/services/api';
import { getApiErrorMessage } from '@/utils/apiError';

const { Title, Paragraph, Text } = Typography;

type ResetRequestResponse = {
  message?: string;
  debug_token?: string;
};

export default function ForgotPasswordPage() {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const onFinish = async (values: { email: string }) => {
    setLoading(true);
    try {
      const res = await api.post<ResetRequestResponse>('/users/password/reset-request', {
        email: values.email.trim(),
      });
      const data = res.data as ResetRequestResponse;
      message.success(data.message || '若该邮箱已注册，将收到重置说明');
      if (data.debug_token) {
        Modal.info({
          title: '开发环境：重置令牌',
          width: 520,
          content: (
            <div>
              <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                生产环境将通过邮件发送重置链接；当前后端返回了联调令牌，请复制后在「重置密码」页粘贴使用。
              </Paragraph>
              <Text copyable={{ text: data.debug_token }} code>
                {data.debug_token}
              </Text>
            </div>
          ),
        });
      }
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
          忘记密码
        </Title>
        <Paragraph type="secondary">
          请输入注册时使用的邮箱。若账户存在，系统将发送重置说明（开发环境可能直接返回调试令牌）。
        </Paragraph>
        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '邮箱格式不正确' },
            ]}
          >
            <Input placeholder="you@example.com" autoComplete="email" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={loading}>
              发送重置申请
            </Button>
          </Form.Item>
        </Form>
        <Paragraph style={{ marginBottom: 0, textAlign: 'center' }}>
          <Link to="/login">返回登录</Link>
          {' · '}
          <Link to="/reset-password">已有令牌？去重置密码</Link>
        </Paragraph>
      </Card>
    </div>
  );
}
