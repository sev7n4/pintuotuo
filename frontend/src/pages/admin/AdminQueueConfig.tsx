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
import {
  OrderedListOutlined,
  ReloadOutlined,
  EditOutlined,
} from '@ant-design/icons';
import api from '@services/api';

const { Text } = Typography;

interface QueueItem {
  key: string;
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

const AdminQueueConfig: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState<QueueItem[]>([]);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [form] = Form.useForm();

  const fetchStats = async () => {
    setLoading(true);
    try {
      const response = await api.get<{ queues: Record<string, QueueItem> }>('/admin/queue/stats');
      const queues = response.data.queues || {};
      const data = Object.entries(queues).map(([key, stats]: [string, any]) => ({
        key,
        ...stats,
      }));
      setStats(data);
    } catch (error) {
      message.error('获取队列统计失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
  }, []);

  const handleEdit = (record: QueueItem) => {
    setEditingKey(record.key);
    form.setFieldsValue({ key: record.key, max_size: record.max_size });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      await api.put('/admin/queue/config', values);
      message.success('队列配置已更新');
      setModalVisible(false);
      fetchStats();
    } catch (error) {
      message.error('操作失败');
    }
  };

  const handleReset = async (key?: string) => {
    try {
      const url = key
        ? `/admin/queue/reset?key=${encodeURIComponent(key)}`
        : '/admin/queue/reset';
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
      title: '队列大小',
      key: 'size',
      render: (_: any, record: QueueItem) => (
        <Progress
          percent={Math.round((record.current_size / (record.max_size || 1)) * 100)}
          size="small"
          format={() => `${record.current_size}/${record.max_size}`}
          status={record.current_size > record.max_size * 0.8 ? 'exception' : 'normal'}
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
    {
      title: '最大等待',
      dataIndex: 'max_wait_time_ms',
      key: 'max_wait_time_ms',
      render: (val: number) => `${Math.round(val || 0)}ms`,
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: QueueItem) => (
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
            title="确定重置此队列的统计数据？"
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
          <OrderedListOutlined style={{ marginRight: 8 }} />
          队列配置
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
        message="队列说明"
        description="请求队列用于管理并发请求，按优先级排序处理。Max Size 表示队列最大容量，超过后新请求将被丢弃。"
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
        scroll={{ x: 1000 }}
      />

      <Modal
        title="编辑队列配置"
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="key"
            label="Key"
          >
            <Text code>{editingKey}</Text>
          </Form.Item>
          <Form.Item
            name="max_size"
            label="最大队列大小"
            rules={[{ required: true, message: '请输入最大队列大小' }]}
          >
            <InputNumber min={1} max={100000} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default AdminQueueConfig;
