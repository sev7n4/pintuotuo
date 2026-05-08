import { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Space, Modal, Form, Select, InputNumber, message, Popconfirm, Row, Col } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { endpointPricingService } from '@/services/endpointPricing';
import type { EndpointPricing, EndpointPricingCreateRequest } from '@/types/sku';
import { ENDPOINT_TYPE_LABELS, ENDPOINT_TYPE_COLORS, BILLING_UNIT_LABELS } from '@/types/sku';
import { getApiErrorMessage } from '@/utils/apiError';

const endpointTypeOptions = Object.entries(ENDPOINT_TYPE_LABELS).map(([value, label]) => ({ value, label }));
const billingUnitOptions = Object.entries(BILLING_UNIT_LABELS).map(([value, label]) => ({ value, label }));

const AdminEndpointPricing = () => {
  const [items, setItems] = useState<EndpointPricing[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingItem, setEditingItem] = useState<EndpointPricing | null>(null);
  const [form] = Form.useForm();
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 });
  const [endpointTypeFilter, setEndpointTypeFilter] = useState<string | undefined>(undefined);

  useEffect(() => {
    fetchItems();
  }, [pagination.current, pagination.pageSize, endpointTypeFilter]);

  const fetchItems = async () => {
    setLoading(true);
    try {
      const response = await endpointPricingService.getList({
        page: pagination.current,
        per_page: pagination.pageSize,
        endpoint_type: endpointTypeFilter,
      });
      setItems(response.data.data || []);
      setPagination((prev) => ({ ...prev, total: response.data.total }));
    } catch (e) {
      message.error(getApiErrorMessage(e, '获取端点计费配置失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleAdd = () => {
    setEditingItem(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: EndpointPricing) => {
    setEditingItem(record);
    form.setFieldsValue(record);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await endpointPricingService.delete(id);
      message.success('删除成功');
      fetchItems();
    } catch (e) {
      message.error(getApiErrorMessage(e, '删除失败'));
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (editingItem) {
        await endpointPricingService.update(editingItem.id, values);
        message.success('更新成功');
      } else {
        await endpointPricingService.create(values as EndpointPricingCreateRequest);
        message.success('创建成功');
      }
      setModalVisible(false);
      fetchItems();
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error(getApiErrorMessage(e, editingItem ? '更新失败' : '创建失败'));
    }
  };

  const columns = [
    {
      title: '端点类型',
      dataIndex: 'endpoint_type',
      key: 'endpoint_type',
      render: (type: string) => (
        <Tag color={ENDPOINT_TYPE_COLORS[type] || 'default'}>
          {ENDPOINT_TYPE_LABELS[type] || type}
        </Tag>
      ),
    },
    {
      title: '厂商代码',
      dataIndex: 'provider_code',
      key: 'provider_code',
    },
    {
      title: '计费单位',
      dataIndex: 'unit_type',
      key: 'unit_type',
      render: (type: string) => BILLING_UNIT_LABELS[type] || type,
    },
    {
      title: '单价（元）',
      dataIndex: 'unit_price',
      key: 'unit_price',
      render: (price: number) => price.toFixed(6),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      render: (v: string) => v ? new Date(v).toLocaleString() : '-',
    },
    {
      title: '操作',
      key: 'action',
      render: (_: unknown, record: EndpointPricing) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)} okText="确定" cancelText="取消">
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
      <Card>
        <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
          <Col span={8}>
            <Select
              placeholder="筛选端点类型"
              allowClear
              style={{ width: '100%' }}
              value={endpointTypeFilter}
              onChange={(v) => {
                setEndpointTypeFilter(v);
                setPagination((prev) => ({ ...prev, current: 1 }));
              }}
              options={endpointTypeOptions}
            />
          </Col>
          <Col span={16} style={{ textAlign: 'right' }}>
            <Space>
              <Button onClick={fetchItems}>刷新</Button>
              <Button type="primary" icon={<PlusOutlined />} onClick={handleAdd}>
                新增
              </Button>
            </Space>
          </Col>
        </Row>

        <Table
          columns={columns}
          dataSource={items}
          rowKey="id"
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (page, pageSize) => setPagination({ ...pagination, current: page, pageSize }),
          }}
        />
      </Card>

      <Modal
        title={editingItem ? '编辑端点计费' : '新增端点计费'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        okText="保存"
        cancelText="取消"
        width={600}
      >
        <Form form={form} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="endpoint_type" label="端点类型" rules={[{ required: true, message: '请选择端点类型' }]}>
                <Select options={endpointTypeOptions} placeholder="选择端点类型" disabled={!!editingItem} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="provider_code" label="厂商代码" rules={[{ required: true, message: '请输入厂商代码' }]}>
                <Select
                  placeholder="选择厂商"
                  disabled={!!editingItem}
                  options={[
                    { value: 'openai', label: 'OpenAI' },
                    { value: 'anthropic', label: 'Anthropic' },
                    { value: 'google', label: 'Google' },
                  ]}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="unit_type" label="计费单位" rules={[{ required: true, message: '请选择计费单位' }]}>
                <Select options={billingUnitOptions} placeholder="选择计费单位" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="unit_price" label="单价（元）" rules={[{ required: true, message: '请输入单价' }]}>
                <InputNumber style={{ width: '100%' }} min={0} precision={6} step={0.000001} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminEndpointPricing;
