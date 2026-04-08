import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  Modal,
  Form,
  Input,
  InputNumber,
  Switch,
  message,
  Tag,
  Popconfirm,
  Slider,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
} from '@ant-design/icons';
import { api } from '@services/api';

interface RoutingStrategyConfig {
  id: number;
  name: string;
  code: string;
  description: string | null;
  price_weight: number;
  latency_weight: number;
  reliability_weight: number;
  max_retry_count: number;
  retry_backoff_base: number;
  circuit_breaker_threshold: number;
  circuit_breaker_timeout: number;
  is_default: boolean;
  status: string;
  created_at: string;
  updated_at: string;
}

const AdminRoutingStrategies: React.FC = () => {
  const [strategies, setStrategies] = useState<RoutingStrategyConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingStrategy, setEditingStrategy] = useState<RoutingStrategyConfig | null>(null);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  useEffect(() => {
    fetchStrategies();
  }, [pagination.current, pagination.pageSize]);

  const fetchStrategies = async () => {
    setLoading(true);
    try {
      const response = await api.get<{ strategies: RoutingStrategyConfig[]; total: number }>(
        `/admin/routing-strategies?page=${pagination.current}&page_size=${pagination.pageSize}`
      );
      setStrategies(response.data.strategies || []);
      setPagination({ ...pagination, total: response.data.total });
    } catch (error) {
      message.error('获取路由策略列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingStrategy(null);
    form.resetFields();
    form.setFieldsValue({
      price_weight: 0.33,
      latency_weight: 0.34,
      reliability_weight: 0.33,
      max_retry_count: 3,
      retry_backoff_base: 1000,
      circuit_breaker_threshold: 5,
      circuit_breaker_timeout: 60,
      is_default: false,
      status: 'active',
    });
    setModalVisible(true);
  };

  const handleEdit = (strategy: RoutingStrategyConfig) => {
    setEditingStrategy(strategy);
    form.setFieldsValue(strategy);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await api.delete(`/admin/routing-strategies/${id}`);
      message.success('删除成功');
      fetchStrategies();
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingStrategy) {
        await api.put(`/admin/routing-strategies/${editingStrategy.id}`, values);
        message.success('更新成功');
      } else {
        await api.post('/admin/routing-strategies', values);
        message.success('创建成功');
      }
      setModalVisible(false);
      fetchStrategies();
    } catch (error) {
      message.error(editingStrategy ? '更新失败' : '创建失败');
    }
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '编码',
      dataIndex: 'code',
      key: 'code',
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: '权重配置',
      key: 'weights',
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space direction="vertical" size="small">
          <span>价格: {(record.price_weight * 100).toFixed(0)}%</span>
          <span>延迟: {(record.latency_weight * 100).toFixed(0)}%</span>
          <span>可靠性: {(record.reliability_weight * 100).toFixed(0)}%</span>
        </Space>
      ),
    },
    {
      title: '重试配置',
      key: 'retry',
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space direction="vertical" size="small">
          <span>最大重试: {record.max_retry_count}次</span>
          <span>退避基数: {record.retry_backoff_base}ms</span>
        </Space>
      ),
    },
    {
      title: '熔断器',
      key: 'circuit_breaker',
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space direction="vertical" size="small">
          <span>阈值: {record.circuit_breaker_threshold}次</span>
          <span>超时: {record.circuit_breaker_timeout}s</span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string, record: RoutingStrategyConfig) => (
        <Space>
          <Tag color={status === 'active' ? 'green' : 'default'}>
            {status === 'active' ? '启用' : '禁用'}
          </Tag>
          {record.is_default && <Tag color="blue">默认</Tag>}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除此策略吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="路由策略管理"
      extra={
        <Button
          type="primary"
          icon={<PlusOutlined />}
          onClick={handleCreate}
        >
          新建策略
        </Button>
      }
    >
      <Table
        columns={columns}
        dataSource={strategies}
        rowKey="id"
        loading={loading}
        pagination={{
          ...pagination,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, pageSize) =>
            setPagination({ ...pagination, current: page, pageSize }),
        }}
      />

      <Modal
        title={editingStrategy ? '编辑策略' : '新建策略'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="name"
            label="策略名称"
            rules={[{ required: true, message: '请输入策略名称' }]}
          >
            <Input placeholder="请输入策略名称" />
          </Form.Item>

          <Form.Item
            name="code"
            label="策略编码"
            rules={[{ required: true, message: '请输入策略编码' }]}
          >
            <Input placeholder="请输入策略编码" disabled={!!editingStrategy} />
          </Form.Item>

          <Form.Item name="description" label="策略描述">
            <Input.TextArea rows={3} placeholder="请输入策略描述" />
          </Form.Item>

          <Form.Item label="权重配置（总和应为100%）">
            <Space direction="vertical" style={{ width: '100%' }}>
              <Form.Item name="price_weight" noStyle>
                <div>
                  <span>价格权重: </span>
                  <Slider
                    min={0}
                    max={100}
                    value={form.getFieldValue('price_weight') * 100}
                    onChange={(value) => form.setFieldsValue({ price_weight: value / 100 })}
                    style={{ width: 300, display: 'inline-block' }}
                  />
                  <span> {(form.getFieldValue('price_weight') * 100).toFixed(0)}%</span>
                </div>
              </Form.Item>
              <Form.Item name="latency_weight" noStyle>
                <div>
                  <span>延迟权重: </span>
                  <Slider
                    min={0}
                    max={100}
                    value={form.getFieldValue('latency_weight') * 100}
                    onChange={(value) => form.setFieldsValue({ latency_weight: value / 100 })}
                    style={{ width: 300, display: 'inline-block' }}
                  />
                  <span> {(form.getFieldValue('latency_weight') * 100).toFixed(0)}%</span>
                </div>
              </Form.Item>
              <Form.Item name="reliability_weight" noStyle>
                <div>
                  <span>可靠性权重: </span>
                  <Slider
                    min={0}
                    max={100}
                    value={form.getFieldValue('reliability_weight') * 100}
                    onChange={(value) => form.setFieldsValue({ reliability_weight: value / 100 })}
                    style={{ width: 300, display: 'inline-block' }}
                  />
                  <span> {(form.getFieldValue('reliability_weight') * 100).toFixed(0)}%</span>
                </div>
              </Form.Item>
            </Space>
          </Form.Item>

          <Form.Item label="重试配置">
            <Space>
              <Form.Item name="max_retry_count" noStyle>
                <div>
                  <span>最大重试次数: </span>
                  <InputNumber min={0} max={10} />
                </div>
              </Form.Item>
              <Form.Item name="retry_backoff_base" noStyle>
                <div>
                  <span>退避基数(ms): </span>
                  <InputNumber min={100} max={10000} step={100} />
                </div>
              </Form.Item>
            </Space>
          </Form.Item>

          <Form.Item label="熔断器配置">
            <Space>
              <Form.Item name="circuit_breaker_threshold" noStyle>
                <div>
                  <span>阈值(次): </span>
                  <InputNumber min={1} max={20} />
                </div>
              </Form.Item>
              <Form.Item name="circuit_breaker_timeout" noStyle>
                <div>
                  <span>超时时间(s): </span>
                  <InputNumber min={10} max={300} />
                </div>
              </Form.Item>
            </Space>
          </Form.Item>

          <Form.Item name="is_default" label="设为默认策略" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item name="status" label="状态">
            <Input disabled />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default AdminRoutingStrategies;
