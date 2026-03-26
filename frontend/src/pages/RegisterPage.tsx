import React, { useState, useEffect } from 'react'
import { Form, Input, Button, Card, message, Radio, Space } from 'antd'
import { UserOutlined, ShopOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@stores/authStore'

type UserRole = 'user' | 'merchant'

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate()
  const { register, isLoading, error, user, isAuthenticated } = useAuthStore()
  const [form] = Form.useForm()
  const [selectedRole, setSelectedRole] = useState<UserRole>('user')

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

    const role = selectedRole || 'user'

    try {
      await register(values.email, values.name, values.password, role)
      message.success('注册成功')
    } catch (err) {
      message.error(error || '注册失败，请稍后重试')
    }
  }

  useEffect(() => {
    if (isAuthenticated && user) {
      if (user.role === 'merchant') {
        navigate('/merchant', { replace: true })
      } else {
        navigate('/', { replace: true })
      }
    }
  }, [isAuthenticated, user, navigate])

  return (
    <div className="auth-page">
      <Card className="auth-card" title="拼脱脱 - 注册" style={{ maxWidth: 400 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          autoComplete="off"
        >
          <Form.Item label="选择角色" name="role">
            <Radio.Group
              style={{ width: '100%' }}
              value={selectedRole}
              onChange={(e) => setSelectedRole(e.target.value)}
            >
              <Space direction="vertical" style={{ width: '100%' }}>
                <Radio.Button
                  value="user"
                  style={{ width: '100%', height: 60, display: 'flex', alignItems: 'center' }}
                >
                  <UserOutlined style={{ marginRight: 8 }} />
                  <span>
                    <strong>普通用户</strong>
                    <br />
                    <small>购买 Token、参与拼团</small>
                  </span>
                </Radio.Button>
                <Radio.Button
                  value="merchant"
                  style={{ width: '100%', height: 60, display: 'flex', alignItems: 'center' }}
                >
                  <ShopOutlined style={{ marginRight: 8 }} />
                  <span>
                    <strong>商家</strong>
                    <br />
                    <small>上架商品、管理订单</small>
                  </span>
                </Radio.Button>
              </Space>
            </Radio.Group>
          </Form.Item>

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
