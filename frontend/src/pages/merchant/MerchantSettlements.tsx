import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Tag,
  Button,
  Modal,
  Descriptions,
  message,
  Statistic,
  Row,
  Col,
  Input,
  Form,
  Space,
  Pagination,
} from 'antd';
import dayjs from 'dayjs';
import {
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  WarningOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import { useMerchantStore } from '@/stores/merchantStore';
import { MerchantSettlement, BillingRecord } from '@/types';
import styles from './MerchantSettlements.module.css';

const getAuthToken = () => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token') || '';
};

const MerchantSettlements = () => {
  const { settlements, fetchSettlements, requestSettlement, isLoading } = useMerchantStore();
  const [detailVisible, setDetailVisible] = useState(false);
  const [disputeVisible, setDisputeVisible] = useState(false);
  const [billingVisible, setBillingVisible] = useState(false);
  const [selectedSettlement, setSelectedSettlement] = useState<MerchantSettlement | null>(null);
  const [disputeReason, setDisputeReason] = useState('');
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [disputeLoading, setDisputeLoading] = useState(false);
  const [billingRecords, setBillingRecords] = useState<BillingRecord[]>([]);
  const [billingLoading, setBillingLoading] = useState(false);
  const [billingPage, setBillingPage] = useState(1);
  const [billingPageSize, setBillingPageSize] = useState(20);
  const [billingTotal, setBillingTotal] = useState(0);

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

  const handleViewBilling = async (settlement: MerchantSettlement) => {
    setSelectedSettlement(settlement);
    setBillingVisible(true);
    await fetchBillingRecords(settlement);
  };

  const fetchBillingRecords = async (settlement: MerchantSettlement) => {
    setBillingLoading(true);
    try {
      const params = new URLSearchParams();
      params.append('page', billingPage.toString());
      params.append('page_size', billingPageSize.toString());
      params.append('start_date', dayjs(settlement.period_start).format('YYYY-MM-DD'));
      params.append('end_date', dayjs(settlement.period_end).format('YYYY-MM-DD'));

      const response = await fetch(`/api/v1/merchant/billings?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setBillingRecords(data.billings || []);
        setBillingTotal(data.total || 0);
      } else {
        message.error('获取账单明细失败');
      }
    } catch (error) {
      message.error('获取账单明细失败');
    } finally {
      setBillingLoading(false);
    }
  };

  const handleConfirmSettlement = async () => {
    if (!selectedSettlement) return;

    setConfirmLoading(true);
    try {
      const response = await fetch(
        `/api/v1/merchant/settlements/${selectedSettlement.id}/confirm`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${getAuthToken()}`,
          },
        }
      );

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
      const response = await fetch(
        `/api/v1/merchant/settlements/${selectedSettlement.id}/dispute`,
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${getAuthToken()}`,
          },
          body: JSON.stringify({ reason: disputeReason }),
        }
      );

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

  const handleExportBilling = () => {
    message.info('导出功能开发中...');
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
      title: '销售总额(¥)',
      dataIndex: 'total_sales_cny',
      key: 'total_sales_cny',
      render: (amount: number) => `¥${amount.toFixed(6)}`,
    },
    {
      title: 'Token使用量',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      render: (tokens: number) => tokens?.toLocaleString() || '0',
    },
    {
      title: '平台费用',
      dataIndex: 'platform_fee',
      key: 'platform_fee',
      render: (fee: number) => `¥${fee.toFixed(6)}`,
    },
    {
      title: '结算金额',
      dataIndex: 'settlement_amount',
      key: 'settlement_amount',
      render: (amount: number) => <span className={styles.amount}>¥{amount.toFixed(6)}</span>,
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
      width: 150,
      render: (_: unknown, record: MerchantSettlement) => (
        <Space>
          <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
            详情
          </Button>
          <Button type="link" size="small" onClick={() => handleViewBilling(record)}>
            账单明细
          </Button>
        </Space>
      ),
    },
  ];

  const billingColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'Provider',
      dataIndex: 'provider',
      key: 'provider',
      width: 100,
    },
    {
      title: 'Model',
      dataIndex: 'model',
      key: 'model',
      width: 150,
    },
    {
      title: '输入Tokens',
      dataIndex: 'input_tokens',
      key: 'input_tokens',
      width: 100,
      render: (num: number) => num.toLocaleString(),
    },
    {
      title: '输出Tokens',
      dataIndex: 'output_tokens',
      key: 'output_tokens',
      width: 100,
      render: (num: number) => num.toLocaleString(),
    },
    {
      title: '成本',
      dataIndex: 'cost',
      key: 'cost',
      width: 100,
      render: (cost: number) => `¥${cost.toFixed(6)}`,
    },
    {
      title: '延迟(ms)',
      dataIndex: 'latency_ms',
      key: 'latency_ms',
      width: 100,
      render: (time: number) => time.toFixed(2),
    },
    {
      title: '状态',
      dataIndex: 'status_code',
      key: 'status_code',
      width: 80,
      render: (code: number) => <Tag color={code === 200 ? 'success' : 'error'}>{code}</Tag>,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
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
          scroll={{ x: 1200 }}
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
              <Descriptions.Item label="销售总额(¥)">
                ¥{selectedSettlement.total_sales_cny?.toFixed(6) || selectedSettlement.total_sales?.toFixed(6) || '0.000000'}
              </Descriptions.Item>
              <Descriptions.Item label="Token使用量">
                {selectedSettlement.total_tokens?.toLocaleString() || '0'}
              </Descriptions.Item>
              <Descriptions.Item label="平台费用（5%）">
                ¥{selectedSettlement.platform_fee.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="结算金额" span={2}>
                <span className={styles.amount}>
                  ¥{selectedSettlement.settlement_amount.toFixed(6)}
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
              {!selectedSettlement.merchant_confirmed &&
                selectedSettlement.status === 'pending' && (
                  <>
                    <Button
                      style={{ marginRight: 8 }}
                      onClick={() => {
                        setDetailVisible(false);
                        handleViewBilling(selectedSettlement);
                      }}
                    >
                      查看账单明细
                    </Button>
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
        title="账单明细"
        open={billingVisible}
        onCancel={() => {
          setBillingVisible(false);
          setBillingRecords([]);
          setBillingTotal(0);
        }}
        footer={null}
        width={1200}
      >
        {selectedSettlement && (
          <>
            <div style={{ marginBottom: 16 }}>
              <Space>
                <span>
                  结算周期: {dayjs(selectedSettlement.period_start).format('YYYY-MM-DD')} -{' '}
                  {dayjs(selectedSettlement.period_end).format('YYYY-MM-DD')}
                </span>
                <Button icon={<DownloadOutlined />} onClick={handleExportBilling}>
                  导出
                </Button>
              </Space>
            </div>

            <Table
              columns={billingColumns}
              dataSource={billingRecords}
              rowKey="id"
              loading={billingLoading}
              scroll={{ x: 1000 }}
              pagination={false}
            />

            <div style={{ marginTop: 16, textAlign: 'right' }}>
              <Pagination
                current={billingPage}
                pageSize={billingPageSize}
                total={billingTotal}
                showSizeChanger
                showQuickJumper
                showTotal={(total) => `共 ${total} 条记录`}
                onChange={(page, pageSize) => {
                  setBillingPage(page);
                  setBillingPageSize(pageSize);
                  if (selectedSettlement) {
                    fetchBillingRecords(selectedSettlement);
                  }
                }}
              />
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
