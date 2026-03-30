import { useEffect, useState } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  message,
  Upload,
  Avatar,
  Space,
  Tag,
  Typography,
  Divider,
  Alert,
  Spin,
} from 'antd';
import {
  UserOutlined,
  UploadOutlined,
  SaveOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { merchantService } from '@/services/merchant';
import type { UploadProps } from 'antd';
import styles from './MerchantSettings.module.css';

const { Text, Paragraph } = Typography;

interface MerchantStatus {
  status: string;
  can_submit: boolean;
  rejection_reason?: string;
}

const statusConfig: Record<string, { label: string; color: string; icon: React.ReactNode }> = {
  pending: { label: '未审核', color: 'default', icon: <ExclamationCircleOutlined /> },
  reviewing: { label: '审核中', color: 'processing', icon: <ClockCircleOutlined /> },
  active: { label: '已审核', color: 'success', icon: <CheckCircleOutlined /> },
  rejected: { label: '已拒绝', color: 'error', icon: <CloseCircleOutlined /> },
  suspended: { label: '已暂停', color: 'warning', icon: <ExclamationCircleOutlined /> },
};

const MerchantSettings = () => {
  const { profile, fetchProfile, updateProfile, isLoading } = useMerchantStore();
  const [form] = Form.useForm();
  const [documentForm] = Form.useForm();
  const [merchantStatus, setMerchantStatus] = useState<MerchantStatus | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchProfile();
    fetchMerchantStatus();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (profile) {
      form.setFieldsValue(profile);
      documentForm.setFieldsValue({
        contact_name: profile.contact_name,
        contact_phone: profile.contact_phone,
        contact_email: profile.contact_email,
        address: profile.address,
      });
    }
  }, [profile, form, documentForm]);

  const fetchMerchantStatus = async () => {
    try {
      const response = await merchantService.getMerchantStatus();
      setMerchantStatus(response.data);
    } catch {
      // Ignore error
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const success = await updateProfile(values);
      if (success) {
        message.success('店铺信息已更新');
      }
    } catch {
      message.error('更新失败，请检查输入');
    }
  };

  const handleDocumentSubmit = async () => {
    try {
      const values = await documentForm.validateFields();
      if (!values.business_license_url) {
        message.error('请上传营业执照');
        return;
      }

      setSubmitting(true);
      await merchantService.submitDocuments(values);
      message.success('资料提交成功，请等待审核');
      fetchProfile();
      fetchMerchantStatus();
    } catch {
      message.error('提交失败，请重试');
    } finally {
      setSubmitting(false);
    }
  };

  const uploadProps: UploadProps = {
    name: 'file',
    showUploadList: false,
    beforeUpload: () => {
      message.info('文件上传功能开发中');
      return false;
    },
  };

  const currentStatus = merchantStatus?.status || profile?.status || 'pending';
  const statusInfo = statusConfig[currentStatus] || statusConfig.pending;

  return (
    <div className={styles.settings}>
      <h2 className={styles.pageTitle}>店铺设置</h2>

      <Card className={styles.card}>
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
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
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={isLoading}>
                保存设置
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      <Card title="认证状态" className={styles.card}>
        <Spin spinning={!merchantStatus}>
          <div className={styles.statusSection}>
            <div className={styles.statusItem}>
              <span className={styles.label}>认证状态：</span>
              <Tag color={statusInfo.color} icon={statusInfo.icon}>
                {statusInfo.label}
              </Tag>
            </div>

            {merchantStatus?.rejection_reason && (
              <Alert
                type="error"
                message="审核未通过"
                description={merchantStatus.rejection_reason}
                showIcon
                style={{ marginBottom: 16 }}
              />
            )}

            {profile?.business_license_url && (
              <div className={styles.statusItem}>
                <span className={styles.label}>营业执照：</span>
                <span className={styles.value}>已上传</span>
              </div>
            )}
            {profile?.id_card_front_url && (
              <div className={styles.statusItem}>
                <span className={styles.label}>身份证正面：</span>
                <span className={styles.value}>已上传</span>
              </div>
            )}
            {profile?.id_card_back_url && (
              <div className={styles.statusItem}>
                <span className={styles.label}>身份证背面：</span>
                <span className={styles.value}>已上传</span>
              </div>
            )}
          </div>
        </Spin>
      </Card>

      {(currentStatus === 'pending' || currentStatus === 'rejected') && (
        <Card title="提交审核资料" className={styles.card}>
          <Alert
            type="info"
            message="提交审核资料后，我们将在1-3个工作日内完成审核。审核通过后即可使用商品管理和SKU管理功能。"
            showIcon
            style={{ marginBottom: 24 }}
          />

          <Form form={documentForm} layout="vertical" onFinish={handleDocumentSubmit}>
            <Form.Item
              name="business_license_url"
              label="营业执照"
              rules={[{ required: true, message: '请上传营业执照' }]}
            >
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传营业执照</Button>
              </Upload>
              <Text type="secondary" style={{ marginLeft: 8 }}>
                支持 JPG、PNG 格式，文件大小不超过 5MB
              </Text>
            </Form.Item>

            <Form.Item name="id_card_front_url" label="身份证正面">
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传身份证正面</Button>
              </Upload>
            </Form.Item>

            <Form.Item name="id_card_back_url" label="身份证背面">
              <Upload {...uploadProps}>
                <Button icon={<UploadOutlined />}>上传身份证背面</Button>
              </Upload>
            </Form.Item>

            <Divider />

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

            <Form.Item>
              <Button type="primary" htmlType="submit" loading={submitting}>
                提交审核
              </Button>
            </Form.Item>
          </Form>
        </Card>
      )}

      {currentStatus === 'reviewing' && (
        <Card className={styles.card}>
          <Alert
            type="info"
            message="审核中"
            description="您的资料正在审核中，请耐心等待。我们会在1-3个工作日内完成审核。"
            showIcon
          />
        </Card>
      )}
    </div>
  );
};

export default MerchantSettings;
