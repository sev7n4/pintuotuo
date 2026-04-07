import { useEffect, useState } from 'react';
import { Card, Table, Button, Tag, Modal, Form, Input, Select, InputNumber, message, Space } from 'antd';
import { EditOutlined, PlusOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { ModelProvider } from '@/types/sku';

type ModalMode = 'create' | 'edit';

const AdminModelProviders = () => {
  const [rows, setRows] = useState<ModelProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [modalMode, setModalMode] = useState<ModalMode>('edit');
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

  const openCreate = () => {
    setModalMode('create');
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      api_format: 'openai',
      billing_type: 'flat',
      status: 'active',
      sort_order: 0,
    });
    setModalVisible(true);
  };

  const handleEdit = (record: ModelProvider) => {
    setModalMode('edit');
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
      if (modalMode === 'create') {
        await skuService.createModelProvider({
          code: (values.code as string).trim().toLowerCase(),
          name: (values.name as string).trim(),
          api_base_url: values.api_base_url ? String(values.api_base_url).trim() : undefined,
          api_format: values.api_format,
          billing_type: values.billing_type ? String(values.billing_type).trim() : undefined,
          status: values.status,
          sort_order: values.sort_order as number,
        });
        message.success('已创建');
      } else {
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
      }
      setModalVisible(false);
      setEditing(null);
      fetchList();
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error(modalMode === 'create' ? '创建失败' : '保存失败');
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

  const apiFormatOptions = [
    { value: 'openai', label: 'openai（OpenAI 兼容 /chat/completions）' },
    { value: 'anthropic', label: 'anthropic（/messages + x-api-key）' },
    { value: 'baidu', label: 'baidu（千帆等，需与代理实现一致）' },
  ];

  return (
    <div>
      <Card
        title="模型厂商"
        extra={
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()}>
              新增厂商
            </Button>
            <Button onClick={() => fetchList()}>刷新</Button>
          </Space>
        }
      >
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
        title={modalMode === 'create' ? '新增模型厂商' : editing ? `编辑：${editing.code}` : '编辑'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditing(null);
        }}
        destroyOnClose
        width={560}
      >
        <Form form={form} layout="vertical">
          {modalMode === 'create' ? (
            <Form.Item
              name="code"
              label="代码（唯一，小写字母/数字/下划线）"
              rules={[
                { required: true, message: '请输入代码' },
                {
                  pattern: /^[a-z][a-z0-9_]{0,48}$/,
                  message: '须以小写字母开头，仅 a-z、0-9、下划线，最长 50 字符',
                },
              ]}
            >
              <Input placeholder="例如 minimax" autoComplete="off" />
            </Form.Item>
          ) : (
            editing && (
              <Form.Item label="代码（只读）">
                <Input value={editing.code} disabled />
              </Form.Item>
            )
          )}
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="api_base_url" label="API Base URL">
            <Input placeholder="例如 https://api.openai.com/v1" />
          </Form.Item>
          <Form.Item name="api_format" label="API 格式" rules={[{ required: true }]}>
            <Select options={apiFormatOptions} placeholder="选择格式" />
          </Form.Item>
          <Form.Item name="billing_type" label="计费类型">
            <Input placeholder="可选，默认 flat" />
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
      </Modal>
    </div>
  );
};

export default AdminModelProviders;
