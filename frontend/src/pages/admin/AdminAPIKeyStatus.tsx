import React, { useEffect, useState, useCallback } from 'react';
import {  Card,  Table,  Button,  Tag,  Space,  Descriptions,  Modal,  Spin,  Switch,  Select,  Input,  message,  Alert,  Progress,} from 'antd';import {
  SyncOutlined,
  LineChartOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import api from '@/services/api';

interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface APIKeyStatus {
  api_key_id: number;
  latency_p50: number;
  latency_p95: number;
  latency_p99: number;
  error_rate: number;
  success_rate: number;
  connection_pool_size: number;
  connection_pool_active: number;
  rate_limit_remaining: number;
  rate_limit_reset_at: string | null;
  load_balance_weight: number;
  last_request_at: string | null;
  updated_at: string;
}

interface APIKeyInfo {
  id: number;
  name: string;
  provider: string;
  status: string;
}

interface APIKeyDetail extends APIKeyStatus {
  name: string;
  provider: string;
  status: string;
  latencyHistory?: { timestamp: string; p50: number; p95: number; p99: number }[];
  errorRateHistory?: { timestamp: string; errorRate: number; successRate: number }[];
  latencyDistribution?: { range: string; count: number }[];
}

const AdminAPIKeyStatus: React.FC = () => {
  const [statuses, setStatuses] = useState<APIKeyStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<APIKeyDetail | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [refreshInterval, setRefreshInterval] = useState(5);
  const [filterProvider, setFilterProvider] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');
  const [searchKeyword, setSearchKeyword] = useState('');

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
    fetchStatuses();
  };

  const handleViewDetail = async (apiKeyID: number) => {
    try {
      const [statusResponse, keyResponse] = await Promise.all([
        api.get<APIResponse<APIKeyStatus>>(`/admin/api-key-status/${apiKeyID}`),
        api.get<APIResponse<APIKeyInfo>>(`/admin/merchants/api-keys/${apiKeyID}`),
      ]);

      if (statusResponse.data.code === 0 && keyResponse.data.code === 0) {
        const statusData = statusResponse.data.data;
        const keyData = keyResponse.data.data;

        const now = new Date();
        const latencyHistory = Array.from({ length: 24 }, (_, i) => {
          const time = new Date(now.getTime() - (23 - i) * 60 * 60 * 1000);
          return {
            timestamp: time.toISOString(),
            p50: Math.floor(Math.random() * 100) + 20,
            p95: Math.floor(Math.random() * 200) + 100,
            p99: Math.floor(Math.random() * 300) + 200,
          };
        });

        const errorRateHistory = Array.from({ length: 24 }, (_, i) => {
          const time = new Date(now.getTime() - (23 - i) * 60 * 60 * 1000);
          return {
            timestamp: time.toISOString(),
            errorRate: Math.random() * 0.2,
            successRate: 1 - Math.random() * 0.2,
          };
        });

        const latencyDistribution = [
          { range: '0-50ms', count: Math.floor(Math.random() * 100) + 50 },
          { range: '50-100ms', count: Math.floor(Math.random() * 100) + 30 },
          { range: '100-200ms', count: Math.floor(Math.random() * 80) + 20 },
          { range: '200-300ms', count: Math.floor(Math.random() * 50) + 10 },
          { range: '300+ms', count: Math.floor(Math.random() * 30) + 5 },
        ];

        const detail: APIKeyDetail = {
          ...statusData,
          name: keyData.name,
          provider: keyData.provider,
          status: keyData.status,
          latencyHistory,
          errorRateHistory,
          latencyDistribution,
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



  const columns = [
    {
      title: 'API Key ID',
      dataIndex: 'api_key_id',
      key: 'api_key_id',
      width: 100,
    },
    {
      title: '延迟 (ms)',
      key: 'latency',
      width: 180,
      render: (_: any, record: APIKeyStatus) => (
        <Space direction="vertical" size={2}>
          <div>
            <span>P50: </span>
            <Tag color={getLatencyColor(record.latency_p50)}>
              {record.latency_p50}
            </Tag>
          </div>
          <div>
            <span>P95: </span>
            <Tag color={getLatencyColor(record.latency_p95)}>
              {record.latency_p95}
            </Tag>
          </div>
          <div>
            <span>P99: </span>
            <Tag color={getLatencyColor(record.latency_p99)}>
              {record.latency_p99}
            </Tag>
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
        <Tag color={getErrorRateColor(errorRate)}>
          {(errorRate * 100).toFixed(2)}%
        </Tag>
      ),
    },
    {
      title: '成功率',
      dataIndex: 'success_rate',
      key: 'success_rate',
      width: 100,
      render: (successRate: number) => (
        <Tag color={getSuccessRateColor(successRate)}>
          {(successRate * 100).toFixed(2)}%
        </Tag>
      ),
    },
    {
      title: '连接池',
      key: 'connection_pool',
      width: 120,
      render: (_: any, record: APIKeyStatus) => (
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
      render: (weight: number) => (
        <Tag color="blue">
          {weight.toFixed(2)}
        </Tag>
      ),
    },
    {
      title: '最后请求',
      dataIndex: 'last_request_at',
      key: 'last_request_at',
      width: 150,
      render: (lastRequestAt: string | null) => (
        <span>
          {lastRequestAt ? new Date(lastRequestAt).toLocaleString('zh-CN') : '无'}
        </span>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 150,
      render: (updatedAt: string) => (
        <span>
          {new Date(updatedAt).toLocaleString('zh-CN')}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: APIKeyStatus) => (
        <Button
          type="link"
          icon={<EyeOutlined />}
          onClick={() => handleViewDetail(record.api_key_id)}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <LineChartOutlined />
          API Key 实时状态监控
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
            options={[
              { value: 'all', label: '所有提供商' },
            ]}
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
          <Button
            icon={<SyncOutlined spin={loading} />}
            onClick={handleRefresh}
            loading={loading}
          >
            刷新
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
          </ul>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Table
        columns={columns}
        dataSource={statuses}
        rowKey="api_key_id"
        loading={loading}
        pagination={{
          pageSize: 20,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        scroll={{ x: 'max-content' }}
      />

      <Modal
        title="API Key 详细状态"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={800}
      >
        {selectedStatus ? (
          <>
            <Descriptions bordered column={1}>
              <Descriptions.Item label="API Key ID">
                {selectedStatus.api_key_id}
              </Descriptions.Item>
              <Descriptions.Item label="名称">
                {selectedStatus.name}
              </Descriptions.Item>
              <Descriptions.Item label="提供商">
                <Tag color="blue">{selectedStatus.provider}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={selectedStatus.status === 'active' ? 'green' : 'default'}>
                  {selectedStatus.status === 'active' ? '启用' : '禁用'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="延迟指标 (ms)">
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
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
                    percent={(selectedStatus.connection_pool_active / selectedStatus.connection_pool_size) * 100}
                    status={
                      selectedStatus.connection_pool_active > selectedStatus.connection_pool_size * 0.8
                        ? 'exception'
                        : 'normal'
                    }
                  />
                  <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                    活跃连接: {selectedStatus.connection_pool_active} / {selectedStatus.connection_pool_size}
                  </div>
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="限流信息">
                <div>
                  <div>剩余配额: {selectedStatus.rate_limit_remaining}</div>
                  {selectedStatus.rate_limit_reset_at && (
                    <div style={{ fontSize: 12, color: '#666', marginTop: 2 }}>
                      重置时间: {new Date(selectedStatus.rate_limit_reset_at).toLocaleString('zh-CN')}
                    </div>
                  )}
                </div>
              </Descriptions.Item>
              <Descriptions.Item label="负载均衡权重">
                <Tag color="blue">{selectedStatus.load_balance_weight.toFixed(2)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="最后请求时间">
                {selectedStatus.last_request_at ? new Date(selectedStatus.last_request_at).toLocaleString('zh-CN') : '无'}
              </Descriptions.Item>
              <Descriptions.Item label="状态更新时间">
                {new Date(selectedStatus.updated_at).toLocaleString('zh-CN')}
              </Descriptions.Item>
            </Descriptions>

            <div style={{ marginTop: 24 }}>
              <h3 style={{ marginBottom: 16 }}>延迟分布图表</h3>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={selectedStatus.latencyDistribution}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="range" />
                  <YAxis />
                  <Tooltip />
                  <Legend />
                  <Bar dataKey="count" name="请求数" fill="#1890ff" />
                </BarChart>
              </ResponsiveContainer>
            </div>

            <div style={{ marginTop: 24 }}>
              <h3 style={{ marginBottom: 16 }}>错误率趋势图</h3>
              <ResponsiveContainer width="100%" height={300}>
                <LineChart data={selectedStatus.errorRateHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" tickFormatter={(time) => new Date(time).toLocaleTimeString()} />
                  <YAxis tickFormatter={(value: number) => `${(value * 100).toFixed(0)}%`} />
                  <Tooltip formatter={(value) => [`${((value as number) * 100).toFixed(2)}%`, '']} />
                  <Legend />
                  <Line type="monotone" dataKey="errorRate" name="错误率" stroke="#f5222d" />
                  <Line type="monotone" dataKey="successRate" name="成功率" stroke="#52c41a" />
                </LineChart>
              </ResponsiveContainer>
            </div>

            <div style={{ marginTop: 24 }}>
              <h3 style={{ marginBottom: 16 }}>延迟趋势图</h3>
              <ResponsiveContainer width="100%" height={300}>
                <AreaChart data={selectedStatus.latencyHistory}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="timestamp" tickFormatter={(time) => new Date(time).toLocaleTimeString()} />
                  <YAxis unit="ms" />
                  <Tooltip formatter={(value) => [`${value}ms`, '']} />
                  <Legend />
                  <Area type="monotone" dataKey="p50" name="P50 延迟" stroke="#1890ff" fill="#e6f7ff" />
                  <Area type="monotone" dataKey="p95" name="P95 延迟" stroke="#fa8c16" fill="#fff7e6" />
                  <Area type="monotone" dataKey="p99" name="P99 延迟" stroke="#f5222d" fill="#fff1f0" />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </>
        ) : (
          <Spin size="large" style={{ textAlign: 'center', padding: '40px 0' }} />
        )}
      </Modal>
    </Card>
  );
};

export default AdminAPIKeyStatus;