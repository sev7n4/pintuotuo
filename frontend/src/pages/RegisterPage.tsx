import React from 'react'
import { Form, Input, Button, Card, message } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@stores/authStore'

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate()
  const { register, isLoading, error } = useAuthStore()
  const [form] = Form.useForm()

  const onFinish = async (values: {
    email: string
    name: string
    password: string
    confirmPassword: string
  }) => {
    if (values.password !== values.confirmPassword) {
      message.error('两次输入的密码不一致')
      return
    }

    try {
      await register(values.email, values.name, values.password)
      message.success('注册成功，正在重定向...')
      navigate('/products')
    } catch (err) {
      message.error(error || '注册失败，请稍后重试')
    }
  }

  return (
    <div className="auth-page">
      <Card className="auth-card" title="拼脱脱 - 注册" style={{ maxWidth: 400 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
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
            label="名字"
            name="name"
            rules={[
              { required: true, message: '请输入名字' },
              { min: 2, message: '名字至少2个字符' },
            ]}
          >
            <Input placeholder="输入你的名字" />
          </Form.Item>

          <Form.Item
            label="密码"
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' },
            ]}
          >
            <Input.Password placeholder="设置密码" />
          </Form.Item>

          <Form.Item
            label="确认密码"
            name="confirmPassword"
            rules={[{ required: true, message: '请确认密码' }]}
          >
            <Input.Password placeholder="再次输入密码" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={isLoading}>
              创建账户
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <span>已有账户？ </span>
            <Button type="link" onClick={() => navigate('/login')}>
              立即登录
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}

export default RegisterPage
