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
  Switch,
  Divider,
} from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { SKUWithSPU, SPU, SKUCreateRequest, SKUUpdateRequest } from '@/types/sku';
import { SKU_TYPE_LABELS, MODEL_TIER_LABELS, SUBSCRIPTION_PERIOD_LABELS } from '@/types/sku';

const AdminSKUs = () => {
  const [skus, setSKUs] = useState<SKUWithSPU[]>([]);
  const [spus, setSPUs] = useState<SPU[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSKU, setEditingSKU] = useState<SKUWithSPU | null>(null);
  const [form] = Form.useForm();
  const [filters, setFilters] = useState({
    spu_id: '',
    type: '',
    status: 'active',
  });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });
  const [selectedSKUType, setSelectedSKUType] = useState<string>('token_pack');

  useEffect(() => {
    fetchSKUs();
    fetchSPUs();
  }, [filters, pagination.current, pagination.pageSize]);

  const fetchSKUs = async () => {
    setLoading(true);
    try {
      const response = await skuService.getSKUs({
        page: pagination.current,
        per_page: pagination.pageSize,
        ...(filters.spu_id ? { spu_id: parseInt(filters.spu_id, 10) } : {}),
        ...(filters.type ? { type: filters.type } : {}),
        ...(filters.status ? { status: filters.status } : {}),
      });
      setSKUs(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch {
      message.error('获取SKU列表失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchSPUs = async () => {
    try {
      const response = await skuService.getSPUs({ status: 'active', per_page: 100 });
      setSPUs(response.data.data || []);
    } catch {
      console.error('Failed to fetch SPUs');
    }
  };

  const handleAdd = () => {
    setEditingSKU(null);
    form.resetFields();
    form.setFieldsValue({
      sku_type: 'token_pack',
      valid_days: 365,
      stock: -1,
      status: 'active',
      group_enabled: true,
      min_group_size: 2,
      max_group_size: 10,
      is_trial: false,
      is_unlimited: false,
      is_promoted: false,
    });
    setSelectedSKUType('token_pack');
    setModalVisible(true);
  };

  const handleEdit = (record: SKUWithSPU) => {
    setEditingSKU(record);
    form.setFieldsValue(record);
    setSelectedSKUType(record.sku_type);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await skuService.deleteSKU(id);
      message.success('SKU已删除');
      fetchSKUs();
    } catch {
      message.error('删除失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();

      if (editingSKU) {
        const updateData: SKUUpdateRequest = {
          retail_price: values.retail_price,
          wholesale_price: values.wholesale_price,
          original_price: values.original_price,
          stock: values.stock,
          daily_limit: values.daily_limit,
          group_enabled: values.group_enabled,
          min_group_size: values.min_group_size,
          max_group_size: values.max_group_size,
          group_discount_rate: values.group_discount_rate,
          status: values.status,
          is_promoted: values.is_promoted,
        };
        await skuService.updateSKU(editingSKU.id, updateData);
        message.success('SKU已更新');
      } else {
        const createData: SKUCreateRequest = {
          ...values,
          sku_code: values.sku_code.toUpperCase(),
        };
        await skuService.createSKU(createData);
        message.success('SKU已创建');
      }
      setModalVisible(false);
      fetchSKUs();
    } catch (error: unknown) {
      if (error && typeof error === 'object' && 'errorFields' in error) {
        return;
      }
      message.error(editingSKU ? '更新失败' : '创建失败');
    }
  };

  const columns = [
    {
      title: 'SKU编码',
      dataIndex: 'sku_code',
      key: 'sku_code',
      width: 150,
    },
    {
      title: '关联SPU',
      dataIndex: 'spu_name',
      key: 'spu_name',
      width: 150,
      render: (name: string, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <span>{name}</span>
          <Tag>{MODEL_TIER_LABELS[record.model_tier || 'lite']}</Tag>
        </Space>
      ),
    },
    {
      title: '类型',
      dataIndex: 'sku_type',
      key: 'sku_type',
      width: 100,
      render: (type: string) => <Tag color="blue">{SKU_TYPE_LABELS[type]}</Tag>,
    },
    {
      title: '规格',
      key: 'spec',
      width: 150,
      render: (_: unknown, record: SKUWithSPU) => {
        if (record.sku_type === 'token_pack') {
          return (
            <span>
              {record.token_amount?.toLocaleString()} Tokens
              <br />({record.compute_points?.toLocaleString()} 算力点)
            </span>
          );
        }
        if (record.sku_type === 'subscription') {
          return (
            <span>
              {SUBSCRIPTION_PERIOD_LABELS[record.subscription_period || 'monthly']}
              {record.is_unlimited && ' (无限量)'}
            </span>
          );
        }
        if (record.sku_type === 'concurrent') {
          return <span>{record.concurrent_requests} 并发</span>;
        }
        return '-';
      },
    },
    {
      title: '价格',
      dataIndex: 'retail_price',
      key: 'retail_price',
      width: 100,
      render: (price: number, record: SKUWithSPU) => (
        <Space direction="vertical" size={0}>
          <span style={{ color: '#f5222d', fontWeight: 'bold' }}>¥{price.toFixed(2)}</span>
          {record.original_price && record.original_price > price && (
            <span style={{ textDecoration: 'line-through', color: '#999' }}>
              ¥{record.original_price.toFixed(2)}
            </span>
          )}
        </Space>
      ),
    },
    {
      title: '库存',
      dataIndex: 'stock',
      key: 'stock',
      width: 80,
      render: (stock: number) => (stock === -1 ? '无限' : stock),
    },
    {
      title: '拼团',
      key: 'group',
      width: 100,
      render: (_: unknown, record: SKUWithSPU) =>
        record.group_enabled ? (
          <Tag color="green">
            {record.min_group_size}-{record.max_group_size}人
          </Tag>
        ) : (
          <Tag>不支持</Tag>
        ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'success' : 'warning'}>
          {status === 'active' ? '在售' : '下架'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 150,
      render: (_: unknown, record: SKUWithSPU) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定要删除这个SKU吗？"
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
          <h2 style={{ margin: 0 }}>SKU 管理</h2>
        </Col>
        <Col>
          <Space>
            <Select
              value={filters.spu_id}
              onChange={(v) => setFilters({ ...filters, spu_id: v })}
              style={{ width: 180 }}
              placeholder="选择SPU"
              allowClear
              showSearch
              optionFilterProp="label"
            >
              {spus.map((s) => (
                <Select.Option key={s.id} value={s.id} label={s.name}>
                  {s.name}
                </Select.Option>
              ))}
            </Select>
            <Select
              value={filters.type}
              onChange={(v) => setFilters({ ...filters, type: v })}
              style={{ width: 120 }}
              placeholder="类型"
              allowClear
            >
              <Select.Option value="token_pack">Token包</Select.Option>
              <Select.Option value="subscription">订阅套餐</Select.Option>
              <Select.Option value="concurrent">并发套餐</Select.Option>
              <Select.Option value="trial">试用套餐</Select.Option>
            </Select>
            <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
              新增SKU
            </Button>
          </Space>
        </Col>
      </Row>

      <Card>
        <Table
          columns={columns}
          dataSource={skus}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={editingSKU ? '编辑SKU' : '新增SKU'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={800}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="spu_id"
                label="关联SPU"
                rules={[{ required: true, message: '请选择SPU' }]}
              >
                <Select
                  placeholder="请选择SPU"
                  showSearch
                  optionFilterProp="label"
                  disabled={!!editingSKU}
                >
                  {spus.map((s) => (
                    <Select.Option key={s.id} value={s.id} label={s.name}>
                      {s.name} ({s.spu_code})
                    </Select.Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="sku_code"
                label="SKU编码"
                rules={[{ required: true, message: '请输入SKU编码' }]}
                extra="唯一标识，如: DEEPSEEK-V3-100K"
              >
                <Input placeholder="DEEPSEEK-V3-100K" disabled={!!editingSKU} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="sku_type"
                label="SKU类型"
                rules={[{ required: true, message: '请选择SKU类型' }]}
              >
                <Select
                  placeholder="请选择"
                  disabled={!!editingSKU}
                  onChange={(v) => {
                    setSelectedSKUType(v);
                    form.setFieldsValue({
                      is_unlimited: false,
                      group_enabled: v !== 'trial',
                    });
                  }}
                >
                  <Select.Option value="token_pack">Token包</Select.Option>
                  <Select.Option value="subscription">订阅套餐</Select.Option>
                  <Select.Option value="concurrent">并发套餐</Select.Option>
                  <Select.Option value="trial">试用套餐</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="status" label="状态">
                <Select>
                  <Select.Option value="active">在售</Select.Option>
                  <Select.Option value="inactive">下架</Select.Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          {selectedSKUType === 'token_pack' && (
            <>
              <Divider>Token包配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="token_amount"
                    label="Token数量"
                    rules={[{ required: true, message: '请输入Token数量' }]}
                  >
                    <InputNumber
                      min={1}
                      style={{ width: '100%' }}
                      placeholder="100000"
                      formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="compute_points"
                    label="算力点数量"
                    rules={[{ required: true, message: '请输入算力点数量' }]}
                  >
                    <InputNumber
                      min={1}
                      style={{ width: '100%' }}
                      placeholder="100"
                      formatter={(value) => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                    />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="valid_days" label="有效天数">
                    <InputNumber min={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          {selectedSKUType === 'subscription' && (
            <>
              <Divider>订阅配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="subscription_period"
                    label="订阅周期"
                    rules={[{ required: true, message: '请选择订阅周期' }]}
                  >
                    <Select placeholder="请选择">
                      <Select.Option value="monthly">月度</Select.Option>
                      <Select.Option value="quarterly">季度</Select.Option>
                      <Select.Option value="yearly">年度</Select.Option>
                    </Select>
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="is_unlimited" label="是否无限量" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item
                    name="fair_use_limit"
                    label="公平使用上限"
                    extra="无限量套餐的每日使用上限"
                  >
                    <InputNumber min={1} style={{ width: '100%' }} placeholder="Tokens/天" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          {selectedSKUType === 'concurrent' && (
            <>
              <Divider>并发配置</Divider>
              <Row gutter={16}>
                <Col span={8}>
                  <Form.Item
                    name="concurrent_requests"
                    label="并发请求数"
                    rules={[{ required: true, message: '请输入并发请求数' }]}
                  >
                    <InputNumber min={1} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="tpm_limit" label="TPM限制">
                    <InputNumber min={1} style={{ width: '100%' }} placeholder="Tokens/分钟" />
                  </Form.Item>
                </Col>
                <Col span={8}>
                  <Form.Item name="rpm_limit" label="RPM限制">
                    <InputNumber min={1} style={{ width: '100%' }} placeholder="请求/分钟" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Divider>定价与库存</Divider>
          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="retail_price"
                label="零售价"
                rules={[{ required: true, message: '请输入零售价' }]}
              >
                <InputNumber
                  min={0}
                  precision={2}
                  style={{ width: '100%' }}
                  placeholder="9.90"
                  prefix="¥"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="original_price" label="原价">
                <InputNumber
                  min={0}
                  precision={2}
                  style={{ width: '100%' }}
                  placeholder="19.90"
                  prefix="¥"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item name="stock" label="库存" extra="-1表示无限库存">
                <InputNumber style={{ width: '100%' }} placeholder="-1" />
              </Form.Item>
            </Col>
          </Row>

          {selectedSKUType !== 'trial' && (
            <>
              <Divider>拼团配置</Divider>
              <Row gutter={16}>
                <Col span={6}>
                  <Form.Item name="group_enabled" label="支持拼团" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="min_group_size" label="最小人数">
                    <InputNumber min={2} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="max_group_size" label="最大人数">
                    <InputNumber min={2} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
                <Col span={6}>
                  <Form.Item name="group_discount_rate" label="拼团折扣率(%)" extra="如: 20 表示8折">
                    <InputNumber min={0} max={100} style={{ width: '100%' }} />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="is_promoted" label="推荐商品" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="daily_limit" label="每日购买限制">
                <InputNumber min={0} style={{ width: '100%' }} placeholder="0表示不限制" />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminSKUs;
