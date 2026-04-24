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
  Row,
  Col,
  Grid,
  Tooltip,
  Alert,
  Divider,
  Select,
} from 'antd';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  StarOutlined,
  StarFilled,
} from '@ant-design/icons';
import api from '@services/api';

const { useBreakpoint } = Grid;

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

const STRATEGY_PRESETS: Record<
  string,
  {
    weights: { price: number; latency: number; reliability: number };
    retry: { count: number; backoff: number };
    circuitBreaker: { threshold: number; timeout: number };
    description: string;
  }
> = {
  price_first: {
    weights: { price: 60, latency: 20, reliability: 20 },
    retry: { count: 3, backoff: 1000 },
    circuitBreaker: { threshold: 5, timeout: 60 },
    description: '优先选择价格最低的Provider，适合成本敏感场景',
  },
  latency_first: {
    weights: { price: 20, latency: 60, reliability: 20 },
    retry: { count: 2, backoff: 500 },
    circuitBreaker: { threshold: 3, timeout: 30 },
    description: '优先选择延迟最低的Provider，适合实时性要求高的场景',
  },
  reliability_first: {
    weights: { price: 20, latency: 20, reliability: 60 },
    retry: { count: 5, backoff: 2000 },
    circuitBreaker: { threshold: 3, timeout: 120 },
    description: '优先选择最可靠的Provider，适合稳定性要求高的场景',
  },
  balanced: {
    weights: { price: 33, latency: 34, reliability: 33 },
    retry: { count: 3, backoff: 1000 },
    circuitBreaker: { threshold: 5, timeout: 60 },
    description: '均衡考虑价格、延迟和可靠性，适合通用场景',
  },
};

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
  const screens = useBreakpoint();

  const [weightValues, setWeightValues] = useState({
    price: 33,
    latency: 34,
    reliability: 33,
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
    const defaultWeights = { price: 33, latency: 34, reliability: 33 };
    setWeightValues(defaultWeights);
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
    setWeightValues({
      price: Math.round(strategy.price_weight * 100),
      latency: Math.round(strategy.latency_weight * 100),
      reliability: Math.round(strategy.reliability_weight * 100),
    });
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

  const handleSetDefault = async (id: number) => {
    try {
      const { data } = await api.get<{ strategy: RoutingStrategyConfig }>(
        `/admin/routing-strategies/${id}`
      );
      const s = data.strategy;
      await api.put(`/admin/routing-strategies/${id}`, { ...s, is_default: true });
      message.success('已设为默认策略');
      fetchStrategies();
    } catch (error) {
      message.error('设置失败');
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

  const handleWeightChange = (type: 'price' | 'latency' | 'reliability', value: number) => {
    const newValues = { ...weightValues, [type]: value };
    setWeightValues(newValues);
    form.setFieldsValue({
      price_weight: newValues.price / 100,
      latency_weight: newValues.latency / 100,
      reliability_weight: newValues.reliability / 100,
    });
  };

  const applyPreset = (presetKey: string) => {
    const preset = STRATEGY_PRESETS[presetKey];
    if (preset) {
      setWeightValues(preset.weights);
      form.setFieldsValue({
        price_weight: preset.weights.price / 100,
        latency_weight: preset.weights.latency / 100,
        reliability_weight: preset.weights.reliability / 100,
        max_retry_count: preset.retry.count,
        retry_backoff_base: preset.retry.backoff,
        circuit_breaker_threshold: preset.circuitBreaker.threshold,
        circuit_breaker_timeout: preset.circuitBreaker.timeout,
      });
    }
  };

  const isMobile = !screens.md;

  const mobileCard = (record: RoutingStrategyConfig) => (
    <Card
      size="small"
      style={{ marginBottom: 12 }}
      title={
        <Space>
          <span>{record.name}</span>
          {record.is_default && (
            <Tag color="blue" icon={<StarFilled />}>
              默认
            </Tag>
          )}
          <Tag color={record.status === 'active' ? 'green' : 'default'}>
            {record.status === 'active' ? '启用' : '禁用'}
          </Tag>
        </Space>
      }
      extra={
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          />
          {!record.is_default && (
            <Tooltip title="设为默认">
              <Button
                type="link"
                size="small"
                icon={<StarOutlined />}
                onClick={() => handleSetDefault(record.id)}
              />
            </Tooltip>
          )}
          <Popconfirm
            title="确定要删除此策略吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      }
    >
      <Row gutter={[8, 8]}>
        <Col span={12}>
          <div style={{ fontSize: 12, color: '#666' }}>编码</div>
          <div>{record.code}</div>
        </Col>
        <Col span={12}>
          <div style={{ fontSize: 12, color: '#666' }}>描述</div>
          <div style={{ fontSize: 12 }}>{record.description || '-'}</div>
        </Col>
        <Col span={24}>
          <div style={{ fontSize: 12, color: '#666', marginBottom: 4 }}>权重配置</div>
          <Space size={4}>
            <Tag>价格 {Math.round(record.price_weight * 100)}%</Tag>
            <Tag>延迟 {Math.round(record.latency_weight * 100)}%</Tag>
            <Tag>可靠 {Math.round(record.reliability_weight * 100)}%</Tag>
          </Space>
        </Col>
        <Col span={12}>
          <div style={{ fontSize: 12, color: '#666' }}>重试配置</div>
          <div style={{ fontSize: 12 }}>
            最大{record.max_retry_count}次 / 退避{record.retry_backoff_base}ms
          </div>
        </Col>
        <Col span={12}>
          <div style={{ fontSize: 12, color: '#666' }}>熔断器</div>
          <div style={{ fontSize: 12 }}>
            阈值{record.circuit_breaker_threshold}次 / 超时{record.circuit_breaker_timeout}s
          </div>
        </Col>
      </Row>
    </Card>
  );

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
      responsive: ['md'] as any,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: RoutingStrategyConfig) => (
        <Space>
          {name}
          {record.is_default && (
            <Tag color="blue" icon={<StarFilled />}>
              默认
            </Tag>
          )}
        </Space>
      ),
    },
    {
      title: '编码',
      dataIndex: 'code',
      key: 'code',
      responsive: ['lg'] as any,
    },
    {
      title: '权重配置',
      key: 'weights',
      responsive: ['lg'] as any,
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space size={4}>
          <Tag>价格 {Math.round(record.price_weight * 100)}%</Tag>
          <Tag>延迟 {Math.round(record.latency_weight * 100)}%</Tag>
          <Tag>可靠 {Math.round(record.reliability_weight * 100)}%</Tag>
        </Space>
      ),
    },
    {
      title: '重试/熔断',
      key: 'config',
      responsive: ['xl'] as any,
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space direction="vertical" size={0}>
          <span style={{ fontSize: 12 }}>
            重试: {record.max_retry_count}次/{record.retry_backoff_base}ms
          </span>
          <span style={{ fontSize: 12 }}>
            熔断: {record.circuit_breaker_threshold}次/{record.circuit_breaker_timeout}s
          </span>
        </Space>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : status === 'inactive' ? '禁用' : status || '-'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      responsive: ['xl'] as any,
      render: (createdAt: string) => (
        <span style={{ fontSize: 12 }}>
          {createdAt ? new Date(createdAt).toLocaleString('zh-CN') : '-'}
        </span>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 160,
      responsive: ['xl'] as any,
      render: (updatedAt: string) => (
        <span style={{ fontSize: 12 }}>
          {updatedAt ? new Date(updatedAt).toLocaleString('zh-CN') : '-'}
        </span>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: isMobile ? 80 : 200,
      fixed: isMobile ? ('right' as const) : undefined,
      render: (_: any, record: RoutingStrategyConfig) => (
        <Space size={isMobile ? 0 : 8}>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            {!isMobile && '编辑'}
          </Button>
          {!record.is_default && (
            <Tooltip title="设为默认">
              <Button
                type="link"
                size="small"
                icon={<StarOutlined />}
                onClick={() => handleSetDefault(record.id)}
              >
                {!isMobile && '设为默认'}
              </Button>
            </Tooltip>
          )}
          <Popconfirm
            title="确定要删除此策略吗？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {!isMobile && '删除'}
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
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          新建策略
        </Button>
      }
    >
      <Alert
        message="策略说明"
        description="默认策略是系统全局使用的路由策略。每个策略通过权重配置决定Provider选择偏好，重试和熔断器配置控制故障处理行为。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      {isMobile ? (
        <div>
          {strategies.map((s) => mobileCard(s))}
          <div style={{ textAlign: 'center', marginTop: 16 }}>
            <Space>
              <Button
                disabled={pagination.current <= 1}
                onClick={() => setPagination({ ...pagination, current: pagination.current - 1 })}
              >
                上一页
              </Button>
              <span>
                {pagination.current} / {Math.ceil(pagination.total / pagination.pageSize)}
              </span>
              <Button
                disabled={pagination.current >= Math.ceil(pagination.total / pagination.pageSize)}
                onClick={() => setPagination({ ...pagination, current: pagination.current + 1 })}
              >
                下一页
              </Button>
            </Space>
          </div>
        </div>
      ) : (
        <Table
          columns={columns}
          dataSource={strategies}
          rowKey="id"
          loading={loading}
          scroll={{ x: 900 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      )}

      <Modal
        title={editingStrategy ? '编辑策略' : '新建策略'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="策略名称"
                rules={[{ required: true, message: '请输入策略名称' }]}
              >
                <Input placeholder="请输入策略名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="code"
                label="策略编码"
                rules={[{ required: true, message: '请输入策略编码' }]}
              >
                <Input placeholder="请输入策略编码" disabled={!!editingStrategy} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="description" label="策略描述">
            <Input.TextArea rows={2} placeholder="请输入策略描述" />
          </Form.Item>

          <Divider orientation="left">
            <Space>
              权重配置
              <Tooltip title="权重总和应为100%，决定Provider选择的偏好程度">
                <InfoCircleOutlined style={{ color: '#1890ff' }} />
              </Tooltip>
            </Space>
          </Divider>

          <Alert
            message="预设策略模板"
            description={
              <Space wrap>
                {Object.keys(STRATEGY_PRESETS).map((key) => (
                  <Button key={key} size="small" onClick={() => applyPreset(key)}>
                    {key === 'price_first'
                      ? '价格优先'
                      : key === 'latency_first'
                        ? '延迟优先'
                        : key === 'reliability_first'
                          ? '可靠性优先'
                          : '均衡策略'}
                  </Button>
                ))}
              </Space>
            }
            type="info"
            style={{ marginBottom: 16 }}
          />

          <Row gutter={16}>
            <Col span={24}>
              <Form.Item
                label={
                  <Space>
                    价格权重
                    <Tooltip title="价格优先策略建议: 60%">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <Row gutter={16} align="middle">
                  <Col flex="auto">
                    <Slider
                      min={0}
                      max={100}
                      value={weightValues.price}
                      onChange={(v) => handleWeightChange('price', v)}
                    />
                  </Col>
                  <Col span={4}>
                    <InputNumber
                      min={0}
                      max={100}
                      value={weightValues.price}
                      onChange={(v) => handleWeightChange('price', v || 0)}
                      formatter={(v) => `${v}%`}
                      parser={(v) => Number(v?.replace('%', '')) || 0}
                    />
                  </Col>
                </Row>
                <Form.Item name="price_weight" noStyle>
                  <input type="hidden" />
                </Form.Item>
              </Form.Item>
            </Col>

            <Col span={24}>
              <Form.Item
                label={
                  <Space>
                    延迟权重
                    <Tooltip title="延迟优先策略建议: 60%">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <Row gutter={16} align="middle">
                  <Col flex="auto">
                    <Slider
                      min={0}
                      max={100}
                      value={weightValues.latency}
                      onChange={(v) => handleWeightChange('latency', v)}
                    />
                  </Col>
                  <Col span={4}>
                    <InputNumber
                      min={0}
                      max={100}
                      value={weightValues.latency}
                      onChange={(v) => handleWeightChange('latency', v || 0)}
                      formatter={(v) => `${v}%`}
                      parser={(v) => Number(v?.replace('%', '')) || 0}
                    />
                  </Col>
                </Row>
                <Form.Item name="latency_weight" noStyle>
                  <input type="hidden" />
                </Form.Item>
              </Form.Item>
            </Col>

            <Col span={24}>
              <Form.Item
                label={
                  <Space>
                    可靠性权重
                    <Tooltip title="可靠性优先策略建议: 60%">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <Row gutter={16} align="middle">
                  <Col flex="auto">
                    <Slider
                      min={0}
                      max={100}
                      value={weightValues.reliability}
                      onChange={(v) => handleWeightChange('reliability', v)}
                    />
                  </Col>
                  <Col span={4}>
                    <InputNumber
                      min={0}
                      max={100}
                      value={weightValues.reliability}
                      onChange={(v) => handleWeightChange('reliability', v || 0)}
                      formatter={(v) => `${v}%`}
                      parser={(v) => Number(v?.replace('%', '')) || 0}
                    />
                  </Col>
                </Row>
                <Form.Item name="reliability_weight" noStyle>
                  <input type="hidden" />
                </Form.Item>
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left">
            <Space>
              重试配置
              <Tooltip title="控制请求失败后的重试行为">
                <InfoCircleOutlined style={{ color: '#1890ff' }} />
              </Tooltip>
            </Space>
          </Divider>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="max_retry_count"
                label={
                  <Space>
                    最大重试次数
                    <Tooltip title="建议: 均衡3次 / 延迟优先2次 / 可靠性优先5次">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <InputNumber min={0} max={10} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="retry_backoff_base"
                label={
                  <Space>
                    退避基数(ms)
                    <Tooltip title="建议: 延迟优先500ms / 均衡1000ms / 可靠性优先2000ms">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <InputNumber min={100} max={10000} step={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left">
            <Space>
              熔断器配置
              <Tooltip title="控制故障Provider的熔断行为">
                <InfoCircleOutlined style={{ color: '#1890ff' }} />
              </Tooltip>
            </Space>
          </Divider>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="circuit_breaker_threshold"
                label={
                  <Space>
                    熔断阈值(次)
                    <Tooltip title="连续失败多少次后熔断。建议: 延迟优先3次 / 均衡5次 / 可靠性优先3次">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <InputNumber min={1} max={20} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="circuit_breaker_timeout"
                label={
                  <Space>
                    熔断超时(s)
                    <Tooltip title="熔断后多久尝试恢复。建议: 延迟优先30s / 均衡60s / 可靠性优先120s">
                      <InfoCircleOutlined style={{ color: '#999', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
              >
                <InputNumber min={10} max={300} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Divider />

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="is_default"
                label={
                  <Space>
                    设为默认策略
                    <Tooltip title="设为默认后，系统将使用此策略进行路由选择">
                      <InfoCircleOutlined style={{ color: '#1890ff' }} />
                    </Tooltip>
                  </Space>
                }
                valuePropName="checked"
              >
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="status"
                label="状态"
                rules={[{ required: true, message: '请选择启用或禁用' }]}
              >
                <Select
                  placeholder="启用或禁用"
                  options={[
                    { value: 'active', label: '启用' },
                    { value: 'inactive', label: '禁用' },
                  ]}
                />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </Card>
  );
};

export default AdminRoutingStrategies;
