import { useEffect } from 'react'
import { Card, Form, Input, Button, message, Upload, Avatar, Space } from 'antd'
import { UserOutlined, UploadOutlined, SaveOutlined } from '@ant-design/icons'
import { useMerchantStore } from '@/stores/merchantStore'
import type { UploadProps } from 'antd'
import styles from './MerchantSettings.module.css'

const MerchantSettings = () => {
  const { profile, fetchProfile, updateProfile, isLoading } = useMerchantStore()
  const [form] = Form.useForm()

  useEffect(() => {
    fetchProfile()
  }, [fetchProfile])

  useEffect(() => {
    if (profile) {
      form.setFieldsValue(profile)
    }
  }, [profile, form])

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields()
      const success = await updateProfile(values)
      if (success) {
        message.success('店铺信息已更新')
      }
    } catch (error) {
      message.error('更新失败，请检查输入')
    }
  }

  const uploadProps: UploadProps = {
    name: 'logo',
    showUploadList: false,
    beforeUpload: () => {
      message.info('Logo上传功能开发中')
      return false
    },
  }

  return (
    <div className={styles.settings}>
      <h2 className={styles.pageTitle}>店铺设置</h2>

      <Card className={styles.card}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
        >
          <div className={styles.logoSection}>
            <Avatar
              size={80}
              icon={<UserOutlined />}
              src={profile?.logo_url}
              className={styles.logo}
            />
            <Upload {...uploadProps}>
              <Button icon={<UploadOutlined />}>更换Logo</Button>
            </Upload>
          </div>

          <Form.Item
            name="company_name"
            label="公司名称"
            rules={[{ required: true, message: '请输入公司名称' }]}
          >
            <Input placeholder="请输入公司名称" />
          </Form.Item>

          <Form.Item name="contact_name" label="联系人">
            <Input placeholder="请输入联系人姓名" />
          </Form.Item>

          <Form.Item name="contact_phone" label="联系电话">
            <Input placeholder="请输入联系电话" />
          </Form.Item>

          <Form.Item name="contact_email" label="联系邮箱">
            <Input type="email" placeholder="请输入联系邮箱" />
          </Form.Item>

          <Form.Item name="address" label="公司地址">
            <Input.TextArea rows={2} placeholder="请输入公司地址" />
          </Form.Item>

          <Form.Item name="description" label="店铺简介">
            <Input.TextArea rows={4} placeholder="请输入店铺简介" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<SaveOutlined />}
                loading={isLoading}
              >
                保存设置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card title="认证状态" className={styles.card}>
        <div className={styles.statusSection}>
          <div className={styles.statusItem}>
            <span className={styles.label}>认证状态：</span>
            <span className={styles.value}>
              {profile?.status === 'active' && '已认证'}
              {profile?.status === 'pending' && '审核中'}
              {profile?.status === 'suspended' && '已暂停'}
              {profile?.status === 'rejected' && '认证失败'}
            </span>
          </div>
          {profile?.business_license && (
            <div className={styles.statusItem}>
              <span className={styles.label}>营业执照：</span>
              <span className={styles.value}>已上传</span>
            </div>
          )}
          {profile?.verified_at && (
            <div className={styles.statusItem}>
              <span className={styles.label}>认证时间：</span>
              <span className={styles.value}>
                {new Date(profile.verified_at).toLocaleDateString('zh-CN')}
              </span>
            </div>
          )}
        </div>
      </Card>
    </div>
  )
}

export default MerchantSettings
