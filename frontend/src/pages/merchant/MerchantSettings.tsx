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
  Divider,
  Alert,
  Spin,
  Image,
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
import { uploadService } from '@/services/upload';
import type { UploadProps, UploadFile } from 'antd';
import styles from './MerchantSettings.module.css';

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
  const [uploadingLogo, setUploadingLogo] = useState(false);
  const [logoUrl, setLogoUrl] = useState<string>('');
  const [businessLicenseFileList, setBusinessLicenseFileList] = useState<UploadFile[]>([]);
  const [idCardFrontFileList, setIdCardFrontFileList] = useState<UploadFile[]>([]);
  const [idCardBackFileList, setIdCardBackFileList] = useState<UploadFile[]>([]);

  useEffect(() => {
    fetchProfile();
    fetchMerchantStatus();
  }, [fetchProfile]);

  useEffect(() => {
    if (profile) {
      form.setFieldsValue(profile);
      documentForm.setFieldsValue({
        contact_name: profile.contact_name,
        contact_phone: profile.contact_phone,
        contact_email: profile.contact_email,
        address: profile.address,
      });
      setLogoUrl(profile.logo_url || '');
      if (profile.business_license_url) {
        setBusinessLicenseFileList([
          { uid: '-1', name: '营业执照', status: 'done', url: profile.business_license_url },
        ]);
      }
      if (profile.id_card_front_url) {
        setIdCardFrontFileList([
          { uid: '-2', name: '身份证正面', status: 'done', url: profile.id_card_front_url },
        ]);
      }
      if (profile.id_card_back_url) {
        setIdCardBackFileList([
          { uid: '-3', name: '身份证背面', status: 'done', url: profile.id_card_back_url },
        ]);
      }
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
      if (logoUrl) {
        values.logo_url = logoUrl;
      }
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

      if (businessLicenseFileList.length === 0 || !businessLicenseFileList[0].url) {
        message.error('请上传营业执照');
        return;
      }

      setSubmitting(true);
      const submitData = {
        ...values,
        business_license_url: businessLicenseFileList[0].url || businessLicenseFileList[0].response?.url,
        id_card_front_url: idCardFrontFileList[0]?.url || idCardFrontFileList[0]?.response?.url,
        id_card_back_url: idCardBackFileList[0]?.url || idCardBackFileList[0]?.response?.url,
      };

      await merchantService.submitDocuments(submitData);
      message.success('资料提交成功，请等待审核');
      fetchProfile();
      fetchMerchantStatus();
    } catch {
      message.error('提交失败，请重试');
    } finally {
      setSubmitting(false);
    }
  };

  const handleLogoUpload = async (options: any) => {
    const { file, onSuccess, onError } = options;
    setUploadingLogo(true);
    try {
      const url = await uploadService.uploadFile(file, 'logo');
      setLogoUrl(url);
      onSuccess({ url });
      message.success('Logo上传成功');
    } catch (error) {
      onError(error);
      message.error('Logo上传失败');
    } finally {
      setUploadingLogo(false);
    }
  };

  const createUploadProps = (
    type: 'license' | 'idcard',
    fileList: UploadFile[],
    setFileList: React.Dispatch<React.SetStateAction<UploadFile[]>>
  ): UploadProps => ({
    name: 'file',
    listType: 'picture-card',
    fileList,
    accept: '.jpg,.jpeg,.png,.gif,.webp',
    maxCount: 1,
    beforeUpload: (file) => {
      const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png' || file.type === 'image/gif' || file.type === 'image/webp';
      if (!isJpgOrPng) {
        message.error('只能上传 JPG/PNG/GIF/WEBP 格式的图片');
        return false;
      }
      const isLt5M = file.size / 1024 / 1024 < 5;
      if (!isLt5M) {
        message.error('图片大小不能超过 5MB');
        return false;
      }
      return true;
    },
    customRequest: async (options) => {
      const { file, onSuccess, onError } = options;
      try {
        const url = await uploadService.uploadFile(file as File, type);
        onSuccess?.({ url });
      } catch (error) {
        onError?.(error as Error);
      }
    },
    onChange: (info) => {
      const updatedFileList = info.fileList.map(file => {
        if (file.status === 'done' && file.response?.url) {
          return { ...file, url: file.response.url };
        }
        return file;
      });
      setFileList(updatedFileList);
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 上传成功`);
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 上传失败`);
      }
    },
    onRemove: () => {
      setFileList([]);
    },
  });

  const logoUploadProps: UploadProps = {
    name: 'file',
    showUploadList: false,
    accept: '.jpg,.jpeg,.png,.gif,.webp',
    customRequest: handleLogoUpload,
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
              src={logoUrl || profile?.logo_url}
              className={styles.logo}
            />
            <Upload {...logoUploadProps}>
              <Button icon={<UploadOutlined />} loading={uploadingLogo}>
                更换Logo
              </Button>
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
                <Image
                  src={profile.business_license_url}
                  alt="营业执照"
                  width={100}
                  height={100}
                  style={{ objectFit: 'cover', borderRadius: 4 }}
                />
              </div>
            )}
            {profile?.id_card_front_url && (
              <div className={styles.statusItem}>
                <span className={styles.label}>身份证正面：</span>
                <Image
                  src={profile.id_card_front_url}
                  alt="身份证正面"
                  width={100}
                  height={100}
                  style={{ objectFit: 'cover', borderRadius: 4 }}
                />
              </div>
            )}
            {profile?.id_card_back_url && (
              <div className={styles.statusItem}>
                <span className={styles.label}>身份证背面：</span>
                <Image
                  src={profile.id_card_back_url}
                  alt="身份证背面"
                  width={100}
                  height={100}
                  style={{ objectFit: 'cover', borderRadius: 4 }}
                />
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
              label="营业执照"
              required
              tooltip="请上传清晰的营业执照照片"
            >
              <Upload {...createUploadProps('license', businessLicenseFileList, setBusinessLicenseFileList)}>
                {businessLicenseFileList.length === 0 && (
                  <div>
                    <UploadOutlined />
                    <div style={{ marginTop: 8 }}>上传营业执照</div>
                  </div>
                )}
              </Upload>
              <span className="ant-form-item-extra" style={{ color: '#999', fontSize: 12 }}>
                支持 JPG、PNG、GIF、WEBP 格式，文件大小不超过 5MB
              </span>
            </Form.Item>

            <Form.Item label="身份证正面">
              <Upload {...createUploadProps('idcard', idCardFrontFileList, setIdCardFrontFileList)}>
                {idCardFrontFileList.length === 0 && (
                  <div>
                    <UploadOutlined />
                    <div style={{ marginTop: 8 }}>上传身份证正面</div>
                  </div>
                )}
              </Upload>
            </Form.Item>

            <Form.Item label="身份证背面">
              <Upload {...createUploadProps('idcard', idCardBackFileList, setIdCardBackFileList)}>
                {idCardBackFileList.length === 0 && (
                  <div>
                    <UploadOutlined />
                    <div style={{ marginTop: 8 }}>上传身份证背面</div>
                  </div>
                )}
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
