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
  Form,
  Select,
  DatePicker,
  InputNumber,
  Space,
} from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import {
  DollarOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  AuditOutlined,
  BankOutlined,
  EyeOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { MerchantSettlement, SettlementItem } from '@/types';
import styles from './AdminSettlements.module.css';

const getAuthToken = () => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token') || '';
};

const AdminSettlements = () => {
  const [settlements, setSettlements] = useState<MerchantSettlement[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedSettlement, setSelectedSettlement] = useState<MerchantSettlement | null>(null);
  const [approveLoading, setApproveLoading] = useState(false);
  const [markPaidLoading, setMarkPaidLoading] = useState(false);
  const [generateVisible, setGenerateVisible] = useState(false);
  const [generateLoading, setGenerateLoading] = useState(false);
  const [filterStatus, setFilterStatus] = useState<string>('');
  const [filterYearMonth, setFilterYearMonth] = useState<Dayjs | null>(null);
  const [generateYearMonth, setGenerateYearMonth] = useState<Dayjs | null>(null);
  const [itemsVisible, setItemsVisible] = useState(false);
  const [settlementItems, setSettlementItems] = useState<SettlementItem[]>([]);
  const [itemsLoading, setItemsLoading] = useState(false);
  const [merchants, setMerchants] = useState<Array<{ id: number; company_name: string }>>([]);
  const [generateMerchantVisible, setGenerateMerchantVisible] = useState(false);
  const [generateMerchantLoading, setGenerateMerchantLoading] = useState(false);
  const [selectedMerchantId, setSelectedMerchantId] = useState<number | null>(null);
  const [generateMerchantYearMonth, setGenerateMerchantYearMonth] = useState<Dayjs | null>(null);
  const [filterMerchantId, setFilterMerchantId] = useState<number | null>(null);
  const [filterMerchantConfirmed, setFilterMerchantConfirmed] = useState<string>('');
  const [filterFinanceApproved, setFilterFinanceApproved] = useState<string>('');
  const [filterMinAmount, setFilterMinAmount] = useState<number | null>(null);
  const [filterMaxAmount, setFilterMaxAmount] = useState<number | null>(null);

  useEffect(() => {
    fetchSettlements();
    fetchMerchants();
  }, [
    filterStatus,
    filterYearMonth,
    filterMerchantId,
    filterMerchantConfirmed,
    filterFinanceApproved,
    filterMinAmount,
    filterMaxAmount,
  ]);

  const fetchSettlements = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (filterStatus) params.append('status', filterStatus);
      if (filterYearMonth) {
        params.append('year', filterYearMonth.year().toString());
        params.append('month', (filterYearMonth.month() + 1).toString());
      }
      if (filterMerchantId) params.append('merchant_id', filterMerchantId.toString());
      if (filterMerchantConfirmed) params.append('merchant_confirmed', filterMerchantConfirmed);
      if (filterFinanceApproved) params.append('finance_approved', filterFinanceApproved);
      if (filterMinAmount !== null) params.append('min_amount', filterMinAmount.toString());
      if (filterMaxAmount !== null) params.append('max_amount', filterMaxAmount.toString());

      const response = await fetch(`/api/v1/admin/settlements?${params.toString()}`, {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSettlements(data.settlements || []);
      } else {
        message.error('获取结算列表失败');
      }
    } catch (error) {
      message.error('获取结算列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleResetFilters = () => {
    setFilterStatus('');
    setFilterYearMonth(null);
    setFilterMerchantId(null);
    setFilterMerchantConfirmed('');
    setFilterFinanceApproved('');
    setFilterMinAmount(null);
    setFilterMaxAmount(null);
  };

  const fetchMerchants = async () => {
    try {
      const response = await fetch('/api/v1/admin/merchants', {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setMerchants(data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch merchants:', error);
    }
  };

  const handleViewDetail = async (record: MerchantSettlement) => {
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/settlements/${record.id}`, {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSelectedSettlement(data.settlement);
        setDetailVisible(true);
      } else {
        message.error('获取结算详情失败');
      }
    } catch (error) {
      message.error('获取结算详情失败');
    } finally {
      setLoading(false);
    }
  };

  const handleApproveSettlement = async () => {
    if (!selectedSettlement) return;

    setApproveLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/settlements/${selectedSettlement.id}/approve`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        message.success('审批成功');
        setDetailVisible(false);
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '审批失败');
      }
    } catch (error) {
      message.error('审批失败，请重试');
    } finally {
      setApproveLoading(false);
    }
  };

  const handleMarkAsPaid = async () => {
    if (!selectedSettlement) return;

    setMarkPaidLoading(true);
    try {
      const response = await fetch(`/api/v1/admin/settlements/${selectedSettlement.id}/mark-paid`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        message.success('已标记为打款');
        setDetailVisible(false);
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '标记失败');
      }
    } catch (error) {
      message.error('标记失败，请重试');
    } finally {
      setMarkPaidLoading(false);
    }
  };

  const handleGenerateSettlements = async (values: { year: number; month: number }) => {
    setGenerateLoading(true);
    try {
      const response = await fetch('/api/v1/admin/settlements/generate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${getAuthToken()}`,
        },
        body: JSON.stringify(values),
      });

      if (response.ok) {
        const data = await response.json();
        message.success(`成功生成 ${data.count} 个结算单`);
        setGenerateVisible(false);
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '生成失败');
      }
    } catch (error) {
      message.error('生成失败，请重试');
    } finally {
      setGenerateLoading(false);
    }
  };

  const handleGenerateForMerchant = async () => {
    if (!selectedMerchantId || !generateMerchantYearMonth) {
      message.warning('请选择商户和年月');
      return;
    }

    setGenerateMerchantLoading(true);
    try {
      const response = await fetch('/api/v1/admin/settlements/generate/merchant', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${getAuthToken()}`,
        },
        body: JSON.stringify({
          merchant_id: selectedMerchantId,
          year: generateMerchantYearMonth.year(),
          month: generateMerchantYearMonth.month() + 1,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        message.success(`成功生成结算单: ${data.settlement.id}`);
        setGenerateMerchantVisible(false);
        setSelectedMerchantId(null);
        setGenerateMerchantYearMonth(null);
        fetchSettlements();
      } else {
        const error = await response.json();
        message.error(error.error || '生成失败');
      }
    } catch (error) {
      message.error('生成失败，请重试');
    } finally {
      setGenerateMerchantLoading(false);
    }
  };

  const handleViewItems = async (record: MerchantSettlement) => {
    setItemsLoading(true);
    setItemsVisible(true);
    try {
      const response = await fetch(`/api/v1/admin/settlements/${record.id}/items`, {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSettlementItems(data.items || []);
      } else {
        message.error('获取结算明细失败');
      }
    } catch (error) {
      message.error('获取结算明细失败');
    } finally {
      setItemsLoading(false);
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
      title: '商户ID',
      dataIndex: 'merchant_id',
      key: 'merchant_id',
      width: 100,
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
      key: 'total_sales_cny',
      render: (_: unknown, record: MerchantSettlement) => {
        const v = record.total_sales_cny ?? record.total_sales ?? 0;
        return `¥${Number(v).toFixed(6)}`;
      },
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
      title: '商户确认',
      key: 'merchant_confirm',
      render: (_: unknown, record: MerchantSettlement) => (
        <Tag color={record.merchant_confirmed ? 'success' : 'default'}>
          {record.merchant_confirmed ? '已确认' : '待确认'}
        </Tag>
      ),
    },
    {
      title: '财务审批',
      key: 'finance_approve',
      render: (_: unknown, record: MerchantSettlement) => (
        <Tag color={record.finance_approved ? 'success' : 'default'}>
          {record.finance_approved ? '已审批' : '待审批'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: MerchantSettlement) => (
        <>
          <Button type="link" size="small" onClick={() => handleViewDetail(record)}>
            查看详情
          </Button>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewItems(record)}
          >
            明细
          </Button>
        </>
      ),
    },
  ];

  const totalSettlements = settlements
    .filter((s) => s.status === 'completed')
    .reduce((sum, s) => sum + s.settlement_amount, 0);
  const pendingApprovals = settlements.filter(
    (s) => s.merchant_confirmed && !s.finance_approved
  ).length;

  return (
    <div className={styles.settlements}>
      <h2 className={styles.pageTitle}>结算管理</h2>

      <Row gutter={[16, 16]} className={styles.statsRow}>
        <Col xs={24} sm={8}>
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
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="待审批"
              value={pendingApprovals}
              prefix={<AuditOutlined />}
              suffix="个"
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="结算总数"
              value={settlements.length}
              prefix={<BankOutlined />}
              suffix="个"
            />
          </Card>
        </Col>
      </Row>

      <Card className={styles.tableCard}>
        <div className={styles.cardHeader}>
          <Space size="middle" wrap>
            <Select
              placeholder="商户筛选"
              style={{ width: 200 }}
              value={filterMerchantId}
              onChange={setFilterMerchantId}
              allowClear
              showSearch
              filterOption={(input, option) =>
                (option?.children as unknown as string)?.toLowerCase().includes(input.toLowerCase())
              }
            >
              {merchants.map((merchant) => (
                <Select.Option key={merchant.id} value={merchant.id}>
                  {merchant.company_name} (ID: {merchant.id})
                </Select.Option>
              ))}
            </Select>
            <Select
              placeholder="状态筛选"
              style={{ width: 120 }}
              value={filterStatus}
              onChange={setFilterStatus}
              allowClear
            >
              <Select.Option value="pending">待处理</Select.Option>
              <Select.Option value="processing">处理中</Select.Option>
              <Select.Option value="completed">已完成</Select.Option>
            </Select>
            <Select
              placeholder="商户确认"
              style={{ width: 120 }}
              value={filterMerchantConfirmed}
              onChange={setFilterMerchantConfirmed}
              allowClear
            >
              <Select.Option value="true">已确认</Select.Option>
              <Select.Option value="false">未确认</Select.Option>
            </Select>
            <Select
              placeholder="财务审批"
              style={{ width: 120 }}
              value={filterFinanceApproved}
              onChange={setFilterFinanceApproved}
              allowClear
            >
              <Select.Option value="true">已审批</Select.Option>
              <Select.Option value="false">未审批</Select.Option>
            </Select>
            <DatePicker
              picker="month"
              placeholder="选择年月"
              value={filterYearMonth}
              onChange={setFilterYearMonth}
              style={{ width: 150 }}
            />
            <Space.Compact>
              <InputNumber
                placeholder="最小金额"
                style={{ width: 120 }}
                value={filterMinAmount}
                onChange={(value) => setFilterMinAmount(value)}
                min={0}
                precision={2}
              />
              <InputNumber
                placeholder="最大金额"
                style={{ width: 120 }}
                value={filterMaxAmount}
                onChange={(value) => setFilterMaxAmount(value)}
                min={0}
                precision={2}
              />
            </Space.Compact>
            <Button icon={<ReloadOutlined />} onClick={handleResetFilters}>
              重置
            </Button>
          </Space>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button onClick={() => setGenerateMerchantVisible(true)}>按商户生成</Button>
            <Button type="primary" onClick={() => setGenerateVisible(true)}>
              生成月度结算
            </Button>
          </div>
        </div>
        <Table
          columns={columns}
          dataSource={settlements}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{ pageSize: 20 }}
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
              <Descriptions.Item label="商户ID">{selectedSettlement.merchant_id}</Descriptions.Item>
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
                ¥
                {selectedSettlement.total_sales_cny?.toFixed(6) ||
                  selectedSettlement.total_sales?.toFixed(6) ||
                  '0.000000'}
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
              {selectedSettlement.merchant_confirmed && !selectedSettlement.finance_approved && (
                <Button
                  type="primary"
                  icon={<CheckCircleOutlined />}
                  onClick={handleApproveSettlement}
                  loading={approveLoading}
                  style={{ marginRight: 8 }}
                >
                  审批通过
                </Button>
              )}
              {selectedSettlement.finance_approved &&
                selectedSettlement.status === 'processing' && (
                  <Button
                    type="primary"
                    icon={<BankOutlined />}
                    onClick={handleMarkAsPaid}
                    loading={markPaidLoading}
                  >
                    标记已打款
                  </Button>
                )}
            </div>
          </>
        )}
      </Modal>

      <Modal
        title="生成月度结算"
        open={generateVisible}
        onCancel={() => {
          setGenerateVisible(false);
          setGenerateYearMonth(null);
        }}
        footer={null}
      >
        <Form
          layout="vertical"
          onFinish={() => {
            if (!generateYearMonth) {
              message.warning('请选择年月');
              return;
            }
            handleGenerateSettlements({
              year: generateYearMonth.year(),
              month: generateYearMonth.month() + 1,
            });
          }}
        >
          <Form.Item label="选择年月" required>
            <DatePicker
              picker="month"
              placeholder="选择年月"
              value={generateYearMonth}
              onChange={setGenerateYearMonth}
              style={{ width: '100%' }}
              disabledDate={(current) => current && current > dayjs().endOf('month')}
            />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={generateLoading} block>
              生成结算
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="结算明细"
        open={itemsVisible}
        onCancel={() => {
          setItemsVisible(false);
          setSettlementItems([]);
        }}
        footer={null}
        width={1000}
      >
        <Table
          dataSource={settlementItems}
          rowKey="id"
          loading={itemsLoading}
          scroll={{ x: 900 }}
          pagination={{ pageSize: 10 }}
          columns={[
            {
              title: 'ID',
              dataIndex: 'id',
              key: 'id',
              width: 60,
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
              render: (v: number) => v.toLocaleString(),
            },
            {
              title: '输出Tokens',
              dataIndex: 'output_tokens',
              key: 'output_tokens',
              width: 100,
              render: (v: number) => v.toLocaleString(),
            },
            {
              title: '费用',
              dataIndex: 'cost',
              key: 'cost',
              width: 100,
              render: (v: number) => `$${v.toFixed(4)}`,
            },
            {
              title: '时间',
              dataIndex: 'created_at',
              key: 'created_at',
              width: 180,
              render: (v: string) => new Date(v).toLocaleString('zh-CN'),
            },
          ]}
        />
      </Modal>

      <Modal
        title="按商户生成结算"
        open={generateMerchantVisible}
        onCancel={() => {
          setGenerateMerchantVisible(false);
          setSelectedMerchantId(null);
          setGenerateMerchantYearMonth(null);
        }}
        onOk={handleGenerateForMerchant}
        confirmLoading={generateMerchantLoading}
        okText="生成结算"
        cancelText="取消"
      >
        <Form layout="vertical">
          <Form.Item label="选择商户" required>
            <Select
              placeholder="请选择商户"
              value={selectedMerchantId}
              onChange={setSelectedMerchantId}
              style={{ width: '100%' }}
              showSearch
              filterOption={(input, option) =>
                (option?.children as unknown as string)?.toLowerCase().includes(input.toLowerCase())
              }
            >
              {merchants.map((merchant) => (
                <Select.Option key={merchant.id} value={merchant.id}>
                  {merchant.company_name} (ID: {merchant.id})
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item label="选择年月" required>
            <DatePicker
              picker="month"
              placeholder="选择年月"
              value={generateMerchantYearMonth}
              onChange={setGenerateMerchantYearMonth}
              style={{ width: '100%' }}
              disabledDate={(current) => current && current > dayjs().endOf('month')}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminSettlements;
