import React, { useEffect } from 'react'
import { Form, Input, Button, Card, message, Checkbox } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@stores/authStore'

export const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const { login, isLoading, user, isAuthenticated } = useAuthStore()
  const [form] = Form.useForm()

  const onFinish = async (values: { email: string; password: string; rememberMe?: boolean }) => {
    try {
      await login(values.email, values.password, values.rememberMe || false)
      message.success('登录成功')
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : '登录失败，请检查邮箱和密码'
      message.error(errorMsg)
    }
  }

  useEffect(() => {
    if (isAuthenticated && user) {
      if (user.role === 'admin') {
        navigate('/admin', { replace: true })
      } else if (user.role === 'merchant') {
        navigate('/merchant', { replace: true })
      } else {
        navigate('/products', { replace: true })
      }
    }
  }, [isAuthenticated, user, navigate])

  return (
    <div className="auth-page">
      <Card className="auth-card" title="拼脱脱 - 登录" style={{ maxWidth: 400 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
          initialValues={{ rememberMe: true }}
        >
          <Form.Item
            label="邮箱"
            name="email"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '邮箱格式不正确' },
            ]}
          >
            <Input placeholder="example@email.com" />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password placeholder="输入密码" />
          </Form.Item>

          <Form.Item name="rememberMe" valuePropName="checked">
            <Checkbox>记住我</Checkbox>
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={isLoading}>
              登录
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <span>没有账户？ </span>
            <Button type="link" onClick={() => navigate('/register')}>
              创建新账户
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}

export default LoginPage
