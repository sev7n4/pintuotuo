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
  Input,
  Select,
  DatePicker,
  message,
  Alert,
  Tooltip,
  Badge,
} from 'antd';
import {
  SearchOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  FilterOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import api from '@/services/api';

interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
}

const { RangePicker } = DatePicker;

interface RouteDecisionLog {
  id: number;
  request_id: string;
  merchant_id?: number;
  api_key_id?: number;
  strategy_layer_goal: string;
  strategy_layer_input: Record<string, any>;
  strategy_layer_output: Record<string, any>;
  decision_layer_candidates: Array<{
    api_key_id?: number;
    provider?: string;
    score?: number;
    health_status?: string;
    selected?: boolean;
  }>;
  decision_layer_output: Record<string, any>;
  execution_layer_result: Record<string, any>;
  decision_duration_ms: number;
  decision_result: string;
  error_message?: string;
  created_at: string;
}

interface LogsResponse {
  logs: RouteDecisionLog[];
  total: number;
  page: number;
  size: number;
}

const AdminRouteDecisionLogs: React.FC = () => {
  const [logs, setLogs] = useState<RouteDecisionLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [detailModalVisible, setDetailModalVisible] = useState(false);
  const [selectedLog, setSelectedLog] = useState<RouteDecisionLog | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // 筛选条件
  const [filters, setFilters] = useState({
    merchant_id: '',
    api_key_id: '',
    strategy: '',
    decision_result: '',
    request_id: '',
    time_range: null as [string, string] | null,
  });

  const fetchLogs = useCallback(async () => {
    try {
      setLoading(true);

      const params: Record<string, any> = {
        page: pagination.current,
        page_size: pagination.pageSize,
      };

      if (filters.merchant_id) {
        params.merchant_id = filters.merchant_id;
      }

      if (filters.api_key_id) {
        params.api_key_id = filters.api_key_id;
      }

      if (filters.strategy) {
        params.strategy = filters.strategy;
      }

      if (filters.decision_result) {
        params.decision_result = filters.decision_result;
      }

      if (filters.request_id) {
        params.request_id = filters.request_id;
      }

      if (filters.time_range) {
        params.start_time = filters.time_range[0];
        params.end_time = filters.time_range[1];
      }

      const response = await api.get<APIResponse<LogsResponse>>('/admin/route-decision-logs', {
        params,
      });
      if (response.data && response.data.code === 0) {
        const data = response.data.data;
        setLogs(data.logs || []);
        setPagination((prev) => ({
          ...prev,
          total: data.total,
        }));
      }
    } catch (error) {
      message.error('获取路由决策日志失败');
    } finally {
      setLoading(false);
    }
  }, [pagination.current, pagination.pageSize, filters]);

  useEffect(() => {
    fetchLogs();
  }, [fetchLogs]);

  const handleRefresh = () => {
    fetchLogs();
  };

  const handleViewDetail = (log: RouteDecisionLog) => {
    setSelectedLog(log);
    setDetailModalVisible(true);
  };

  const handlePageChange = (page: number, pageSize: number) => {
    setPagination({
      ...pagination,
      current: page,
      pageSize,
    });
  };

  const handleFilterChange = (key: keyof typeof filters, value: any) => {
    setFilters({
      ...filters,
      [key]: value,
    });
  };

  const handleTimeRangeChange = (_: any, dateStrings: [string, string]) => {
    setFilters({
      ...filters,
      time_range: dateStrings,
    });
  };

  const handleResetFilters = () => {
    setFilters({
      merchant_id: '',
      api_key_id: '',
      strategy: '',
      decision_result: '',
      request_id: '',
      time_range: null,
    });
  };

  const getStrategyLabel = (strategy: string) => {
    const strategyMap: Record<string, string> = {
      performance_first: '性能优先',
      price_first: '价格优先',
      reliability_first: '可靠性优先',
      security_first: '安全优先',
      balanced: '均衡策略',
      auto: '自动模式',
    };
    return strategyMap[strategy] || strategy;
  };

  const columns = [
    {
      title: '日志ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
    },
    {
      title: '请求ID',
      dataIndex: 'request_id',
      key: 'request_id',
      width: 180,
      render: (requestId: string) => (
        <Tooltip title={requestId}>
          <span>{requestId.substring(0, 12)}...</span>
        </Tooltip>
      ),
    },
    {
      title: '商户ID',
      dataIndex: 'merchant_id',
      key: 'merchant_id',
      width: 100,
    },
    {
      title: 'API Key ID',
      dataIndex: 'api_key_id',
      key: 'api_key_id',
      width: 120,
    },
    {
      title: '策略目标',
      dataIndex: 'strategy_layer_goal',
      key: 'strategy_layer_goal',
      width: 120,
      render: (strategy: string) => <Tag color="blue">{getStrategyLabel(strategy)}</Tag>,
    },
    {
      title: '决策耗时',
      dataIndex: 'decision_duration_ms',
      key: 'decision_duration_ms',
      width: 100,
      render: (duration: number) => (
        <Tag color={duration > 10 ? 'orange' : 'green'}>{duration ? `${duration}ms` : '-'}</Tag>
      ),
    },
    {
      title: '决策结果',
      dataIndex: 'decision_result',
      key: 'decision_result',
      width: 100,
      render: (result: string) => {
        if (!result) return '-';
        const colorMap: Record<string, string> = {
          success: 'green',
          failed: 'red',
          fallback: 'orange',
        };
        const labelMap: Record<string, string> = {
          success: '成功',
          failed: '失败',
          fallback: '降级',
        };
        return <Tag color={colorMap[result] || 'default'}>{labelMap[result] || result}</Tag>;
      },
    },
    {
      title: '选中 API Key',
      key: 'selected_api_key',
      width: 200,
      render: (_: any, record: RouteDecisionLog) => {
        const decision = record.decision_layer_output;
        if (!decision) return '-';

        const selectedAPIKey = decision.selected_api_key_id || decision.api_key_id;
        const routeMode = decision.route_mode || decision.mode;

        return (
          <Space direction="vertical" size={2}>
            {selectedAPIKey && <div>API Key: {selectedAPIKey}</div>}
            {routeMode && <div>模式: {routeMode}</div>}
          </Space>
        );
      },
    },
    {
      title: '执行结果',
      key: 'execution_result',
      width: 120,
      render: (_: any, record: RouteDecisionLog) => {
        const execution = record.execution_layer_result;
        if (!execution) return '-';

        const status = execution.status || execution.success ? 'success' : 'error';

        return (
          <Badge
            status={status as 'success' | 'error' | 'warning' | 'default'}
            text={status === 'success' ? '成功' : '失败'}
          />
        );
      },
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (createdAt: string) => <span>{new Date(createdAt).toLocaleString('zh-CN')}</span>,
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_: any, record: RouteDecisionLog) => (
        <Button type="link" icon={<EyeOutlined />} onClick={() => handleViewDetail(record)}>
          详情
        </Button>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <ClockCircleOutlined />
          路由决策日志查询
        </Space>
      }
      extra={
        <Button icon={<ReloadOutlined spin={loading} />} onClick={handleRefresh} loading={loading}>
          刷新
        </Button>
      }
    >
      <Alert
        message="日志说明"
        description={
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            <li>路由决策日志记录了每次API请求的路由决策过程</li>
            <li>策略层目标：记录了本次路由的策略目标（如性能优先、价格优先等）</li>
            <li>决策层输出：记录了路由决策的详细信息，包括选中的API Key、路由模式等</li>
            <li>执行层结果：记录了路由执行的结果，包括成功/失败状态、响应时间等</li>
          </ul>
        }
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Card
        title={
          <Space>
            <FilterOutlined />
            筛选条件
          </Space>
        }
        size="small"
        style={{ marginBottom: 16 }}
      >
        <Space wrap size={16}>
          <Input
            allowClear
            placeholder="请求ID"
            style={{ width: 200 }}
            value={filters.request_id}
            onChange={(e) => handleFilterChange('request_id', e.target.value)}
          />
          <Input
            allowClear
            placeholder="商户ID"
            style={{ width: 120 }}
            value={filters.merchant_id}
            onChange={(e) => handleFilterChange('merchant_id', e.target.value)}
          />
          <Input
            allowClear
            placeholder="API Key ID"
            style={{ width: 120 }}
            value={filters.api_key_id}
            onChange={(e) => handleFilterChange('api_key_id', e.target.value)}
          />
          <Select
            allowClear
            placeholder="策略类型"
            style={{ width: 130 }}
            value={filters.strategy}
            onChange={(value) => handleFilterChange('strategy', value)}
            options={[
              { value: 'performance_first', label: '性能优先' },
              { value: 'price_first', label: '价格优先' },
              { value: 'reliability_first', label: '可靠性优先' },
              { value: 'security_first', label: '安全优先' },
              { value: 'balanced', label: '均衡策略' },
              { value: 'auto', label: '自动模式' },
            ]}
          />
          <Select
            allowClear
            placeholder="决策结果"
            style={{ width: 120 }}
            value={filters.decision_result}
            onChange={(value) => handleFilterChange('decision_result', value)}
            options={[
              { value: 'success', label: '成功' },
              { value: 'failed', label: '失败' },
              { value: 'fallback', label: '降级' },
            ]}
          />
          <RangePicker style={{ width: 280 }} onChange={handleTimeRangeChange} />
          <Button onClick={handleResetFilters}>重置</Button>
          <Button type="primary" onClick={handleRefresh}>
            <SearchOutlined />
            查询
          </Button>
        </Space>
      </Card>

      <Table
        columns={columns}
        dataSource={logs}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: handlePageChange,
        }}
        scroll={{ x: 'max-content' }}
      />

      <Modal
        title="路由决策日志详情"
        open={detailModalVisible}
        onCancel={() => setDetailModalVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailModalVisible(false)}>
            关闭
          </Button>,
        ]}
        width={900}
      >
        {selectedLog ? (
          <div>
            <Descriptions bordered column={2} style={{ marginBottom: 16 }}>
              <Descriptions.Item label="日志ID">{selectedLog.id}</Descriptions.Item>
              <Descriptions.Item label="请求ID">{selectedLog.request_id}</Descriptions.Item>
              <Descriptions.Item label="商户ID">{selectedLog.merchant_id}</Descriptions.Item>
              <Descriptions.Item label="API Key ID">{selectedLog.api_key_id}</Descriptions.Item>
              <Descriptions.Item label="策略目标">
                <Tag color="blue">{getStrategyLabel(selectedLog.strategy_layer_goal)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="决策耗时">
                <Tag color={selectedLog.decision_duration_ms > 10 ? 'orange' : 'green'}>
                  {selectedLog.decision_duration_ms ? `${selectedLog.decision_duration_ms}ms` : '-'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="决策结果">
                {(() => {
                  if (!selectedLog.decision_result) return '-';
                  const colorMap: Record<string, string> = {
                    success: 'green',
                    failed: 'red',
                    fallback: 'orange',
                  };
                  const labelMap: Record<string, string> = {
                    success: '成功',
                    failed: '失败',
                    fallback: '降级',
                  };
                  return (
                    <Tag color={colorMap[selectedLog.decision_result] || 'default'}>
                      {labelMap[selectedLog.decision_result] || selectedLog.decision_result}
                    </Tag>
                  );
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="创建时间">
                {new Date(selectedLog.created_at).toLocaleString('zh-CN')}
              </Descriptions.Item>
              {selectedLog.error_message && (
                <Descriptions.Item label="错误信息" span={2}>
                  <Alert type="error" message={selectedLog.error_message} />
                </Descriptions.Item>
              )}
            </Descriptions>

            <Card title="策略层" style={{ marginBottom: 16 }}>
              <Descriptions bordered column={2} size="small">
                <Descriptions.Item label="策略目标">
                  <Tag color="blue">{getStrategyLabel(selectedLog.strategy_layer_goal)}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="决策耗时">
                  <Tag color={selectedLog.decision_duration_ms > 10 ? 'orange' : 'green'}>
                    {selectedLog.decision_duration_ms
                      ? `${selectedLog.decision_duration_ms}ms`
                      : '-'}
                  </Tag>
                </Descriptions.Item>
              </Descriptions>
              {selectedLog.strategy_layer_input &&
                Object.keys(selectedLog.strategy_layer_input).length > 0 && (
                  <div style={{ marginTop: 12 }}>
                    <div style={{ fontWeight: 'bold', marginBottom: 8 }}>策略层输入：</div>
                    <pre
                      style={{
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-all',
                        background: '#f5f5f5',
                        padding: 12,
                        borderRadius: 4,
                      }}
                    >
                      {JSON.stringify(selectedLog.strategy_layer_input, null, 2)}
                    </pre>
                  </div>
                )}
              {selectedLog.strategy_layer_output &&
                Object.keys(selectedLog.strategy_layer_output).length > 0 && (
                  <div style={{ marginTop: 12 }}>
                    <div style={{ fontWeight: 'bold', marginBottom: 8 }}>策略层输出：</div>
                    <pre
                      style={{
                        whiteSpace: 'pre-wrap',
                        wordBreak: 'break-all',
                        background: '#f5f5f5',
                        padding: 12,
                        borderRadius: 4,
                      }}
                    >
                      {JSON.stringify(selectedLog.strategy_layer_output, null, 2)}
                    </pre>
                  </div>
                )}
            </Card>

            <Card title="决策层" style={{ marginBottom: 16 }}>
              {selectedLog.decision_layer_candidates &&
                selectedLog.decision_layer_candidates.length > 0 && (
                  <div style={{ marginBottom: 12 }}>
                    <div style={{ fontWeight: 'bold', marginBottom: 8 }}>
                      候选 API Keys ({selectedLog.decision_layer_candidates.length} 个)：
                    </div>
                    <Table
                      dataSource={selectedLog.decision_layer_candidates}
                      rowKey={(_, index) => `candidate-${index}`}
                      size="small"
                      pagination={false}
                      columns={[
                        {
                          title: 'API Key ID',
                          dataIndex: 'api_key_id',
                          key: 'api_key_id',
                          width: 100,
                        },
                        { title: '提供商', dataIndex: 'provider', key: 'provider', width: 100 },
                        {
                          title: '评分',
                          dataIndex: 'score',
                          key: 'score',
                          width: 80,
                          render: (v: number) => v?.toFixed(2) || '-',
                        },
                        {
                          title: '健康状态',
                          dataIndex: 'health_status',
                          key: 'health_status',
                          width: 100,
                        },
                        {
                          title: '选中',
                          dataIndex: 'selected',
                          key: 'selected',
                          width: 60,
                          render: (v: boolean) => (v ? <Tag color="green">是</Tag> : <Tag>否</Tag>),
                        },
                      ]}
                    />
                  </div>
                )}
              <div style={{ fontWeight: 'bold', marginBottom: 8 }}>决策层输出：</div>
              <pre
                style={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-all',
                  background: '#f5f5f5',
                  padding: 12,
                  borderRadius: 4,
                }}
              >
                {JSON.stringify(selectedLog.decision_layer_output, null, 2)}
              </pre>
            </Card>

            <Card title="执行层" style={{ marginBottom: 16 }}>
              <Descriptions bordered column={2} size="small">
                <Descriptions.Item label="决策结果">
                  {(() => {
                    if (!selectedLog.decision_result) return '-';
                    const colorMap: Record<string, string> = {
                      success: 'green',
                      failed: 'red',
                      fallback: 'orange',
                    };
                    const labelMap: Record<string, string> = {
                      success: '成功',
                      failed: '失败',
                      fallback: '降级',
                    };
                    return (
                      <Tag color={colorMap[selectedLog.decision_result] || 'default'}>
                        {labelMap[selectedLog.decision_result] || selectedLog.decision_result}
                      </Tag>
                    );
                  })()}
                </Descriptions.Item>
                <Descriptions.Item label="执行状态">
                  {selectedLog.execution_layer_result?.status ? (
                    <Badge
                      status={
                        selectedLog.execution_layer_result.status === 'success'
                          ? 'success'
                          : 'error'
                      }
                      text={
                        selectedLog.execution_layer_result.status === 'success' ? '成功' : '失败'
                      }
                    />
                  ) : (
                    '-'
                  )}
                </Descriptions.Item>
              </Descriptions>
              {selectedLog.error_message && (
                <Alert
                  type="error"
                  message="错误信息"
                  description={selectedLog.error_message}
                  style={{ marginTop: 12 }}
                  showIcon
                />
              )}
              <div style={{ marginTop: 12 }}>
                <div style={{ fontWeight: 'bold', marginBottom: 8 }}>执行层结果：</div>
                <pre
                  style={{
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-all',
                    background: '#f5f5f5',
                    padding: 12,
                    borderRadius: 4,
                  }}
                >
                  {JSON.stringify(selectedLog.execution_layer_result, null, 2)}
                </pre>
              </div>
            </Card>
          </div>
        ) : (
          <Spin size="large" style={{ textAlign: 'center', padding: '40px 0' }} />
        )}
      </Modal>
    </Card>
  );
};

export default AdminRouteDecisionLogs;
