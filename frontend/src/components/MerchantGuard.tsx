import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Modal, Button, Tag, Space, Typography } from 'antd';
import {
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import { merchantService } from '@/services/merchant';

const { Text, Paragraph } = Typography;

interface MerchantGuardProps {
  children: React.ReactNode;
  requiredStatus?: 'active';
}

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

export default function MerchantGuard({ children, requiredStatus = 'active' }: MerchantGuardProps) {
  const navigate = useNavigate();
  const [status, setStatus] = useState<MerchantStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);

  useEffect(() => {
    checkMerchantStatus();
  }, []);

  const checkMerchantStatus = async () => {
    try {
      const response = await merchantService.getMerchantStatus();
      setStatus(response.data);

      if (response.data.status !== requiredStatus && response.data.status !== 'not_registered') {
        setModalVisible(true);
      }
    } catch {
      // User might not be a merchant yet
      setStatus({ status: 'not_registered', can_submit: false });
    } finally {
      setLoading(false);
    }
  };

  const handleNavigateToSettings = () => {
    setModalVisible(false);
    navigate('/merchant/settings');
  };

  if (loading) {
    return null;
  }

  if (!status || status.status === 'not_registered') {
    return <>{children}</>;
  }

  if (status.status !== requiredStatus) {
    const config = statusConfig[status.status] || statusConfig.pending;

    return (
      <>
        <Modal
          open={modalVisible}
          title={
            <Space>
              {config.icon}
              <span>店铺状态：{config.label}</span>
            </Space>
          }
          closable={false}
          footer={[
            <Button key="later" onClick={() => setModalVisible(false)}>
              稍后再说
            </Button>,
            status.can_submit && (
              <Button key="submit" type="primary" onClick={handleNavigateToSettings}>
                去提交资料
              </Button>
            ),
          ]}
        >
          <Space direction="vertical" style={{ width: '100%' }}>
            <Tag color={config.color}>{config.label}</Tag>

            {status.status === 'pending' && (
              <Paragraph>
                您的店铺尚未提交审核资料，请先提交营业执照、身份证等信息完成审核后才能使用此功能。
              </Paragraph>
            )}

            {status.status === 'reviewing' && (
              <Paragraph>
                您的店铺资料正在审核中，审核通过后即可使用此功能。请耐心等待，我们会在1-3个工作日内完成审核。
              </Paragraph>
            )}

            {status.status === 'rejected' && (
              <>
                <Paragraph type="danger">
                  您的店铺审核未通过，请根据以下原因修改后重新提交：
                </Paragraph>
                {status.rejection_reason && (
                  <Paragraph
                    type="warning"
                    style={{ background: '#fffbe6', padding: 12, borderRadius: 4 }}
                  >
                    拒绝原因：{status.rejection_reason}
                  </Paragraph>
                )}
              </>
            )}

            {status.status === 'suspended' && (
              <Paragraph type="warning">您的店铺已被暂停，如有疑问请联系客服。</Paragraph>
            )}
          </Space>
        </Modal>
        {children}
      </>
    );
  }

  return <>{children}</>;
}
