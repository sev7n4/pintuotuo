import React, { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  DatePicker,
  Select,
  Statistic,
  Row,
  Col,
  message,
  Descriptions,
} from 'antd';
import {
  DollarOutlined,
  ApiOutlined,
  ClockCircleOutlined,
  EyeOutlined,
  DownloadOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { billingService } from '@/services/billing';
import type { BillingRecord, BillingStats } from '@/types';

const { RangePicker } = DatePicker;

interface UserBillingRecord extends BillingRecord {
  company_name?: string;
  username?: string;
  total_tokens?: number;
  request_time?: number;
}

const AdminUserBillings: React.FC = () => {
  const [billings, setBillings] = useState<UserBillingRecord[]>([]);
  const [stats, setStats] = useState<BillingStats | null>(null);
  const [loading, setLoading] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [filterUserId, setFilterUserId] = useState<number | undefined>();
  const [filterProvider, setFilterProvider] = useState<string | undefined>();
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null);
  const [selectedBilling, setSelectedBilling] = useState<UserBillingRecord | null>(null);
  const [detailVisible, setDetailVisible] = useState(false);

  useEffect(() => {
    fetchBillings();
    fetchStats();
  }, [pagination.current, pagination.pageSize, filterUserId, filterProvider, dateRange]);

  const fetchBillings = async () => {
    setLoading(true);
    try {
      const params: any = {
        page: pagination.current,
        page_size: pagination.pageSize,
      };
      if (filterUserId) params.user_id = filterUserId;
      if (filterProvider) params.provider = filterProvider;
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await billingService.getUserBillings(params);
      const data = response.data;
      setBillings(data.billings || []);
      setPagination((prev) => ({ ...prev, total: data.total || 0 }));
    } catch (error) {
      message.error('获取用户账单失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const params: any = {};
      if (filterUserId) params.user_id = filterUserId;
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await billingService.getUserBillingStats(params);
      setStats(response.data);
    } catch (error) {
      console.error('获取统计数据失败', error);
    }
  };

  const handleExport = async () => {
    try {
      const params: any = {};
      if (filterUserId) params.user_id = filterUserId;
      if (dateRange) {
        params.start_date = dateRange[0].format('YYYY-MM-DD');
        params.end_date = dateRange[1].format('YYYY-MM-DD');
      }
      const response = await billingService.exportUserBillingsCSV(params);
      const blob = new Blob([response.data], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `user_billings_${dayjs().format('YYYYMMDDHHmmss')}.csv`;
      link.click();
      window.URL.revokeObjectURL(url);
      message.success('导出成功');
    } catch (error) {
      message.error('导出失败');
    }
  };

  const handleTableChange = (pag: any) => {
    setPagination({
      ...pagination,
      current: pag.current,
      pageSize: pag.pageSize,
    });
  };

  const columns: ColumnsType<UserBillingRecord> = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '用户ID',
      dataIndex: 'user_id',
      key: 'user_id',
      width: 80,
    },
    {
      title: '用户名',
      dataIndex: 'username',
      key: 'username',
      width: 120,
      render: (name: string) => name || '-',
    },
    {
      title: '商户',
      dataIndex: 'company_name',
      key: 'company_name',
      width: 150,
      render: (name: string) => name || '-',
    },
    {
      title: 'Provider',
      dataIndex: 'provider',
      key: 'provider',
      width: 100,
      render: (provider: string) => (
        <Tag color={provider === 'openai' ? 'green' : provider === 'anthropic' ? 'blue' : 'orange'}>
          {provider}
        </Tag>
      ),
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
      render: (num: number) => num?.toLocaleString() || '-',
    },
    {
      title: '输出Tokens',
      dataIndex: 'output_tokens',
      key: 'output_tokens',
      width: 100,
      render: (num: number) => num?.toLocaleString() || '-',
    },
    {
      title: '总Tokens',
      dataIndex: 'total_tokens',
      key: 'total_tokens',
      width: 100,
      render: (num: number) => num?.toLocaleString() || '-',
    },
    {
      title: '成本',
      dataIndex: 'cost',
      key: 'cost',
      width: 100,
      render: (cost: number) => `$${(cost || 0).toFixed(4)}`,
    },
    {
      title: '延迟(ms)',
      dataIndex: 'request_time',
      key: 'request_time',
      width: 100,
      render: (time: number) => (time || 0).toFixed(2),
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
      fixed: 'right',
      render: (_: unknown, record: UserBillingRecord) => (
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
    <div>
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

      <Card>
        <div style={{ marginBottom: 16 }}>
          <Space size="middle" wrap>
            <Select
              placeholder="用户ID筛选"
              style={{ width: 150 }}
              value={filterUserId}
              onChange={setFilterUserId}
              allowClear
            />
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
            <RangePicker
              value={dateRange}
              onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs] | null)}
            />
            <Button icon={<DownloadOutlined />} onClick={handleExport}>
              导出CSV
            </Button>
          </Space>
        </div>

        <Table
          columns={columns}
          dataSource={billings}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1800 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      <Modal
        title="账单详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={null}
        width={700}
      >
        {selectedBilling && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{selectedBilling.id}</Descriptions.Item>
            <Descriptions.Item label="用户ID">{selectedBilling.user_id}</Descriptions.Item>
            <Descriptions.Item label="用户名">{selectedBilling.username || '-'}</Descriptions.Item>
            <Descriptions.Item label="商户">{selectedBilling.company_name || '-'}</Descriptions.Item>
            <Descriptions.Item label="Provider">{selectedBilling.provider}</Descriptions.Item>
            <Descriptions.Item label="Model">{selectedBilling.model}</Descriptions.Item>
            <Descriptions.Item label="输入Tokens">{selectedBilling.input_tokens?.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="输出Tokens">{selectedBilling.output_tokens?.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="总Tokens">{selectedBilling.total_tokens?.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="成本">${(selectedBilling.cost || 0).toFixed(6)}</Descriptions.Item>
            <Descriptions.Item label="延迟">{(selectedBilling.request_time || 0).toFixed(2)} ms</Descriptions.Item>
            <Descriptions.Item label="状态码">
              <Tag color={selectedBilling.status_code === 200 ? 'success' : 'error'}>
                {selectedBilling.status_code}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="创建时间" span={2}>
              {dayjs(selectedBilling.created_at).format('YYYY-MM-DD HH:mm:ss')}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>
    </div>
  );
};

export default AdminUserBillings;
