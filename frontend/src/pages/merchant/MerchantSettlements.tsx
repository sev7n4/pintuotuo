import { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Modal, Descriptions, message, Statistic, Row, Col, Input, Form } from 'antd';
import {
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  WarningOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { MerchantSettlement } from '@/types';
import styles from './MerchantSettlements.module.css';

const getAuthToken = () => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token') || '';
};

const MerchantSettlements = () => {
  const { settlements, fetchSettlements, requestSettlement, isLoading } = useMerchantStore();
  const [detailVisible, setDetailVisible] = useState(false);
  const [disputeVisible, setDisputeVisible] = useState(false);
  const [selectedSettlement, setSelectedSettlement] = useState<MerchantSettlement | null>(null);
  const [disputeReason, setDisputeReason] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [disputeLoading, setDisputeLoading] = useState(false);

  useEffect(() => {
    fetchSettlements();
  }, [fetchSettlements]);

  const handleRequestSettlement = async () => {
    const success = await requestSettlement();
    if (success) {
      message.success('结算申请已提交');
      fetchSettlements();
    }
  };

  const handleViewDetail = (record: MerchantSettlement) => {
    setSelectedSettlement(record);
    setDetailVisible(true);
  };

  const handleConfirmSettlement = async () => {
    if (!selectedSettlement) return;
    
    setConfirmLoading(true);
    try {
      const response = await fetch(`/api/v1/merchant/settlements/${selectedSettlement.id}/confirm`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        message.success('结算已确认');
        setDetailVisible(false);
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '确认失败');
      }
    } catch (error) {
      message.error('确认失败，请重试');
    } finally {
      setConfirmLoading(false);
    }
  };

  const handleDisputeSettlement = async () => {
    if (!selectedSettlement || !disputeReason.trim()) {
      message.warning('请填写争议原因');
      return;
    }

    setDisputeLoading(true);
    try {
      const response = await fetch(`/api/v1/merchant/settlements/${selectedSettlement.id}/dispute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getAuthToken()}`,
        },
        body: JSON.stringify({ reason: disputeReason }),
      });

      if (response.ok) {
        message.success('争议已提交');
        setDisputeVisible(false);
        setDisputeReason('');
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '提交失败');
      }
    } catch (error) {
      message.error('提交失败，请重试');
    } finally {
      setDisputeLoading(false);
    }
  };

  const statusMap: Record<string, { color: string; text: string; icon: React.ReactNode }> = {
    pending: { color: 'default', text: '待处理', icon: <ClockCircleOutlined /> },
    processing: { color: 'processing', text: '处理中', icon: <SyncOutlined spin /> },
    completed: { color: 'success', text: '已完成', icon: <CheckCircleOutlined /> },
  };

  const columns = [
    {
      title: '结算ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '结算周期',
      key: 'period',
      render: (_: unknown, record: MerchantSettlement) => (
        <span>
          {new Date(record.period_start).toLocaleDateString('zh-CN')} -{' '}
          {new Date(record.period_end).toLocaleDateString('zh-CN')}
        </span>
      ),
    },
    {
      title: '销售总额',
      dataIndex: 'total_sales',
      key: 'total_sales',
      render: (amount: number) => `¥${amount.toFixed(2)}`,
    },
    {
      title: '平台费用',
      dataIndex: 'platform_fee',
      key: 'platform_fee',
      render: (fee: number) => `¥${fee.toFixed(2)}`,
    },
    {
      title: '结算金额',
      dataIndex: 'settlement_amount',
      key: 'settlement_amount',
      render: (amount: number) => <span className={styles.amount}>¥{amount.toFixed(2)}</span>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const { color, text, icon } = statusMap[status] || {
          color: 'default',
          text: status,
          icon: null,
        };
        return (
          <Tag color={color} icon={icon}>
            {text}
          </Tag>
        );
      },
    },
    {
      title: '确认状态',
      key: 'confirm_status',
      render: (_: unknown, record: MerchantSettlement) => (
        <Tag color={record.merchant_confirmed ? 'success' : 'default'}>
          {record.merchant_confirmed ? '已确认' : '待确认'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: MerchantSettlement) => (
        <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
          查看详情
        </Button>
      ),
    },
  ];

  const totalSettlements = settlements
    .filter((s) => s.status === 'completed')
    .reduce((sum, s) => sum + s.settlement_amount, 0);
  const pendingSettlements = settlements
    .filter((s) => s.status === 'pending' || s.status === 'processing')
    .reduce((sum, s) => sum + s.settlement_amount, 0);

  return (
    <div className={styles.settlements}>
      <h2 className={styles.pageTitle}>结算管理</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={12}>
          <Card>
            <Statistic
              title="累计结算"
              value={totalSettlements}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="元"
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12}>
          <Card>
            <Statistic
              title="待结算"
              value={pendingSettlements}
              precision={2}
              prefix={<ClockCircleOutlined />}
              suffix="元"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Card className={styles.tableCard}>
        <div className={styles.cardHeader}>
          <span>结算记录</span>
          <Button type="primary" onClick={handleRequestSettlement} loading={isLoading}>
            申请结算
          </Button>
        </div>
        <Table
          columns={columns}
          dataSource={settlements}
          rowKey="id"
          loading={isLoading}
          scroll={{ x: 1000 }}
          pagination={false}
        />
      </Card>

      <Modal
        title="结算详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedSettlement && (
          <>
            <Descriptions column={2} bordered>
              <Descriptions.Item label="结算ID">{selectedSettlement.id}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={statusMap[selectedSettlement.status]?.color}>
                  {statusMap[selectedSettlement.status]?.text}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="结算周期" span={2}>
                {new Date(selectedSettlement.period_start).toLocaleDateString('zh-CN')} -{' '}
                {new Date(selectedSettlement.period_end).toLocaleDateString('zh-CN')}
              </Descriptions.Item>
              <Descriptions.Item label="销售总额">
                ¥{selectedSettlement.total_sales.toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="平台费用（5%）">
                ¥{selectedSettlement.platform_fee.toFixed(2)}
              </Descriptions.Item>
              <Descriptions.Item label="结算金额" span={2}>
                <span className={styles.amount}>
                  ¥{selectedSettlement.settlement_amount.toFixed(2)}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label="商户确认">
                <Tag color={selectedSettlement.merchant_confirmed ? 'success' : 'default'}>
                  {selectedSettlement.merchant_confirmed ? '已确认' : '待确认'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="确认时间">
                {selectedSettlement.merchant_confirmed_at
                  ? new Date(selectedSettlement.merchant_confirmed_at).toLocaleString('zh-CN')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="财务审批">
                <Tag color={selectedSettlement.finance_approved ? 'success' : 'default'}>
                  {selectedSettlement.finance_approved ? '已审批' : '待审批'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="审批时间">
                {selectedSettlement.finance_approved_at
                  ? new Date(selectedSettlement.finance_approved_at).toLocaleString('zh-CN')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(selectedSettlement.created_at).toLocaleString('zh-CN')}
              </Descriptions.Item>
              <Descriptions.Item label="结算时间">
                {selectedSettlement.settled_at
                  ? new Date(selectedSettlement.settled_at).toLocaleString('zh-CN')
                  : '-'}
              </Descriptions.Item>
            </Descriptions>

            <div style={{ marginTop: 24, textAlign: 'right' }}>
              {!selectedSettlement.merchant_confirmed && selectedSettlement.status === 'pending' && (
                <>
                  <Button
                    type="primary"
                    onClick={handleConfirmSettlement}
                    loading={confirmLoading}
                    style={{ marginRight: 8 }}
                  >
                    确认结算
                  </Button>
                  <Button
                    danger
                    icon={<WarningOutlined />}
                    onClick={() => {
                      setDetailVisible(false);
                      setDisputeVisible(true);
                    }}
                  >
                    提交争议
                  </Button>
                </>
              )}
            </div>
          </>
        )}
      </Modal>

      <Modal
        title="提交结算争议"
        open={disputeVisible}
        onCancel={() => {
          setDisputeVisible(false);
          setDisputeReason('');
        }}
        onOk={handleDisputeSettlement}
        confirmLoading={disputeLoading}
        okText="提交"
        cancelText="取消"
      >
        <Form layout="vertical">
          <Form.Item label="争议原因" required>
            <Input.TextArea
              rows={4}
              value={disputeReason}
              onChange={(e) => setDisputeReason(e.target.value)}
              placeholder="请详细说明争议原因..."
              maxLength={500}
              showCount
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default MerchantSettlements;
