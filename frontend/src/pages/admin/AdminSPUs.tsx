import { useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Space,
  Modal,
  Form,
  Input,
  Select,
  InputNumber,
  message,
  Popconfirm,
  Row,
  Col,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { SPU, ModelProvider, SPUCreateRequest } from '@/types/sku';
import { MODEL_TIER_LABELS } from '@/types/sku';

const AdminSPUs = () => {
  const [spus, setSPUs] = useState<SPU[]>([]);
  const [providers, setProviders] = useState<ModelProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSPU, setEditingSPU] = useState<SPU | null>(null);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState({
    provider: '',
    tier: '',
    status: 'active',
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  useEffect(() => {
    fetchSPUs();
    fetchProviders();
  }, [filters, pagination.current, pagination.pageSize]);

  const fetchSPUs = async () => {
    setLoading(true);
    try {
      const response = await skuService.getSPUs({
        page: pagination.current,
        per_page: pagination.pageSize,
        ...filters,
      });
      setSPUs(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch {
      message.error('获取SPU列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchProviders = async () => {
    try {
      const response = await skuService.getModelProviders();
      setProviders(response.data.data || []);
    } catch {
      console.error('Failed to fetch providers');
    }
  };

  const handleAdd = () => {
    setEditingSPU(null);
    form.resetFields();
    form.setFieldsValue({
      base_compute_points: 1.0,
      status: 'active',
      sort_order: 0,
    });
    setModalVisible(true);
  };

  const handleEdit = (record: SPU) => {
    setEditingSPU(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await skuService.deleteSPU(id);
      message.success('SPU已删除');
      fetchSPUs();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const data: SPUCreateRequest = {
        ...values,
        spu_code: values.spu_code.toUpperCase(),
      };

      if (editingSPU) {
        await skuService.updateSPU(editingSPU.id, data);
        message.success('SPU已更新');
      } else {
        await skuService.createSPU(data);
        message.success('SPU已创建');
      }
      setModalVisible(false);
      fetchSPUs();
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return;
      }
      message.error(editingSPU ? '更新失败' : '创建失败');
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
      title: 'SPU编码',
      dataIndex: 'spu_code',
      key: 'spu_code',
      width: 150,
    },
    {
      title: '产品名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '模型厂商',
      dataIndex: 'model_provider',
      key: 'model_provider',
      width: 120,
    },
    {
      title: '模型层级',
      dataIndex: 'model_tier',
      key: 'model_tier',
      width: 100,
      render: (tier: string) => (
        <Tag color={tier === 'pro' ? 'red' : tier === 'lite' ? 'blue' : tier === 'mini' ? 'green' : 'purple'}>
          {MODEL_TIER_LABELS[tier] || tier}
        </Tag>
      ),
    },
    {
      title: '算力点系数',
      dataIndex: 'base_compute_points',
      key: 'base_compute_points',
      width: 100,
      render: (v: number) => v.toFixed(4),
    },
    {
      title: '销量',
      dataIndex: 'total_sales_count',
      key: 'total_sales_count',
      width: 80,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: SPU) => (
        <Space size="small">
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个SPU吗？关联的SKU也会被删除。"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <Row gutter={16} style={{ marginBottom: 16 }}>
        <Col flex="auto">
          <h2 style={{ margin: 0 }}>SPU 管理</h2>
        </Col>
        <Col>
          <Space>
            <Select
              value={filters.provider}
              onChange={(v) => setFilters({ ...filters, provider: v })}
              style={{ width: 120 }}
              placeholder="厂商"
              allowClear
            >
              {providers.map((p) => (
                <Select.Option key={p.code} value={p.code}>
                  {p.name}
                </Select.Option>
              ))}
            </Select>
            <Select
              value={filters.tier}
              onChange={(v) => setFilters({ ...filters, tier: v })}
              style={{ width: 120 }}
              placeholder="层级"
              allowClear
            >
              <Select.Option value="pro">旗舰版</Select.Option>
              <Select.Option value="lite">标准版</Select.Option>
              <Select.Option value="mini">轻量版</Select.Option>
              <Select.Option value="vision">多模态版</Select.Option>
            </Select>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              新增SPU
            </Button>
          </Space>
        </Col>
      </Row>

      <Card>
        <Table
          columns={columns}
          dataSource={spus}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1100 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={editingSPU ? '编辑SPU' : '新增SPU'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={700}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="spu_code"
                label="SPU编码"
                rules={[{ required: true, message: '请输入SPU编码' }]}
                extra="唯一标识，如: DEEPSEEK-V3"
              >
                <Input placeholder="DEEPSEEK-V3" disabled={!!editingSPU} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="name"
                label="产品名称"
                rules={[{ required: true, message: '请输入产品名称' }]}
              >
                <Input placeholder="DeepSeek V3 模型服务" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="model_provider"
                label="模型厂商"
                rules={[{ required: true, message: '请选择模型厂商' }]}
              >
                <Select placeholder="请选择">
                  {providers.map((p) => (
                    <Select.Option key={p.code} value={p.code}>
                      {p.name}
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="model_name"
                label="模型名称"
                rules={[{ required: true, message: '请输入模型名称' }]}
                extra="API调用时使用的模型标识"
              >
                <Input placeholder="deepseek-chat" />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item name="model_version" label="模型版本">
                <Input placeholder="v3" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="model_tier"
                label="模型层级"
                rules={[{ required: true, message: '请选择模型层级' }]}
              >
                <Select placeholder="请选择">
                  <Select.Option value="pro">旗舰版 (Pro)</Select.Option>
                  <Select.Option value="lite">标准版 (Lite)</Select.Option>
                  <Select.Option value="mini">轻量版 (Mini)</Select.Option>
                  <Select.Option value="vision">多模态版 (Vision)</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="base_compute_points"
                label="算力点消耗系数"
                rules={[{ required: true, message: '请输入算力点消耗系数' }]}
                extra="1.0 = 基准系数"
              >
                <InputNumber min={0.0001} step={0.1} precision={4} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="context_window" label="上下文窗口 (K)">
                <InputNumber min={1} style={{ width: '100%' }} placeholder="64" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="max_output_tokens" label="最大输出Token">
                <InputNumber min={1} style={{ width: '100%' }} placeholder="4096" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="产品描述">
            <Input.TextArea rows={3} placeholder="请输入产品描述" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="status" label="状态">
                <Select>
                  <Select.Option value="active">启用</Select.Option>
                  <Select.Option value="inactive">禁用</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="sort_order" label="排序权重">
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminSPUs;
