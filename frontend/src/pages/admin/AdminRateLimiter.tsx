import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  InputNumber,
  message,
  Tag,
  Space,
  Popconfirm,
  Progress,
  Typography,
  Alert,
} from 'antd';
import { SafetyOutlined, ReloadOutlined, EditOutlined } from '@ant-design/icons';
import api from '@services/api';

const { Text } = Typography;

interface RateLimiterItem {
  key: string;
  requests: number;
  allowed: number;
  denied: number;
  denied_rate: number;
  last_reset: string;
}

const AdminRateLimiter: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<RateLimiterItem[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [form] = Form.useForm();

  const fetchStats = async () => {
    setLoading(true);
    try {
      const response = await api.get<{ limiters: Record<string, RateLimiterItem> }>(
        '/admin/rate-limiter/stats'
      );
      const limiters = response.data.limiters || {};
      const data = Object.entries(limiters).map(([key, stats]: [string, any]) => ({
        key,
        ...stats,
      }));
      setStats(data);
    } catch (error) {
      message.error('获取限流器统计失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const handleEdit = (record: RateLimiterItem) => {
    setEditingKey(record.key);
    form.setFieldsValue({ key: record.key, rate: 100, burst: 200 });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await api.put('/admin/rate-limiter/config', values);
      message.success(editingKey ? '限流器配置已更新' : '限流器配置已创建');
      setModalVisible(false);
      fetchStats();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleReset = async (key?: string) => {
    try {
      const url = key
        ? `/admin/rate-limiter/reset?key=${encodeURIComponent(key)}`
        : '/admin/rate-limiter/reset';
      await api.post(url);
      message.success('统计数据已重置');
      fetchStats();
    } catch (error) {
      message.error('重置失败');
    }
  };

  const columns = [
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
      render: (val: number) => (
        <Tag color={val > 0 ? 'red' : 'default'}>{val?.toLocaleString() || 0}</Tag>
      ),
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
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: RateLimiterItem) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            配置
          </Button>
          <Popconfirm
            title="确定重置此限流器的统计数据？"
            onConfirm={() => handleReset(record.key)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" icon={<ReloadOutlined />}>
              重置
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={
        <span>
          <SafetyOutlined style={{ marginRight: 8 }} />
          限流器配置
        </span>
      }
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={fetchStats} loading={loading}>
            刷新
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => handleReset()}>
            重置全部
          </Button>
        </Space>
      }
    >
      <Alert
        message="限流器说明"
        description="限流器使用令牌桶算法控制请求速率。Rate 表示每秒生成的令牌数，Burst 表示桶容量（最大突发请求数）。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Table
        columns={columns}
        dataSource={stats}
        rowKey="key"
        loading={loading}
        pagination={false}
        scroll={{ x: 800 }}
      />

      <Modal
        title={editingKey ? '编辑限流器配置' : '新建限流器配置'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item name="key" label="Key" rules={[{ required: true, message: '请输入 Key' }]}>
            <Text code>{editingKey || '新建'}</Text>
          </Form.Item>
          <Form.Item
            name="rate"
            label="Rate (每秒令牌数)"
            rules={[{ required: true, message: '请输入 Rate' }]}
          >
            <InputNumber min={1} max={10000} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="burst"
            label="Burst (桶容量)"
            rules={[{ required: true, message: '请输入 Burst' }]}
          >
            <InputNumber min={1} max={10000} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default AdminRateLimiter;
