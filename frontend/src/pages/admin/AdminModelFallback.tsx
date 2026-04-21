import { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Table,
  Button,
  Tag,
  Alert,
  Modal,
  Form,
  Input,
  AutoComplete,
  Switch,
  message,
  Space,
  Popconfirm,
} from 'antd';
import { EditOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { skuService } from '@/services/sku';
import type { ModelFallbackRule } from '@/types/sku';
import type { AxiosError } from 'axios';

function splitModelLines(s: string): string[] {
  return s
    .split(/\r?\n/)
    .map((x) => x.trim())
    .filter(Boolean);
}

function apiErrorMessage(err: unknown, fallback: string): string {
  if (err && typeof err === 'object' && 'response' in err) {
    const ax = err as AxiosError<{ message?: string; error?: string }>;
    const d = ax.response?.data;
    if (d && typeof d === 'object') {
      if (typeof d.message === 'string' && d.message) return d.message;
      if (typeof d.error === 'string' && d.error) return d.error;
    }
  }
  return fallback;
}

const AdminModelFallback = () => {
  const [rows, setRows] = useState<ModelFallbackRule[]>([]);
  const [catalogKeys, setCatalogKeys] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState<ModelFallbackRule | null>(null);
  const [form] = Form.useForm();

  const fetchCatalog = useCallback(async () => {
    try {
      const res = await skuService.getModelCatalogKeys();
      setCatalogKeys(res.data.data || []);
    } catch {
      message.warning('加载目录模型列表失败，仍可手动输入 provider/model');
    }
  }, []);

  const fetchRules = useCallback(async () => {
    setLoading(true);
    try {
      const res = await skuService.listModelFallbackRules();
      setRows(res.data.data || []);
    } catch {
      message.error('获取 fallback 规则失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCatalog();
    fetchRules();
  }, [fetchCatalog, fetchRules]);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      enabled: true,
      fallback_lines: '',
      notes: '',
    });
    setModalVisible(true);
  };

  const openEdit = (record: ModelFallbackRule) => {
    setEditing(record);
    form.setFieldsValue({
      source_model: record.source_model,
      fallback_lines: (record.fallback_models || []).join('\n'),
      enabled: record.enabled,
      notes: record.notes ?? '',
    });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const fallback_models = splitModelLines(String(values.fallback_lines ?? ''));
      const enabled = values.enabled !== false;
      if (enabled && fallback_models.length === 0) {
        message.error('启用规则时需至少填写一个备用模型');
        return;
      }
      const payload = {
        source_model: String(values.source_model).trim(),
        fallback_models,
        enabled,
        notes: values.notes ? String(values.notes).trim() : undefined,
      };
      if (editing) {
        await skuService.patchModelFallbackRule(editing.id, payload);
        message.success('已保存');
      } else {
        await skuService.createModelFallbackRule(payload);
        message.success('已创建');
      }
      setModalVisible(false);
      setEditing(null);
      fetchRules();
    } catch (e: unknown) {
      if (e && typeof e === 'object' && 'errorFields' in e) return;
      message.error(apiErrorMessage(e, editing ? '保存失败' : '创建失败'));
    }
  };

  const columns = [
    {
      title: '主模型',
      dataIndex: 'source_model',
      key: 'source_model',
      width: 220,
      ellipsis: true,
    },
    {
      title: '备用链（顺序）',
      dataIndex: 'fallback_models',
      key: 'fallback_models',
      ellipsis: true,
      render: (arr: string[]) =>
        arr && arr.length > 0 ? (
          <span title={arr.join(' → ')}>{arr.join(' → ')}</span>
        ) : (
          <span style={{ color: '#999' }}>—</span>
        ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 88,
      render: (en: boolean) => (en ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>),
    },
    {
      title: '备注',
      dataIndex: 'notes',
      key: 'notes',
      width: 160,
      ellipsis: true,
      render: (t: string) => t || '—',
    },
    {
      title: '操作',
      key: 'action',
      width: 140,
      fixed: 'right' as const,
      render: (_: unknown, record: ModelFallbackRule) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该规则？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const handleDelete = async (id: number) => {
    try {
      await skuService.deleteModelFallbackRule(id);
      message.success('已删除');
      fetchRules();
    } catch (e: unknown) {
      message.error(apiErrorMessage(e, '删除失败'));
    }
  };

  return (
    <div>
      <Card
        title="模型级 Fallback"
        extra={
          <Space>
            <Button type="primary" icon={<PlusOutlined />} onClick={() => openCreate()}>
              新建规则
            </Button>
            <Button onClick={() => fetchRules()}>刷新</Button>
          </Space>
        }
      >
        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 12 }}
          message="配置说明"
          description="主模型与备用模型均须为上架 SPU 目录中的 provider/model。保存时会做存在性校验、去重、禁止自 fallback，并在启用规则之间检测循环依赖。代理层消费该配置将在后续迭代接入。"
        />
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          dataSource={rows}
          scroll={{ x: 960 }}
          pagination={{ pageSize: 30, showSizeChanger: true }}
        />
      </Card>

      <Modal
        title={editing ? `编辑规则 #${editing.id}` : '新建规则'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          setModalVisible(false);
          setEditing(null);
        }}
        destroyOnClose
        width={640}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="source_model"
            label="主模型（provider/model）"
            rules={[{ required: true, message: '请填写主模型' }]}
          >
            <AutoComplete
              options={catalogKeys.map((k) => ({ value: k }))}
              filterOption={(inputValue, option) =>
                String(option?.value ?? '')
                  .toLowerCase()
                  .includes(String(inputValue).toLowerCase())
              }
              placeholder="例如 openai/gpt-4o"
              disabled={!!editing}
            />
          </Form.Item>
          <Form.Item name="fallback_lines" label="备用模型链（每行一个，自上而下为优先顺序）">
            <Input.TextArea
              rows={5}
              placeholder={'例如：\nopenai/gpt-4o-mini\ndeepseek/deepseek-chat'}
            />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked" initialValue={true}>
            <Switch />
          </Form.Item>
          <Form.Item name="notes" label="备注">
            <Input.TextArea rows={2} placeholder="可选" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminModelFallback;
