import React, { useEffect, useState, useCallback } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Descriptions,
  Modal,
  Spin,
  Switch,
  Select,
  Input,
  message,
  Alert,
  Progress,
  Tabs,
} from 'antd';
import {
  SyncOutlined,
  LineChartOutlined,
  EyeOutlined,
  ThunderboltOutlined,
  SearchOutlined,
  UserOutlined,
} from '@ant-design/icons';
import api from '@/services/api';

interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface APIKeyStatus {
  id: number;
  merchant_id: number;
  name: string;
  provider: string;
  status: string;
  region: string;
  security_level: string;
  endpoint_url: string;
  health_status: string;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  error_rate: number;
  success_rate: number;
  connection_pool_size: number;
  connection_pool_active: number;
  rate_limit_remaining: number;
  load_balance_weight: number;
  last_request_at: string;
  status_updated_at: string;
  created_at: string;
}

interface MerchantAPIKeyStatus {
  id: number;
  merchant_id: number;
  name: string;
  provider: string;
  status: string;
  region: string;
  security_level: string;
  endpoint_url: string;
  health_status: string;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  error_rate: number;
  success_rate: number;
  connection_pool_size: number;
  connection_pool_active: number;
  rate_limit_remaining: number;
  load_balance_weight: number;
  last_request_at: string;
  status_updated_at: string;
  created_at: string;
}

interface APIKeyDetail extends APIKeyStatus {
  api_key_id: number;
}

const AdminAPIKeyStatus: React.FC = () => {
  const [statuses, setStatuses] = useState<APIKeyStatus[]>([]);
  const [merchantStatuses, setMerchantStatuses] = useState<MerchantAPIKeyStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [collectLoading, setCollectLoading] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<APIKeyDetail | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [filterProvider, setFilterProvider] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [searchKeyword, setSearchKeyword] = useState('');
  const [merchantIdSearch, setMerchantIdSearch] = useState('');
  const [activeTab, setActiveTab] = useState('all');

  const fetchStatuses = useCallback(async () => {
    try {
      setLoading(true);
      const response = await api.get<APIResponse<APIKeyStatus[]>>('/admin/api-key-status');
      if (response.data && response.data.code === 0) {
        setStatuses(response.data.data || []);
      }
    } catch (error) {
      message.error('获取API Key状态失败');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchStatusesByMerchantId = useCallback(async (merchantId: number) => {
    try {
      setLoading(true);
      const response = await api.get<APIResponse<MerchantAPIKeyStatus[]>>(
        `/admin/api-key-status/merchant?merchant_id=${merchantId}`
      );
      if (response.data && response.data.code === 0) {
        setMerchantStatuses(response.data.data || []);
        setActiveTab('merchant');
      }
    } catch (error) {
      message.error('获取商户API Key状态失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatuses();
  }, [fetchStatuses]);

  useEffect(() => {
    let interval: NodeJS.Timeout;
    if (autoRefresh) {
      interval = setInterval(fetchStatuses, refreshInterval * 1000);
    }
    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [autoRefresh, refreshInterval, fetchStatuses]);

  const handleRefresh = () => {
    if (activeTab === 'merchant' && merchantIdSearch) {
      fetchStatusesByMerchantId(parseInt(merchantIdSearch));
    } else {
      fetchStatuses();
    }
  };

  const handleMerchantIdSearch = () => {
    if (!merchantIdSearch) {
      message.warning('请输入商户ID');
      return;
    }
    const merchantId = parseInt(merchantIdSearch);
    if (isNaN(merchantId)) {
      message.error('商户ID必须是数字');
      return;
    }
    fetchStatusesByMerchantId(merchantId);
  };

  const handleTriggerCollect = async () => {
    try {
      setCollectLoading(true);
      const response = await api.post<APIResponse<{ message: string }>>(
        '/admin/api-key-status/collect'
      );
      if (response.data && response.data.code === 0) {
        message.success('状态采集已触发，请稍后刷新查看结果');
        setTimeout(() => {
          handleRefresh();
        }, 3000);
      }
    } catch (error) {
      message.error('触发状态采集失败');
    } finally {
      setCollectLoading(false);
    }
  };

  const handleViewDetail = async (apiKeyID: number) => {
    try {
      const response = await api.get<APIResponse<APIKeyStatus>>(`/admin/api-key-status/${apiKeyID}`);

      if (response.data.code === 0) {
        const statusData = response.data.data;

        const detail: APIKeyDetail = {
          ...statusData,
          api_key_id: statusData.id,
        };

        setSelectedStatus(detail);
        setDetailModalVisible(true);
      }
    } catch (error) {
      message.error('获取API Key详情失败');
    }
  };

  const getLatencyColor = (latency: number) => {
    if (latency < 100) return 'green';
    if (latency < 500) return 'orange';
    return 'red';
  };

  const getErrorRateColor = (errorRate: number) => {
    if (errorRate === 0) return 'green';
    if (errorRate < 0.1) return 'orange';
    return 'red';
  };

  const getSuccessRateColor = (successRate: number) => {
    if (successRate === 1) return 'green';
    if (successRate > 0.9) return 'orange';
    return 'red';
  };

  const allColumns = [
    {
      title: 'ID',
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
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 120,
      render: (provider: string) => <Tag color="blue">{provider}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      width: 80,
      render: (region: string) => (
        <Tag color="orange">{region === 'domestic' ? '国内' : '海外'}</Tag>
      ),
    },
    {
      title: '健康状态',
      dataIndex: 'health_status',
      key: 'health_status',
      width: 100,
      render: (healthStatus: string) => {
        if (healthStatus === 'healthy') return <Tag color="green">健康</Tag>;
        if (healthStatus === 'degraded') return <Tag color="orange">降级</Tag>;
        return <Tag color="red">异常</Tag>;
      },
    },
    {
      title: '延迟 (ms)',
      key: 'latency',
      width: 180,
      render: (_: unknown, record: APIKeyStatus) => (
        <Space direction="vertical" size={2}>
          <div>
            <span>P50: </span>
            <Tag color={getLatencyColor(record.latency_p50)}>{record.latency_p50}</Tag>
          </div>
          <div>
            <span>P95: </span>
            <Tag color={getLatencyColor(record.latency_p95)}>{record.latency_p95}</Tag>
          </div>
          <div>
            <span>P99: </span>
            <Tag color={getLatencyColor(record.latency_p99)}>{record.latency_p99}</Tag>
          </div>
        </Space>
      ),
    },
    {
      title: '错误率',
      dataIndex: 'error_rate',
      key: 'error_rate',
      width: 100,
      render: (errorRate: number) => (
        <Tag color={getErrorRateColor(errorRate)}>{(errorRate * 100).toFixed(2)}%</Tag>
      ),
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      width: 100,
      render: (successRate: number) => (
        <Tag color={getSuccessRateColor(successRate)}>{(successRate * 100).toFixed(2)}%</Tag>
      ),
    },
    {
      title: '连接池',
      key: 'connection_pool',
      width: 120,
      render: (_: unknown, record: APIKeyStatus) => (
        <div>
          <Progress
            percent={(record.connection_pool_active / record.connection_pool_size) * 100}
            size="small"
            status={
              record.connection_pool_active > record.connection_pool_size * 0.8
                ? 'exception'
                : 'normal'
            }
          />
          <div style={{ fontSize: 12, color: '#666', marginTop: 2 }}>
            {record.connection_pool_active} / {record.connection_pool_size}
          </div>
        </div>
      ),
    },
    {
      title: '限流剩余',
      dataIndex: 'rate_limit_remaining',
      key: 'rate_limit_remaining',
      width: 100,
    },
    {
      title: '负载均衡权重',
      dataIndex: 'load_balance_weight',
      key: 'load_balance_weight',
      width: 120,
      render: (weight: number) => <Tag color="blue">{weight.toFixed(2)}</Tag>,
    },
    {
      title: '最后请求',
      dataIndex: 'last_request_at',
      key: 'last_request_at',
      width: 150,
      render: (lastRequestAt: string) => (
        <span>{lastRequestAt || '-'}</span>
      ),
    },
    {
      title: '状态更新',
      dataIndex: 'status_updated_at',
      key: 'status_updated_at',
      width: 150,
      render: (statusUpdatedAt: string) => <span>{statusUpdatedAt || '-'}</span>,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: unknown, record: APIKeyStatus) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record.id)}
        >
          详情
        </Button>
      ),
    },
  ];

  const merchantColumns = [
    {
      title: 'ID',
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
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 120,
      render: (provider: string) => <Tag color="blue">{provider}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      width: 80,
      render: (region: string) => (
        <Tag color="orange">{region === 'domestic' ? '国内' : '海外'}</Tag>
      ),
    },
    {
      title: '安全等级',
      dataIndex: 'security_level',
      key: 'security_level',
      width: 100,
      render: (level: string) => (
        <Tag color={level === 'high' ? 'red' : 'default'}>{level === 'high' ? '高' : '标准'}</Tag>
      ),
    },
    {
      title: '健康状态',
      dataIndex: 'health_status',
      key: 'health_status',
      width: 100,
      render: (healthStatus: string) => {
        if (healthStatus === 'healthy') return <Tag color="green">健康</Tag>;
        if (healthStatus === 'degraded') return <Tag color="orange">降级</Tag>;
        return <Tag color="red">异常</Tag>;
      },
    },
    {
      title: '延迟 (ms)',
      key: 'latency',
      width: 150,
      render: (_: unknown, record: MerchantAPIKeyStatus) => (
        <Space direction="vertical" size={2}>
          <div>
            P50: <Tag color={getLatencyColor(record.latency_p50)}>{record.latency_p50}</Tag>
          </div>
          <div>
            P95: <Tag color={getLatencyColor(record.latency_p95)}>{record.latency_p95}</Tag>
          </div>
        </Space>
      ),
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      width: 100,
      render: (successRate: number) => (
        <Tag color={getSuccessRateColor(successRate)}>{(successRate * 100).toFixed(2)}%</Tag>
      ),
    },
    {
      title: '错误率',
      dataIndex: 'error_rate',
      key: 'error_rate',
      width: 100,
      render: (errorRate: number) => (
        <Tag color={getErrorRateColor(errorRate)}>{(errorRate * 100).toFixed(2)}%</Tag>
      ),
    },
    {
      title: '状态更新',
      dataIndex: 'status_updated_at',
      key: 'status_updated_at',
      width: 150,
      render: (updatedAt: string) => <span>{updatedAt || '-'}</span>,
    },
  ];

  const tabItems = [
    {
      key: 'all',
      label: '全部API Key状态',
      children: (
        <Table
          columns={allColumns}
          dataSource={statuses}
          rowKey="id"
          loading={loading && activeTab === 'all'}
          pagination={{
            pageSize: 20,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
          }}
          scroll={{ x: 'max-content' }}
        />
      ),
    },
    {
      key: 'merchant',
      label: (
        <span>
          <UserOutlined style={{ marginRight: 4 }} />
          按商户查询
        </span>
      ),
      children: (
        <div>
          <Space style={{ marginBottom: 16 }}>
            <Input
              placeholder="输入商户ID"
              value={merchantIdSearch}
              onChange={(e) => setMerchantIdSearch(e.target.value)}
              onPressEnter={handleMerchantIdSearch}
              style={{ width: 200 }}
              prefix={<SearchOutlined />}
            />
            <Button type="primary" onClick={handleMerchantIdSearch}>
              查询
            </Button>
            <Button onClick={() => setMerchantStatuses([])}>清空</Button>
          </Space>
          <Table
            columns={merchantColumns}
            dataSource={merchantStatuses}
            rowKey="id"
            loading={loading && activeTab === 'merchant'}
            pagination={{
              pageSize: 20,
              showSizeChanger: true,
              showTotal: (total) => `共 ${total} 条`,
            }}
            scroll={{ x: 'max-content' }}
            locale={{ emptyText: '请输入商户ID进行查询' }}
          />
        </div>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <LineChartOutlined />
          API Key 状态管理
        </Space>
      }
      extra={
        <Space>
          <Input
            allowClear
            placeholder="搜索 API Key ID"
            style={{ width: 200 }}
            value={searchKeyword}
            onChange={(e) => setSearchKeyword(e.target.value)}
          />
          <Select
            style={{ width: 150 }}
            value={filterProvider}
            onChange={setFilterProvider}
            options={[{ value: 'all', label: '所有提供商' }]}
          />
          <Select
            style={{ width: 120 }}
            value={filterStatus}
            onChange={setFilterStatus}
            options={[
              { value: 'all', label: '所有状态' },
              { value: 'active', label: '启用' },
              { value: 'inactive', label: '禁用' },
            ]}
          />
          <Space>
            <span style={{ fontSize: 12, color: '#666' }}>自动刷新</span>
            <Switch checked={autoRefresh} onChange={setAutoRefresh} />
          </Space>
          <Select
            style={{ width: 120 }}
            value={refreshInterval}
            onChange={setRefreshInterval}
            disabled={!autoRefresh}
            options={[
              { value: 5, label: '5秒' },
              { value: 10, label: '10秒' },
              { value: 30, label: '30秒' },
              { value: 60, label: '1分钟' },
            ]}
          />
          <Button icon={<SyncOutlined spin={loading} />} onClick={handleRefresh} loading={loading}>
            刷新
          </Button>
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleTriggerCollect}
            loading={collectLoading}
          >
            立即采集状态
          </Button>
        </Space>
      }
    >
      <Alert
        message="状态说明"
        description={
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            <li>延迟：P50/P95/P99 分别表示50%/95%/99%的请求延迟时间</li>
            <li>错误率：最近一段时间内的错误请求比例</li>
            <li>成功率：最近一段时间内的成功请求比例</li>
            <li>连接池：当前活跃连接数/总连接数</li>
            <li>限流剩余：当前剩余的请求配额</li>
            <li>负载均衡权重：API Key在负载均衡中的权重值</li>
            <li>
              <strong>按商户查询</strong>：输入商户ID可查询该商户上传的所有API Key状态（BYOK模式）
            </li>
          </ul>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Tabs activeKey={activeTab} onChange={setActiveTab} items={tabItems} />

      <Modal
        title="API Key 详细状态"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[<Button key="close" onClick={() => setDetailModalVisible(false)}>关闭</Button>]}
        width={800}
      >
        {selectedStatus ? (
          <Descriptions bordered column={2}>
            <Descriptions.Item label="API Key ID">{selectedStatus.api_key_id}</Descriptions.Item>
            <Descriptions.Item label="商户ID">{selectedStatus.merchant_id}</Descriptions.Item>
            <Descriptions.Item label="名称">{selectedStatus.name}</Descriptions.Item>
            <Descriptions.Item label="提供商">
              <Tag color="blue">{selectedStatus.provider}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={selectedStatus.status === 'active' ? 'green' : 'default'}>
                {selectedStatus.status === 'active' ? '启用' : '禁用'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="区域">
              <Tag color="orange">{selectedStatus.region === 'domestic' ? '国内' : '海外'}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="健康状态">
              {selectedStatus.health_status === 'healthy' ? (
                <Tag color="green">健康</Tag>
              ) : selectedStatus.health_status === 'degraded' ? (
                <Tag color="orange">降级</Tag>
              ) : (
                <Tag color="red">异常</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="安全等级">
              <Tag color={selectedStatus.security_level === 'high' ? 'red' : 'default'}>
                {selectedStatus.security_level === 'high' ? '高' : '标准'}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="端点URL" span={2}>
              <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>
                {selectedStatus.endpoint_url || '-'}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label="延迟指标 (ms)">
              <Space direction="vertical" size={4}>
                <div>
                  <span>P50: </span>
                  <Tag color={getLatencyColor(selectedStatus.latency_p50)}>
                    {selectedStatus.latency_p50}
                  </Tag>
                </div>
                <div>
                  <span>P95: </span>
                  <Tag color={getLatencyColor(selectedStatus.latency_p95)}>
                    {selectedStatus.latency_p95}
                  </Tag>
                </div>
                <div>
                  <span>P99: </span>
                  <Tag color={getLatencyColor(selectedStatus.latency_p99)}>
                    {selectedStatus.latency_p99}
                  </Tag>
                </div>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="错误率">
              <Tag color={getErrorRateColor(selectedStatus.error_rate)}>
                {(selectedStatus.error_rate * 100).toFixed(2)}%
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="成功率">
              <Tag color={getSuccessRateColor(selectedStatus.success_rate)}>
                {(selectedStatus.success_rate * 100).toFixed(2)}%
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="连接池状态">
              <div>
                <Progress
                  percent={
                    (selectedStatus.connection_pool_active /
                      selectedStatus.connection_pool_size) *
                    100
                  }
                  status={
                    selectedStatus.connection_pool_active >
                    selectedStatus.connection_pool_size * 0.8
                      ? 'exception'
                      : 'normal'
                  }
                />
                <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                  活跃连接: {selectedStatus.connection_pool_active} /{' '}
                  {selectedStatus.connection_pool_size}
                </div>
              </div>
            </Descriptions.Item>
            <Descriptions.Item label="限流剩余">
              {selectedStatus.rate_limit_remaining}
            </Descriptions.Item>
            <Descriptions.Item label="负载均衡权重">
              <Tag color="blue">{selectedStatus.load_balance_weight.toFixed(2)}</Tag>
            </Descriptions.Item>
            <Descriptions.Item label="最后请求时间">
              {selectedStatus.last_request_at || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="状态更新时间">
              {selectedStatus.status_updated_at || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {selectedStatus.created_at || '-'}
            </Descriptions.Item>
          </Descriptions>
        ) : (
          <Spin size="large" style={{ textAlign: 'center', padding: '40px 0' }} />
        )}
      </Modal>
    </Card>
  );
};

export default AdminAPIKeyStatus;
