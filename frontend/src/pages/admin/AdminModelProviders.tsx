import { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Modal, Form, Input, Select, InputNumber, message } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { ModelProvider } from '@/types/sku';

const AdminModelProviders = () => {
  const [rows, setRows] = useState<ModelProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState<ModelProvider | null>(null);
  const [form] = Form.useForm();

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await skuService.getAllModelProviders();
      setRows(res.data.data || []);
    } catch {
      message.error('获取模型厂商列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const handleEdit = (record: ModelProvider) => {
    setEditing(record);
    form.setFieldsValue({
      name: record.name,
      api_base_url: record.api_base_url ?? '',
      api_format: record.api_format,
      billing_type: record.billing_type ?? '',
      status: record.status,
      sort_order: record.sort_order,
    });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!editing) return;
      await skuService.patchModelProvider(editing.id, {
        name: values.name,
        api_base_url: values.api_base_url || undefined,
        api_format: values.api_format,
        billing_type: values.billing_type || undefined,
        status: values.status,
        sort_order: values.sort_order,
      });
      message.success('已保存');
      setModalVisible(false);
      setEditing(null);
      fetchList();
    } catch (e) {
      if ((e as { errorFields?: unknown }).errorFields) return;
      message.error('保存失败');
    }
  };

  const columns = [
    {
      title: '代码',
      dataIndex: 'code',
      key: 'code',
      width: 120,
      fixed: 'left' as const,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 140,
    },
    {
      title: 'API Base URL',
      dataIndex: 'api_base_url',
      key: 'api_base_url',
      ellipsis: true,
      render: (v: string) => v || '—',
    },
    {
      title: 'API 格式',
      dataIndex: 'api_format',
      key: 'api_format',
      width: 140,
    },
    {
      title: '计费类型',
      dataIndex: 'billing_type',
      key: 'billing_type',
      width: 100,
      render: (v: string) => v || '—',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (s: string) => (s === 'active' ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>),
    },
    {
      title: '排序',
      dataIndex: 'sort_order',
      key: 'sort_order',
      width: 72,
    },
    {
      title: '操作',
      key: 'action',
      width: 88,
      fixed: 'right' as const,
      render: (_: unknown, record: ModelProvider) => (
        <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
          编辑
        </Button>
      ),
    },
  ];

  return (
    <div>
      <Card title="模型厂商" extra={<Button onClick={() => fetchList()}>刷新</Button>}>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={rows}
          scroll={{ x: 1100 }}
          pagination={{ pageSize: 50, showSizeChanger: true }}
        />
      </Card>

      <Modal
        title={editing ? `编辑：${editing.code}` : '编辑'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditing(null);
        }}
        destroyOnClose
        width={560}
      >
        {editing && (
          <Form form={form} layout="vertical">
            <Form.Item label="代码（只读）">
              <Input value={editing.code} disabled />
            </Form.Item>
            <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
              <Input />
            </Form.Item>
            <Form.Item name="api_base_url" label="API Base URL">
              <Input placeholder="例如 https://api.openai.com/v1" />
            </Form.Item>
            <Form.Item name="api_format" label="API 格式" rules={[{ required: true }]}>
              <Input placeholder="如 openai_compatible" />
            </Form.Item>
            <Form.Item name="billing_type" label="计费类型">
              <Input placeholder="可选" />
            </Form.Item>
            <Form.Item name="status" label="状态" rules={[{ required: true }]}>
              <Select
                options={[
                  { value: 'active', label: '启用' },
                  { value: 'inactive', label: '停用' },
                ]}
              />
            </Form.Item>
            <Form.Item name="sort_order" label="排序" rules={[{ required: true }]}>
              <InputNumber min={0} style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        )}
      </Modal>
    </div>
  );
};

export default AdminModelProviders;
