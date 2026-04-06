import { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Modal, Descriptions, message, Statistic, Row, Col, Select, DatePicker, Space, Pagination } from 'antd';
import type { Dayjs } from 'dayjs';
import dayjs from 'dayjs';
import {
  DollarOutlined,
  ApiOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  EyeOutlined,
  ReloadOutlined,
  DownloadOutlined,
} from '@ant-design/icons';
import styles from './AdminSettlements.module.css';

const getAuthToken = () => {
  return localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token') || '';
};

interface BillingRecord {
  id: number;
  merchant_id: number;
  company_name: string;
  user_id?: number;
  username?: string;
  provider: string;
  model: string;
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  cost: number;
  request_time: number;
  status_code: number;
  created_at: string;
}

interface BillingStats {
  total_cost: number;
  total_requests: number;
  total_tokens: number;
  average_latency: number;
  success_rate: number;
  provider_breakdown: Record<string, {
    provider: string;
    total_cost: number;
    total_requests: number;
    percentage: number;
  }>;
}

const AdminBillings = () => {
  const [billings, setBillings] = useState<BillingRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedBilling, setSelectedBilling] = useState<BillingRecord | null>(null);
  const [merchants, setMerchants] = useState<Array<{id: number; company_name: string}>>([]);
  const [stats, setStats] = useState<BillingStats | null>(null);

  const [filterMerchantId, setFilterMerchantId] = useState<number | null>(null);
  const [filterProvider, setFilterProvider] = useState<string>('');
  const [filterModel, setFilterModel] = useState<string>('');
  const [filterStartDate, setFilterStartDate] = useState<Dayjs | null>(null);
  const [filterEndDate, setFilterEndDate] = useState<Dayjs | null>(null);

  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);

  useEffect(() => {
    fetchBillings();
    fetchMerchants();
    fetchStats();
  }, [page, pageSize, filterMerchantId, filterProvider, filterModel, filterStartDate, filterEndDate]);

  const fetchBillings = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      params.append('page', page.toString());
      params.append('page_size', pageSize.toString());

      if (filterMerchantId) params.append('merchant_id', filterMerchantId.toString());
      if (filterProvider) params.append('provider', filterProvider);
      if (filterModel) params.append('model', filterModel);
      if (filterStartDate) params.append('start_date', filterStartDate.format('YYYY-MM-DD'));
      if (filterEndDate) params.append('end_date', filterEndDate.format('YYYY-MM-DD'));

      const response = await fetch(`/api/v1/admin/billings?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setBillings(data.billings || []);
        setTotal(data.total || 0);
      } else {
        message.error('获取账单列表失败');
      }
    } catch (error) {
      message.error('获取账单列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchMerchants = async () => {
    try {
      const response = await fetch('/api/v1/admin/merchants', {
        headers: {
          'Authorization': `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setMerchants(data.merchants || []);
      }
    } catch (error) {
      console.error('Failed to fetch merchants:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const params = new URLSearchParams();
      if (filterMerchantId) params.append('merchant_id', filterMerchantId.toString());
      if (filterStartDate) params.append('start_date', filterStartDate.format('YYYY-MM-DD'));
      if (filterEndDate) params.append('end_date', filterEndDate.format('YYYY-MM-DD'));

      const response = await fetch(`/api/v1/admin/billings/stats?${params.toString()}`, {
        headers: {
          'Authorization': `Bearer ${getAuthToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setStats(data);
      }
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const handleResetFilters = () => {
    setFilterMerchantId(null);
    setFilterProvider('');
    setFilterModel('');
    setFilterStartDate(null);
    setFilterEndDate(null);
    setPage(1);
  };

  const handleExport = () => {
    message.info('导出功能开发中...');
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '商户',
      dataIndex: 'company_name',
      key: 'company_name',
      width: 150,
    },
    {
      title: '用户',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      render: (text: string) => text || '-',
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
      title: '总Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      width: 100,
      render: (num: number) => num.toLocaleString(),
    },
    {
      title: '成本',
      dataIndex: 'cost',
      key: 'cost',
      width: 100,
      render: (cost: number) => `$${cost.toFixed(4)}`,
    },
    {
      title: '延迟(ms)',
      dataIndex: 'request_time',
      key: 'request_time',
      width: 100,
      render: (time: number) => time.toFixed(2),
    },
    {
      title: '状态',
      dataIndex: 'status_code',
      key: 'status_code',
      width: 80,
      render: (code: number) => (
        <Tag color={code === 200 ? 'success' : 'error'}>
          {code}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      fixed: 'right' as const,
      render: (_: any, record: BillingRecord) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => {
            setSelectedBilling(record);
            setDetailVisible(true);
          }}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <div className={styles.container}>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col span={4}>
          <Card>
            <Statistic
              title="总消费"
              value={stats?.total_cost || 0}
              precision={2}
              prefix={<DollarOutlined />}
              suffix="USD"
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="总请求数"
              value={stats?.total_requests || 0}
              prefix={<ApiOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="总Tokens"
              value={stats?.total_tokens || 0}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="平均延迟"
              value={stats?.average_latency || 0}
              precision={2}
              suffix="ms"
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col span={4}>
          <Card>
            <Statistic
              title="成功率"
              value={stats?.success_rate || 0}
              precision={2}
              suffix="%"
              prefix={<CheckCircleOutlined />}
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
              placeholder="Provider筛选"
              style={{ width: 120 }}
              value={filterProvider}
              onChange={setFilterProvider}
              allowClear
            >
              <Select.Option value="openai">OpenAI</Select.Option>
              <Select.Option value="anthropic">Anthropic</Select.Option>
              <Select.Option value="google">Google</Select.Option>
            </Select>
            <DatePicker
              placeholder="开始日期"
              value={filterStartDate}
              onChange={setFilterStartDate}
              style={{ width: 150 }}
            />
            <DatePicker
              placeholder="结束日期"
              value={filterEndDate}
              onChange={setFilterEndDate}
              style={{ width: 150 }}
            />
            <Button icon={<ReloadOutlined />} onClick={handleResetFilters}>
              重置
            </Button>
            <Button icon={<DownloadOutlined />} onClick={handleExport}>
              导出
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={billings}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1500 }}
          pagination={false}
        />

        <div style={{ marginTop: 16, textAlign: 'right' }}>
          <Pagination
            current={page}
            pageSize={pageSize}
            total={total}
            showSizeChanger
            showQuickJumper
            showTotal={(total) => `共 ${total} 条记录`}
            onChange={(page, pageSize) => {
              setPage(page);
              setPageSize(pageSize);
            }}
          />
        </div>
      </Card>

      <Modal
        title="账单详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={800}
      >
        {selectedBilling && (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="ID">{selectedBilling.id}</Descriptions.Item>
            <Descriptions.Item label="商户">{selectedBilling.company_name}</Descriptions.Item>
            <Descriptions.Item label="用户">{selectedBilling.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="Provider">{selectedBilling.provider}</Descriptions.Item>
            <Descriptions.Item label="Model">{selectedBilling.model}</Descriptions.Item>
            <Descriptions.Item label="状态码">{selectedBilling.status_code}</Descriptions.Item>
            <Descriptions.Item label="输入Tokens">{selectedBilling.input_tokens.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="输出Tokens">{selectedBilling.output_tokens.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="总Tokens">{selectedBilling.total_tokens.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="成本">${selectedBilling.cost.toFixed(4)}</Descriptions.Item>
            <Descriptions.Item label="延迟">{selectedBilling.request_time.toFixed(2)} ms</Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {dayjs(selectedBilling.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default AdminBillings;
