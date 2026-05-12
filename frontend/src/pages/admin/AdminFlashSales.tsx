import { useCallback, useEffect, useState } from 'react';
import {
  Card,
  Typography,
  Form,
  Input,
  InputNumber,
  Button,
  Space,
  message,
  Alert,
  Table,
  Drawer,
  Tag,
  Modal,
  Popconfirm,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import api from '@/services/api';
import { skuService } from '@/services/sku';
import type { SKUWithSPU } from '@/types/sku';
import type { APIResponse } from '@/types';

const { Title, Text } = Typography;

type SkuLine = {
  sku_id?: number;
  flash_price?: number;
  stock_limit?: number;
  per_user_limit?: number;
};

type FlashSaleRow = {
  id: number;
  name: string;
  description: string;
  start_time: string;
  end_time: string;
  status: string;
  created_at: string;
  updated_at: string;
};

type FlashSaleProductRow = {
  id: number;
  flash_sale_id: number;
  sku_id: number;
  product_name: string;
  flash_price: number;
  original_price: number;
  stock_limit: number;
  stock_sold: number;
  per_user_limit: number;
  discount: number;
};

/** 与套餐包一致：仅接受可上架 SKU（active + SPU active），通过 ID 查询校验，不提供全量下拉。 */
function FlashSkuIdField({
  value,
  onChange,
}: {
  value?: number;
  onChange?: (v: number | undefined) => void;
}) {
  const [checking, setChecking] = useState(false);
  const [resolved, setResolved] = useState<SKUWithSPU | null>(null);
  const [hint, setHint] = useState<string | null>(null);

  const runCheck = async () => {
    const id = Number(value);
    if (!(id > 0)) {
      message.warning('请先输入有效的 SKU ID');
      return;
    }
    setChecking(true);
    setHint(null);
    setResolved(null);
    try {
      const res = await skuService.getSKU(id);
      const sku = res.data.data as SKUWithSPU;
      if (sku.status !== 'active') {
        setHint('该 SKU 未上架（status 非 active），不可用于秒杀');
        return;
      }
      if (sku.spu_status != null && sku.spu_status !== '' && sku.spu_status !== 'active') {
        setHint('关联 SPU 未上架，不可用于秒杀');
        return;
      }
      setResolved(sku);
    } catch {
      setHint('未找到该 SKU 或无权查看，请确认 ID');
    } finally {
      setChecking(false);
    }
  };

  return (
    <div>
      <Space wrap>
        <InputNumber
          min={1}
          precision={0}
          placeholder="SKU ID"
          value={value}
          onChange={(v) => {
            onChange?.(v ?? undefined);
            setResolved(null);
            setHint(null);
          }}
          style={{ width: 140 }}
        />
        <Button type="default" onClick={() => void runCheck()} loading={checking}>
          校验可售
        </Button>
      </Space>
      {resolved && (
        <Text type="secondary" style={{ display: 'block', marginTop: 6 }}>
          {resolved.spu_name || '—'} · {resolved.sku_code} · 零售价 ¥{Number(resolved.retail_price).toFixed(2)}{' '}
          · 库存 {resolved.stock ?? '—'}
        </Text>
      )}
      {hint && (
        <Text type="danger" style={{ display: 'block', marginTop: 4 }}>
          {hint}
        </Text>
      )}
    </div>
  );
}

function statusTag(status: string) {
  const colors: Record<string, string> = {
    upcoming: 'blue',
    active: 'green',
    ended: 'default',
    canceled: 'red',
  };
  return <Tag color={colors[status] || 'default'}>{status}</Tag>;
}

const AdminFlashSales = () => {
  const [loading, setLoading] = useState(false);
  const [listLoading, setListLoading] = useState(false);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  const [rows, setRows] = useState<FlashSaleRow[]>([]);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [preview, setPreview] = useState<{
    sale: FlashSaleRow;
    skus: FlashSaleProductRow[];
  } | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [editSaving, setEditSaving] = useState(false);
  const [editingRow, setEditingRow] = useState<FlashSaleRow | null>(null);

  const loadList = useCallback(async () => {
    setListLoading(true);
    try {
      const res = await api.get<APIResponse<FlashSaleRow[]>>('/admin/flash-sales');
      setRows(res.data.data || []);
    } catch {
      message.error('加载秒杀列表失败');
    } finally {
      setListLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadList();
  }, [loadList]);

  const isStrictModelSKU = (provider?: string, modelName?: string, providerModelID?: string) => {
    const p = String(provider || '')
      .trim()
      .toLowerCase();
    if (!p || p === 'internal' || p === 'virtual_goods') return false;
    return Boolean(String(providerModelID || '').trim() || String(modelName || '').trim());
  };

  const openPreview = async (id: number) => {
    setPreviewOpen(true);
    setPreviewLoading(true);
    setPreview(null);
    try {
      type Detail = FlashSaleRow & { skus: FlashSaleProductRow[] };
      const res = await api.get<APIResponse<Detail>>(`/admin/flash-sales/${id}`);
      const d = res.data.data;
      if (!d) {
        message.error('详情为空');
        setPreviewOpen(false);
        return;
      }
      const { skus, ...sale } = d;
      setPreview({ sale: sale as FlashSaleRow, skus: skus || [] });
    } catch {
      message.error('加载详情失败');
      setPreviewOpen(false);
    } finally {
      setPreviewLoading(false);
    }
  };

  const setFlashStatus = async (id: number, status: string) => {
    try {
      await api.put(`/flash-sales/${id}/status`, { status });
      message.success('状态已更新');
      void loadList();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      message.error(err?.response?.data?.message || '更新失败');
    }
  };

  const openEdit = (row: FlashSaleRow) => {
    setEditingRow(row);
    editForm.setFieldsValue({
      name: row.name,
      description: row.description || '',
      start_time: dayjs(row.start_time).format('YYYY-MM-DDTHH:mm'),
      end_time: dayjs(row.end_time).format('YYYY-MM-DDTHH:mm'),
    });
    setEditOpen(true);
  };

  const submitEdit = async () => {
    if (!editingRow) return;
    const v = await editForm.validateFields();
    setEditSaving(true);
    try {
      await api.put(`/admin/flash-sales/${editingRow.id}`, {
        name: v.name,
        description: v.description || '',
        start_time: new Date(v.start_time).toISOString(),
        end_time: new Date(v.end_time).toISOString(),
      });
      message.success('已保存');
      setEditOpen(false);
      setEditingRow(null);
      void loadList();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      message.error(err?.response?.data?.message || '保存失败');
    } finally {
      setEditSaving(false);
    }
  };

  const columns: ColumnsType<FlashSaleRow> = [
    { title: 'ID', dataIndex: 'id', width: 72 },
    { title: '名称', dataIndex: 'name', ellipsis: true },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (s: string) => statusTag(s),
    },
    {
      title: '开始',
      dataIndex: 'start_time',
      width: 168,
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '结束',
      dataIndex: 'end_time',
      width: 168,
      render: (t: string) => dayjs(t).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'op',
      width: 280,
      render: (_, r) => (
        <Space wrap size={0}>
          <Button type="link" size="small" onClick={() => void openPreview(r.id)}>
            预览
          </Button>
          {r.status === 'upcoming' && (
            <>
              <Popconfirm title="确认立即上架为进行中？" onConfirm={() => void setFlashStatus(r.id, 'active')}>
                <Button type="link" size="small">
                  上架
                </Button>
              </Popconfirm>
              <Button type="link" size="small" onClick={() => openEdit(r)}>
                编辑
              </Button>
              <Popconfirm title="确认取消该秒杀？" onConfirm={() => void setFlashStatus(r.id, 'canceled')}>
                <Button type="link" size="small" danger>
                  取消
                </Button>
              </Popconfirm>
            </>
          )}
          {r.status === 'active' && (
            <>
              <Popconfirm title="确认提前结束？" onConfirm={() => void setFlashStatus(r.id, 'ended')}>
                <Button type="link" size="small">
                  结束
                </Button>
              </Popconfirm>
              <Popconfirm title="确认取消该秒杀？" onConfirm={() => void setFlashStatus(r.id, 'canceled')}>
                <Button type="link" size="small" danger>
                  取消
                </Button>
              </Popconfirm>
            </>
          )}
        </Space>
      ),
    },
  ];

  const onFinish = async (values: {
    name: string;
    description?: string;
    start_time: string;
    end_time: string;
    skus: SkuLine[];
  }) => {
    const startMs = new Date(values.start_time).getTime();
    const endMs = new Date(values.end_time).getTime();
    if (!(startMs > 0) || !(endMs > 0) || endMs <= startMs) {
      message.warning('请填写有效的开始与结束时间，且结束须晚于开始');
      return;
    }

    const lineRows = values.skus || [];
    const skuIds = lineRows.map((r) => Number(r.sku_id)).filter((id) => id > 0);
    const uniq = new Set(skuIds);
    if (uniq.size !== skuIds.length) {
      message.warning('同一活动中不可重复选择相同 SKU');
      return;
    }

    setLoading(true);
    try {
      for (const row of lineRows) {
        const skuID = Number(row.sku_id);
        const res = await skuService.getSKU(skuID);
        const sku = res.data.data as SKUWithSPU;
        if (sku.status !== 'active') {
          message.warning(`SKU ${skuID} 未上架，秒杀仅支持可售 SKU`);
          setLoading(false);
          return;
        }
        if (sku.spu_status != null && sku.spu_status !== '' && sku.spu_status !== 'active') {
          message.warning(`SKU ${skuID} 关联 SPU 未上架`);
          setLoading(false);
          return;
        }
        if (
          sku.sku_type === 'token_pack' &&
          !isStrictModelSKU(sku.model_provider, sku.model_name, sku.provider_model_id)
        ) {
          message.warning(`SKU ${skuID} 不可单独作为加油包秒杀，请改为带模型商品`);
          setLoading(false);
          return;
        }
        const stock = typeof sku.stock === 'number' ? sku.stock : Number(sku.stock ?? 0);
        if (stock >= 0 && Number(row.stock_limit) > stock) {
          message.warning(`SKU ${skuID} 秒杀库存（${row.stock_limit}）超过可售库存（${stock}）`);
          setLoading(false);
          return;
        }
      }
      const skus = lineRows.map((row) => ({
        sku_id: Number(row.sku_id),
        flash_price: Number(row.flash_price),
        stock_limit: Number(row.stock_limit),
        per_user_limit: row.per_user_limit != null ? Number(row.per_user_limit) : 1,
      }));
      await api.post('/flash-sales', {
        name: values.name,
        description: values.description || '',
        start_time: new Date(values.start_time).toISOString(),
        end_time: new Date(values.end_time).toISOString(),
        skus,
      });
      message.success('秒杀活动已创建');
      form.resetFields();
      void loadList();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      message.error(err?.response?.data?.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  const previewSkuCols: ColumnsType<FlashSaleProductRow> = [
    { title: 'SKU', dataIndex: 'sku_id', width: 72 },
    { title: '商品', dataIndex: 'product_name', ellipsis: true },
    {
      title: '秒杀价',
      dataIndex: 'flash_price',
      width: 88,
      render: (v: number) => `¥${Number(v).toFixed(2)}`,
    },
    {
      title: '原价',
      dataIndex: 'original_price',
      width: 88,
      render: (v: number) => `¥${Number(v).toFixed(2)}`,
    },
    { title: '限量', dataIndex: 'stock_limit', width: 72 },
    { title: '已售', dataIndex: 'stock_sold', width: 72 },
    { title: '每人', dataIndex: 'per_user_limit', width: 64 },
  ];

  return (
    <div>
      <Title level={2}>秒杀配置</Title>
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 12 }}
        message="卖场展示说明"
        description="C 端「限时秒杀」仅展示进行中且仍有剩余库存的场次。开始时间已到会自动变为进行中；未到开始时间则为待开始，卖场不展示。全部场次见下方列表。"
      />
      <Alert
        type="warning"
        showIcon
        style={{ marginBottom: 12 }}
        message="加油包限制"
        description="加油包不可单独购买。秒杀活动中若配置纯加油包 SKU，会在提交前拦截并提示。"
      />

      <Card title="历史与当前活动" style={{ marginBottom: 16 }}>
        <Table<FlashSaleRow>
          rowKey="id"
          size="small"
          loading={listLoading}
          columns={columns}
          dataSource={rows}
          pagination={{ pageSize: 20, showSizeChanger: true }}
        />
      </Card>

      <Card title="新建秒杀活动">
        <Form form={form} layout="vertical" onFinish={onFinish} initialValues={{ skus: [{}] }}>
          <Form.Item
            name="name"
            label="活动名称"
            rules={[{ required: true, message: '请输入名称' }]}
          >
            <Input placeholder="例如：春季限时秒杀" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="可选" />
          </Form.Item>
          <Form.Item
            name="start_time"
            label="开始时间"
            rules={[{ required: true, message: '请选择开始时间' }]}
          >
            <Input type="datetime-local" />
          </Form.Item>
          <Form.Item
            name="end_time"
            label="结束时间"
            rules={[{ required: true, message: '请选择结束时间' }]}
          >
            <Input type="datetime-local" />
          </Form.Item>

          <Form.List name="skus">
            {(fields, { add, remove }) => (
              <>
                {fields.map((field) => (
                  <Space key={field.key} align="start" wrap style={{ marginBottom: 8 }}>
                    <Form.Item
                      name={[field.name, 'sku_id']}
                      label="SKU ID"
                      rules={[{ required: true, message: '请输入 SKU ID' }]}
                    >
                      <FlashSkuIdField />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'flash_price']}
                      label="秒杀价"
                      rules={[{ required: true, message: '必填' }]}
                    >
                      <InputNumber min={0} step={0.01} placeholder="元" style={{ width: 120 }} />
                    </Form.Item>
                    <Form.Item
                      name={[field.name, 'stock_limit']}
                      label="秒杀库存"
                      rules={[{ required: true, message: '必填' }]}
                    >
                      <InputNumber min={1} placeholder="件" style={{ width: 100 }} />
                    </Form.Item>
                    <Form.Item name={[field.name, 'per_user_limit']} label="每人限购">
                      <InputNumber min={1} placeholder="默认 1" style={{ width: 100 }} />
                    </Form.Item>
                    <MinusCircleOutlined
                      style={{ marginTop: 30, color: '#999' }}
                      onClick={() => remove(field.name)}
                    />
                  </Space>
                ))}
                <Form.Item>
                  <Button type="dashed" onClick={() => add()} block icon={<PlusOutlined />}>
                    添加 SKU
                  </Button>
                </Form.Item>
              </>
            )}
          </Form.List>

          <Button type="primary" htmlType="submit" loading={loading}>
            创建秒杀活动
          </Button>
        </Form>
      </Card>

      <Modal
        title={editingRow ? `编辑场次 #${editingRow.id}` : '编辑'}
        open={editOpen}
        onCancel={() => {
          setEditOpen(false);
          setEditingRow(null);
        }}
        onOk={() => void submitEdit()}
        confirmLoading={editSaving}
        destroyOnClose
        width={520}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item name="name" label="活动名称" rules={[{ required: true, message: '必填' }]}>
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="start_time" label="开始时间" rules={[{ required: true, message: '必填' }]}>
            <Input type="datetime-local" />
          </Form.Item>
          <Form.Item name="end_time" label="结束时间" rules={[{ required: true, message: '必填' }]}>
            <Input type="datetime-local" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        title={preview ? `预览 · ${preview.sale.name}` : '预览'}
        width={720}
        open={previewOpen}
        onClose={() => {
          setPreviewOpen(false);
          setPreview(null);
        }}
        destroyOnClose
      >
        {previewLoading ? (
          <Text type="secondary">加载中…</Text>
        ) : preview ? (
          <>
            <Space direction="vertical" size={8} style={{ marginBottom: 16 }}>
              <div>
                状态 {statusTag(preview.sale.status)} · ID {preview.sale.id}
              </div>
              <Text type="secondary">
                {dayjs(preview.sale.start_time).format('YYYY-MM-DD HH:mm')} —{' '}
                {dayjs(preview.sale.end_time).format('YYYY-MM-DD HH:mm')}
              </Text>
              {preview.sale.description ? <div>{preview.sale.description}</div> : null}
            </Space>
            <Table<FlashSaleProductRow>
              rowKey="id"
              size="small"
              columns={previewSkuCols}
              dataSource={preview.skus}
              pagination={false}
            />
          </>
        ) : null}
      </Drawer>
    </div>
  );
};

export default AdminFlashSales;
