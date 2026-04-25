import React, { useState, useEffect } from 'react';
import {
  Card,
  Row,
  Col,
  Statistic,
  Table,
  Spin,
  message,
  Button,
  Tag,
  Progress,
  Typography,
  Divider,
} from 'antd';
import {
  CloudServerOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ReloadOutlined,
  DollarOutlined,
} from '@ant-design/icons';
import api from '@services/api';

const { Text } = Typography;

interface GatewayStats {
  requests: number;
  successes: number;
  failures: number;
  success_rate: number;
  avg_latency_ms: number;
  max_latency_ms: number;
  total_tokens: number;
  total_cost: number;
  last_reset: string;
}

interface RateLimiterStats {
  requests: number;
  allowed: number;
  denied: number;
  denied_rate: number;
  last_reset: string;
}

interface QueueStats {
  current_size: number;
  max_size: number;
  enqueued: number;
  dequeued: number;
  expired: number;
  dropped: number;
  avg_wait_time_ms: number;
  max_wait_time_ms: number;
  last_reset: string;
}

const AdminGatewayStats: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [gatewayStats, setGatewayStats] = useState<GatewayStats | null>(null);
  const [rateLimiterStats, setRateLimiterStats] = useState<Record<string, RateLimiterStats>>({});
  const [queueStats, setQueueStats] = useState<Record<string, QueueStats>>({});

  const fetchStats = async () => {
    setLoading(true);
    try {
      const response = await api.get<{
        gateway: GatewayStats;
        rate_limiters: Record<string, RateLimiterStats>;
        queues: Record<string, QueueStats>;
      }>('/admin/gateway/stats');
      setGatewayStats(response.data.gateway);
      setRateLimiterStats(response.data.rate_limiters || {});
      setQueueStats(response.data.queues || {});
    } catch (error) {
      message.error('获取网关统计信息失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, []);

  const rateLimiterColumns = [
    {
      title: 'Key',
      dataIndex: 'key',
      key: 'key',
      render: (key: string) => <Text code>{key}</Text>,
    },
    {
      title: '请求数',
      dataIndex: 'requests',
      key: 'requests',
      render: (val: number) => val?.toLocaleString() || 0,
    },
    {
      title: '允许',
      dataIndex: 'allowed',
      key: 'allowed',
      render: (val: number) => <Tag color="green">{val?.toLocaleString() || 0}</Tag>,
    },
    {
      title: '拒绝',
      dataIndex: 'denied',
      key: 'denied',
      render: (val: number) => <Tag color="red">{val?.toLocaleString() || 0}</Tag>,
    },
    {
      title: '拒绝率',
      dataIndex: 'denied_rate',
      key: 'denied_rate',
      render: (val: number) => (
        <Progress
          percent={Math.round((val || 0) * 100)}
          size="small"
          status={val > 0.1 ? 'exception' : 'normal'}
          format={(percent) => `${percent}%`}
        />
      ),
    },
  ];

  const queueColumns = [
    {
      title: 'Key',
      dataIndex: 'key',
      key: 'key',
      render: (key: string) => <Text code>{key}</Text>,
    },
    {
      title: '当前大小',
      dataIndex: 'current_size',
      key: 'current_size',
      render: (val: number, record: QueueStats & { key: string }) => (
        <Progress
          percent={Math.round((val / (record.max_size || 1)) * 100)}
          size="small"
          format={() => `${val}/${record.max_size}`}
        />
      ),
    },
    {
      title: '入队',
      dataIndex: 'enqueued',
      key: 'enqueued',
      render: (val: number) => val?.toLocaleString() || 0,
    },
    {
      title: '出队',
      dataIndex: 'dequeued',
      key: 'dequeued',
      render: (val: number) => val?.toLocaleString() || 0,
    },
    {
      title: '过期',
      dataIndex: 'expired',
      key: 'expired',
      render: (val: number) => (val > 0 ? <Tag color="orange">{val}</Tag> : 0),
    },
    {
      title: '丢弃',
      dataIndex: 'dropped',
      key: 'dropped',
      render: (val: number) => (val > 0 ? <Tag color="red">{val}</Tag> : 0),
    },
    {
      title: '平均等待',
      dataIndex: 'avg_wait_time_ms',
      key: 'avg_wait_time_ms',
      render: (val: number) => `${Math.round(val || 0)}ms`,
    },
  ];

  const rateLimiterData = Object.entries(rateLimiterStats).map(([key, stats]) => ({
    key,
    ...stats,
  }));

  const queueData = Object.entries(queueStats).map(([key, stats]) => ({
    key,
    ...stats,
  }));

  return (
    <div>
      <Card
        title={
          <span>
            <CloudServerOutlined style={{ marginRight: 8 }} />
            网关统计
          </span>
        }
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchStats} loading={loading}>
            刷新
          </Button>
        }
      >
        <Spin spinning={loading}>
          <Row gutter={[16, 16]}>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="总请求数"
                  value={gatewayStats?.requests || 0}
                  prefix={<CloudServerOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="成功数"
                  value={gatewayStats?.successes || 0}
                  valueStyle={{ color: '#3f8600' }}
                  prefix={<CheckCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="失败数"
                  value={gatewayStats?.failures || 0}
                  valueStyle={{ color: '#cf1322' }}
                  prefix={<CloseCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="成功率"
                  value={Math.round((gatewayStats?.success_rate || 0) * 100)}
                  suffix="%"
                  valueStyle={{
                    color: (gatewayStats?.success_rate || 0) > 0.9 ? '#3f8600' : '#cf1322',
                  }}
                />
              </Card>
            </Col>
          </Row>

          <Divider />

          <Row gutter={[16, 16]}>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="平均延迟"
                  value={Math.round(gatewayStats?.avg_latency_ms || 0)}
                  suffix="ms"
                  prefix={<ClockCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="最大延迟"
                  value={Math.round(gatewayStats?.max_latency_ms || 0)}
                  suffix="ms"
                />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic title="总 Token" value={gatewayStats?.total_tokens || 0} />
              </Card>
            </Col>
            <Col xs={12} sm={12} md={6}>
              <Card>
                <Statistic
                  title="总成本"
                  value={gatewayStats?.total_cost || 0}
                  precision={4}
                  prefix={<DollarOutlined />}
                />
              </Card>
            </Col>
          </Row>
        </Spin>
      </Card>

      <Card title="限流器统计" style={{ marginTop: 16 }}>
        <Table
          columns={rateLimiterColumns}
          dataSource={rateLimiterData}
          rowKey="key"
          pagination={false}
          scroll={{ x: 600 }}
          locale={{ emptyText: '暂无限流器数据' }}
        />
      </Card>

      <Card title="队列统计" style={{ marginTop: 16 }}>
        <Table
          columns={queueColumns}
          dataSource={queueData}
          rowKey="key"
          pagination={false}
          scroll={{ x: 800 }}
          locale={{ emptyText: '暂无队列数据' }}
        />
      </Card>
    </div>
  );
};

export default AdminGatewayStats;
