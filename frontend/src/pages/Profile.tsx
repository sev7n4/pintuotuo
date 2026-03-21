import { useEffect, useState } from 'react'
import { Card, Form, Input, Button, Avatar, message, Descriptions, Tag, Space, Divider, Modal, Tabs, Row, Col, Statistic, Upload } from 'antd'
import { UserOutlined, MailOutlined, PhoneOutlined, EditOutlined, SafetyOutlined, TrophyOutlined, CameraOutlined, LoadingOutlined } from '@ant-design/icons'
import type { UploadProps } from 'antd'
import { useAuthStore } from '@/stores/authStore'
import { userService } from '@/services/user'
import styles from './Profile.module.css'

const Profile = () => {
  const { user, setUser } = useAuthStore()
  const [isEditing, setIsEditing] = useState(false)
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [form] = Form.useForm()
  const [passwordForm] = Form.useForm()
  const [loading, setLoading] = useState(false)
  const [avatarLoading, setAvatarLoading] = useState(false)

  useEffect(() => {
    if (user) {
      form.setFieldsValue({
        name: user.name,
        email: user.email,
      })
    }
  }, [user, form])

  const handleUpdateProfile = async () => {
    try {
      const values = await form.validateFields()
      setLoading(true)
      const response = await userService.updateCurrentUser(values)
      if (response.data?.data) {
        setUser(response.data.data)
        message.success('个人信息更新成功')
        setIsEditing(false)
      }
    } catch {
      message.error('更新失败')
    } finally {
      setLoading(false)
    }
  }

  const handleChangePassword = async () => {
    try {
      const values = await passwordForm.validateFields()
      if (values.newPassword !== values.confirmPassword) {
        message.error('两次输入的密码不一致')
        return
      }
      setLoading(true)
      message.success('密码修改成功')
      setPasswordModalVisible(false)
      passwordForm.resetFields()
    } catch {
      message.error('修改失败')
    } finally {
      setLoading(false)
    }
  }

  const handleAvatarUpload = async (file: File) => {
    const isImage = ['image/jpeg', 'image/png', 'image/gif'].includes(file.type)
    if (!isImage) {
      message.error('只能上传 JPG/PNG/GIF 格式的图片')
      return false
    }
    const isLt2M = file.size / 1024 / 1024 < 2
    if (!isLt2M) {
      message.error('图片大小不能超过 2MB')
      return false
    }

    setAvatarLoading(true)
    try {
      const response = await userService.uploadAvatar(file)
      if (response.data?.data?.url) {
        setUser({ ...user!, avatar_url: response.data.data.url })
        message.success('头像上传成功')
      }
    } catch {
      message.error('头像上传失败')
    } finally {
      setAvatarLoading(false)
    }
    return false
  }

  const uploadProps: UploadProps = {
    showUploadList: false,
    beforeUpload: (file) => {
      handleAvatarUpload(file)
      return false
    },
  }

  const getRoleTag = (role: string) => {
    const roleMap: Record<string, { color: string; text: string }> = {
      user: { color: 'blue', text: '普通用户' },
      merchant: { color: 'green', text: '商家' },
      admin: { color: 'red', text: '管理员' },
    }
    const { color, text } = roleMap[role] || { color: 'default', text: role }
    return <Tag color={color}>{text}</Tag>
  }

  const getUserLevel = (createdAt: string) => {
    const days = Math.floor((Date.now() - new Date(createdAt).getTime()) / (1000 * 60 * 60 * 24))
    if (days < 30) return { level: 1, name: '新用户', progress: days / 30 * 100 }
    if (days < 90) return { level: 2, name: '活跃用户', progress: (days - 30) / 60 * 100 }
    if (days < 180) return { level: 3, name: '忠诚用户', progress: (days - 90) / 90 * 100 }
    return { level: 4, name: '资深用户', progress: 100 }
  }

  const userLevel = user ? getUserLevel(user.created_at) : { level: 1, name: '新用户', progress: 0 }

  return (
    <div className={styles.profile}>
      <h2 className={styles.pageTitle}>个人中心</h2>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={8}>
          <Card className={styles.avatarCard}>
            <div className={styles.avatarSection}>
              <Upload {...uploadProps}>
                <div className={styles.avatarWrapper}>
                  <Avatar 
                    size={100} 
                    src={user?.avatar_url} 
                    icon={<UserOutlined />} 
                    className={styles.avatar} 
                  />
                  <div className={styles.avatarOverlay}>
                    {avatarLoading ? <LoadingOutlined /> : <CameraOutlined />}
                  </div>
                </div>
              </Upload>
              <h3 className={styles.userName}>{user?.name || '用户'}</h3>
              {user && getRoleTag(user.role)}
              <div className={styles.levelSection}>
                <TrophyOutlined className={styles.levelIcon} />
                <span className={styles.levelName}>{userLevel.name}</span>
                <span className={styles.levelBadge}>Lv.{userLevel.level}</span>
              </div>
            </div>
            <Divider />
            <div className={styles.statsSection}>
              <Row gutter={16}>
                <Col span={12}>
                  <Statistic title="注册天数" value={Math.floor((Date.now() - new Date(user?.created_at || Date.now()).getTime()) / (1000 * 60 * 60 * 24))} suffix="天" />
                </Col>
                <Col span={12}>
                  <Statistic title="用户等级" value={userLevel.level} prefix="Lv." />
                </Col>
              </Row>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={16}>
          <Card className={styles.infoCard}>
            <Tabs defaultActiveKey="basic">
              <Tabs.TabPane tab="基本信息" key="basic">
                {isEditing ? (
                  <Form form={form} layout="vertical" className={styles.form}>
                    <Form.Item
                      name="name"
                      label="用户名"
                      rules={[{ required: true, message: '请输入用户名' }]}
                    >
                      <Input prefix={<UserOutlined />} placeholder="请输入用户名" />
                    </Form.Item>
                    <Form.Item name="email" label="邮箱">
                      <Input prefix={<MailOutlined />} disabled />
                    </Form.Item>
                    <Form.Item name="phone" label="手机号">
                      <Input prefix={<PhoneOutlined />} placeholder="请输入手机号" />
                    </Form.Item>
                    <Form.Item name="address" label="地址">
                      <Input.TextArea placeholder="请输入地址" rows={2} />
                    </Form.Item>
                    <Form.Item>
                      <Space>
                        <Button type="primary" onClick={handleUpdateProfile} loading={loading}>
                          保存
                        </Button>
                        <Button onClick={() => setIsEditing(false)}>取消</Button>
                      </Space>
                    </Form.Item>
                  </Form>
                ) : (
                  <div className={styles.infoSection}>
                    <div className={styles.infoHeader}>
                      <h3>账户信息</h3>
                      <Button type="link" icon={<EditOutlined />} onClick={() => setIsEditing(true)}>
                        编辑
                      </Button>
                    </div>
                    <Descriptions column={{ xs: 1, sm: 2 }} bordered>
                      <Descriptions.Item label="用户名">{user?.name}</Descriptions.Item>
                      <Descriptions.Item label="邮箱">{user?.email}</Descriptions.Item>
                      <Descriptions.Item label="角色">
                        <span>{user && getRoleTag(user.role)}</span>
                      </Descriptions.Item>
                      <Descriptions.Item label="注册时间">
                        {user?.created_at ? new Date(user.created_at).toLocaleDateString('zh-CN') : '-'}
                      </Descriptions.Item>
                    </Descriptions>
                  </div>
                )}
              </Tabs.TabPane>

              <Tabs.TabPane tab="安全设置" key="security">
                <div className={styles.securitySection}>
                  <div className={styles.securityItem}>
                    <div className={styles.securityInfo}>
                      <SafetyOutlined className={styles.securityIcon} />
                      <div>
                        <h4>登录密码</h4>
                        <p className={styles.securityDesc}>定期更换密码可以提高账户安全性</p>
                      </div>
                    </div>
                    <Button type="link" onClick={() => setPasswordModalVisible(true)}>
                      修改密码
                    </Button>
                  </div>
                </div>
              </Tabs.TabPane>
            </Tabs>
          </Card>
        </Col>
      </Row>

      <Modal
        title="修改密码"
        open={passwordModalVisible}
        onOk={handleChangePassword}
        onCancel={() => {
          setPasswordModalVisible(false)
          passwordForm.resetFields()
        }}
        okText="确认修改"
        cancelText="取消"
        confirmLoading={loading}
      >
        <Form form={passwordForm} layout="vertical">
          <Form.Item
            name="oldPassword"
            label="原密码"
            rules={[{ required: true, message: '请输入原密码' }]}
          >
            <Input.Password placeholder="请输入原密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6位' },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认密码"
            rules={[{ required: true, message: '请确认新密码' }]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Profile
