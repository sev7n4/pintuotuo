import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
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
  Modal,
  Popconfirm,
  Select,
  Tag,
  Row,
  Col,
  DatePicker,
  Spin,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { MinusCircleOutlined, PlusOutlined, EyeOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import api from '@/services/api';
import { skuService } from '@/services/sku';
import type { SKUWithSPU } from '@/types/sku';
import type { APIResponse } from '@/types';

const { Title, Text, Paragraph } = Typography;

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
  is_featured?: boolean;
  badge_text?: string;
  badge_text_secondary?: string;
  marketing_line?: string;
  promo_label?: string;
  promo_ends_at?: string | null;
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

type CreateFormValues = {
  name: string;
  description?: string;
  start_time?: dayjs.Dayjs;
  end_time?: dayjs.Dayjs;
  is_featured?: boolean;
  badge_text?: string;
  badge_text_secondary?: string;
  marketing_line?: string;
  promo_label?: string;
  promo_ends_at?: dayjs.Dayjs;
  skus: SkuLine[];
};

type PreviewContext = {
  isDraft: boolean;
  warnings: string[];
};

function collectFlashPreviewWarnings(sale: FlashSaleRow, skus: FlashSaleProductRow[]): string[] {
  const w: string[] = [];
  if (sale.status === 'upcoming') {
    w.push('当前为「待开始」，用户端「限时秒杀」卖场仅展示进行中场次；待开始仅出现在「即将开始」区块。');
  }
  if (sale.status === 'ended' || sale.status === 'canceled') {
    w.push('该场次已结束或已取消，用户端不再展示。');
  }
  const now = dayjs();
  if (sale.start_time && dayjs(sale.start_time).isAfter(now) && sale.status === 'active') {
    w.push('开始时间晚于当前（状态异常），请核对后台调度。');
  }
  if (sale.promo_ends_at && !dayjs(sale.promo_ends_at).isAfter(now)) {
    w.push('活动结束时间（展示）已过期，角标/倒计时类展示可能不再符合运营预期。');
  }
  const live = skus.filter((p) => p.stock_limit > p.stock_sold);
  if (live.length === 0 && skus.length > 0) {
    w.push('全部 SKU 已售罄，用户端进行中列表将隐藏该场次。');
  }
  return w;
}

function buildDraftFlashSalePreview(v: CreateFormValues, skuById: Map<number, SKUWithSPU>): {
  sale: FlashSaleRow;
  skus: FlashSaleProductRow[];
} {
  const startISO = v.start_time?.toISOString() || new Date().toISOString();
  const endISO = v.end_time?.toISOString() || new Date().toISOString();
  const sale: FlashSaleRow = {
    id: 0,
    name: (v.name || '').trim() || '（未命名）',
    description: (v.description || '').trim(),
    start_time: startISO,
    end_time: endISO,
    status: 'upcoming',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    is_featured: !!v.is_featured,
    badge_text: (v.badge_text || '').trim(),
    badge_text_secondary: (v.badge_text_secondary || '').trim(),
    marketing_line: (v.marketing_line || '').trim(),
    promo_label: (v.promo_label || '').trim(),
    promo_ends_at: v.promo_ends_at ? v.promo_ends_at.toISOString() : undefined,
  };
  const skus: FlashSaleProductRow[] = (v.skus || []).map((line, idx) => {
    const sid = Number(line.sku_id);
    const s = skuById.get(sid);
    const name = s ? `${s.spu_name} · ${s.sku_code}` : `SKU #${sid}`;
    const orig = s ? Number(s.retail_price) : 0;
    const fp = Number(line.flash_price ?? 0);
    const disc = orig > 0 ? Math.round((1 - fp / orig) * 100) : 0;
    return {
      id: -(idx + 1),
      flash_sale_id: 0,
      sku_id: sid,
      product_name: name,
      flash_price: fp,
      original_price: orig,
      stock_limit: Number(line.stock_limit ?? 0),
      stock_sold: 0,
      per_user_limit: line.per_user_limit != null ? Number(line.per_user_limit) : 1,
      discount: disc,
    };
  });
  return { sale, skus };
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
  const [form] = Form.useForm<CreateFormValues>();
  const [editForm] = Form.useForm<CreateFormValues>();
  const [rows, setRows] = useState<FlashSaleRow[]>([]);
  const [sellableSkus, setSellableSkus] = useState<SKUWithSPU[]>([]);
  const [skuPoolLoading, setSkuPoolLoading] = useState(false);
  const skuSearchTimer = useRef<number>();

  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [preview, setPreview] = useState<{
    sale: FlashSaleRow;
    skus: FlashSaleProductRow[];
  } | null>(null);
  const [previewCtx, setPreviewCtx] = useState<PreviewContext>({ isDraft: false, warnings: [] });

  const [editOpen, setEditOpen] = useState(false);
  const [editSaving, setEditSaving] = useState(false);
  const [editingRow, setEditingRow] = useState<FlashSaleRow | null>(null);

  const loadSellableSkus = useCallback(async (q?: string) => {
    setSkuPoolLoading(true);
    try {
      const res = await skuService.getSKUs({
        scope: 'sellable',
        per_page: 100,
        q: q?.trim() || undefined,
      });
      setSellableSkus(res.data.data || []);
    } catch {
      message.error('加载可售 SKU 失败');
      setSellableSkus([]);
    } finally {
      setSkuPoolLoading(false);
    }
  }, []);

  const scheduleSkuSearch = (q: string) => {
    window.clearTimeout(skuSearchTimer.current);
    skuSearchTimer.current = window.setTimeout(() => {
      void loadSellableSkus(q);
    }, 320);
  };

  useEffect(() => {
    void loadSellableSkus('');
    return () => window.clearTimeout(skuSearchTimer.current);
  }, [loadSellableSkus]);

  const skuOptions = useMemo(
    () =>
      sellableSkus.map((s) => ({
        value: s.id,
        label: `${s.sku_code} / ${s.spu_name} (¥${Number(s.retail_price).toFixed(2)})`,
      })),
    [sellableSkus]
  );

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

  const openSavedPreview = async (id: number) => {
    setPreviewOpen(true);
    setPreviewLoading(true);
    setPreview(null);
    setPreviewCtx({ isDraft: false, warnings: [] });
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
      const sk = skus || [];
      setPreview({ sale: sale as FlashSaleRow, skus: sk });
      setPreviewCtx({ isDraft: false, warnings: collectFlashPreviewWarnings(sale as FlashSaleRow, sk) });
    } catch {
      message.error('加载详情失败');
      setPreviewOpen(false);
    } finally {
      setPreviewLoading(false);
    }
  };

  const openDraftPreview = async () => {
    try {
      const v = await form.validateFields();
      const lineRows = v.skus || [];
      for (const row of lineRows) {
        if (!row.sku_id || row.sku_id <= 0) {
          message.error('请为每一行选择 SKU（可售池）');
          return;
        }
      }
      if (!v.start_time || !v.end_time || !v.end_time.isAfter(v.start_time)) {
        message.error('请填写有效的开始与结束时间，且结束须晚于开始');
        return;
      }
      const byId = new Map(sellableSkus.map((s) => [s.id, s]));
      const missing = lineRows.some((r) => !byId.has(Number(r.sku_id)));
      const { sale, skus } = buildDraftFlashSalePreview(v, byId);
      const extra: string[] = [];
      if (missing) extra.push('部分 SKU 不在当前搜索结果中，保存前请用搜索重新选中或刷新可售列表。');
      setPreview({ sale, skus });
      setPreviewCtx({
        isDraft: true,
        warnings: [...collectFlashPreviewWarnings(sale, skus), ...extra],
      });
      setPreviewOpen(true);
    } catch {
      message.info('请先完善表单必填项后再预览');
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
      start_time: dayjs(row.start_time),
      end_time: dayjs(row.end_time),
      is_featured: !!row.is_featured,
      badge_text: row.badge_text || '',
      badge_text_secondary: row.badge_text_secondary || '',
      marketing_line: row.marketing_line || '',
      promo_label: row.promo_label || '',
      promo_ends_at: row.promo_ends_at ? dayjs(row.promo_ends_at) : undefined,
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
        start_time: v.start_time?.toISOString(),
        end_time: v.end_time?.toISOString(),
        is_featured: !!v.is_featured,
        badge_text: (v.badge_text || '').trim(),
        badge_text_secondary: (v.badge_text_secondary || '').trim(),
        marketing_line: (v.marketing_line || '').trim(),
        promo_label: (v.promo_label || '').trim(),
        promo_ends_at: v.promo_ends_at ? v.promo_ends_at.toISOString() : '',
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
      title: '运营',
      key: 'ops',
      width: 200,
      render: (_, r) => (
        <Space wrap size={0}>
          {r.is_featured ? <Tag color="gold">推荐</Tag> : null}
          {r.badge_text ? <Tag color="purple">{r.badge_text}</Tag> : null}
          {r.promo_label ? <Tag color="blue">{r.promo_label}</Tag> : null}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'op',
      width: 280,
      render: (_, r) => (
        <Space wrap size={0}>
          <Button type="link" size="small" onClick={() => void openSavedPreview(r.id)}>
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

  const onFinish = async (values: CreateFormValues) => {
    const st = values.start_time;
    const en = values.end_time;
    if (!st || !en || !en.isAfter(st)) {
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
        start_time: st.toISOString(),
        end_time: en.toISOString(),
        is_featured: !!values.is_featured,
        badge_text: (values.badge_text || '').trim(),
        badge_text_secondary: (values.badge_text_secondary || '').trim(),
        marketing_line: (values.marketing_line || '').trim(),
        promo_label: (values.promo_label || '').trim(),
        promo_ends_at: values.promo_ends_at ? values.promo_ends_at.toISOString() : undefined,
        skus,
      });
      message.success('秒杀活动已创建');
      form.resetFields();
      form.setFieldsValue({
        skus: [{}],
        is_featured: false,
      });
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
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{ skus: [{}], is_featured: false }}
        >
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 12 }}
            message="SKU 选择"
            description="与「套餐包」SKU 明细一致：候选池为「可售」SKU（SKU 与所属 SPU 均在售）。支持输入编码或名称远程搜索。"
          />

          <Card size="small" title="基础信息" style={{ marginBottom: 12 }}>
            <Row gutter={[16, 8]}>
              <Col xs={24} md={12}>
                <Form.Item name="name" label="活动名称" rules={[{ required: true, message: '请输入名称' }]}>
                  <Input placeholder="例如：春季限时秒杀" />
                </Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="is_featured" label="推荐位">
                  <Select
                    options={[
                      { label: '否', value: false },
                      { label: '是', value: true },
                    ]}
                  />
                </Form.Item>
              </Col>
              <Col span={24}>
                <Form.Item name="description" label="描述">
                  <Input.TextArea rows={2} placeholder="可选" />
                </Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="marketing_line" label="一句话卖点（前台）">
                  <Input placeholder="可选，展示在卖场活动标题旁" allowClear />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Card size="small" title="展示与活动" style={{ marginBottom: 12 }}>
            <Row gutter={[16, 8]}>
              <Col xs={24} md={8}>
                <Form.Item name="badge_text" label="主角标">
                  <Input placeholder="如 限时特惠" allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="badge_text_secondary" label="次角标">
                  <Input placeholder="如 赠算力" allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="promo_label" label="活动标签（轻量）">
                  <Input placeholder="如 限时立减（展示用）" allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="promo_ends_at" label="活动结束时间（展示）">
                  <DatePicker showTime style={{ width: '100%' }} allowClear />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Card size="small" title="场次时间（定时开抢 / 结束）" style={{ marginBottom: 12 }}>
            <Row gutter={[16, 8]}>
              <Col xs={24} md={12}>
                <Form.Item
                  name="start_time"
                  label="开始时间"
                  rules={[{ required: true, message: '请选择开始时间' }]}
                >
                  <DatePicker showTime style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item
                  name="end_time"
                  label="结束时间"
                  rules={[{ required: true, message: '请选择结束时间' }]}
                >
                  <DatePicker showTime style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Card size="small" title="SKU 明细（可售池）" style={{ marginBottom: 12 }}>
            <Form.List name="skus">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <Space key={field.key} align="start" wrap style={{ marginBottom: 8 }}>
                      <Form.Item
                        name={[field.name, 'sku_id']}
                        label="SKU"
                        rules={[{ required: true, message: '请选择 SKU' }]}
                      >
                        <Select
                          showSearch
                          filterOption={false}
                          placeholder="搜索 SKU 编码 / SPU 名称"
                          options={skuOptions}
                          loading={skuPoolLoading}
                          style={{ minWidth: 280 }}
                          onSearch={(q) => scheduleSkuSearch(q)}
                          onDropdownVisibleChange={(open) => {
                            if (open && sellableSkus.length === 0) void loadSellableSkus('');
                          }}
                        />
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
          </Card>

          <Space wrap>
            <Button type="primary" htmlType="submit" loading={loading}>
              创建秒杀活动
            </Button>
            <Button icon={<EyeOutlined />} onClick={() => void openDraftPreview()}>
              预览用户端
            </Button>
          </Space>
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
        width={720}
      >
        <Form form={editForm} layout="vertical">
          <Card size="small" title="基础信息" style={{ marginBottom: 12 }}>
            <Form.Item name="name" label="活动名称" rules={[{ required: true, message: '必填' }]}>
              <Input />
            </Form.Item>
            <Form.Item name="description" label="描述">
              <Input.TextArea rows={2} />
            </Form.Item>
            <Form.Item name="is_featured" label="推荐位">
              <Select
                options={[
                  { label: '否', value: false },
                  { label: '是', value: true },
                ]}
              />
            </Form.Item>
            <Form.Item name="marketing_line" label="一句话卖点（前台）">
              <Input allowClear />
            </Form.Item>
          </Card>
          <Card size="small" title="展示与活动" style={{ marginBottom: 12 }}>
            <Row gutter={[12, 8]}>
              <Col xs={24} md={8}>
                <Form.Item name="badge_text" label="主角标">
                  <Input allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="badge_text_secondary" label="次角标">
                  <Input allowClear />
                </Form.Item>
              </Col>
              <Col xs={24} md={8}>
                <Form.Item name="promo_label" label="活动标签">
                  <Input allowClear />
                </Form.Item>
              </Col>
              <Col span={24}>
                <Form.Item name="promo_ends_at" label="活动结束时间（展示）">
                  <DatePicker showTime style={{ width: '100%' }} allowClear />
                </Form.Item>
              </Col>
            </Row>
          </Card>
          <Card size="small" title="场次时间">
            <Row gutter={[12, 8]}>
              <Col xs={24} md={12}>
                <Form.Item name="start_time" label="开始时间" rules={[{ required: true, message: '必填' }]}>
                  <DatePicker showTime style={{ width: '100%' }} />
                </Form.Item>
              </Col>
              <Col xs={24} md={12}>
                <Form.Item name="end_time" label="结束时间" rules={[{ required: true, message: '必填' }]}>
                  <DatePicker showTime style={{ width: '100%' }} />
                </Form.Item>
              </Col>
            </Row>
          </Card>
        </Form>
      </Modal>

      <Modal
        title="用户端预览（限时秒杀卖场）"
        open={previewOpen}
        onCancel={() => {
          setPreviewOpen(false);
          setPreview(null);
        }}
        footer={null}
        width={640}
        destroyOnClose
      >
        {previewLoading ? (
          <Spin />
        ) : preview ? (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            {previewCtx.isDraft ? (
              <Alert
                type="warning"
                showIcon
                message="草稿预览"
                description="展示为当前表单未保存状态；保存并上架后，用户端以接口数据为准。"
              />
            ) : null}
            {previewCtx.warnings.length > 0 ? (
              <Alert
                type="info"
                showIcon
                message="用户端可见性提示"
                description={
                  <ul style={{ margin: 0, paddingLeft: 18 }}>
                    {previewCtx.warnings.map((t) => (
                      <li key={t}>{t}</li>
                    ))}
                  </ul>
                }
              />
            ) : null}
            <Card
              size="small"
              style={{ borderRadius: 12 }}
              title={
                <Space wrap>
                  <span>{preview.sale.name}</span>
                  {preview.sale.is_featured ? <Tag color="gold">推荐</Tag> : null}
                  {preview.sale.badge_text ? <Tag color="purple">{preview.sale.badge_text}</Tag> : null}
                  {preview.sale.badge_text_secondary ? (
                    <Tag color="cyan">{preview.sale.badge_text_secondary}</Tag>
                  ) : null}
                  {preview.sale.promo_label ? <Tag color="blue">{preview.sale.promo_label}</Tag> : null}
                </Space>
              }
            >
              {preview.sale.marketing_line ? (
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  {preview.sale.marketing_line}
                </Paragraph>
              ) : null}
              {preview.sale.description ? (
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  {preview.sale.description}
                </Paragraph>
              ) : null}
              <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                场次：{dayjs(preview.sale.start_time).format('YYYY-MM-DD HH:mm')} —{' '}
                {dayjs(preview.sale.end_time).format('YYYY-MM-DD HH:mm')}
              </Paragraph>
              {preview.sale.promo_ends_at ? (
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  活动展示结束：{dayjs(preview.sale.promo_ends_at).format('YYYY-MM-DD HH:mm')}
                </Paragraph>
              ) : null}
              <Paragraph type="secondary" style={{ marginBottom: 4 }}>
                秒杀 SKU（与卖场卡片字段一致）
              </Paragraph>
              <Table<FlashSaleProductRow>
                rowKey="id"
                size="small"
                columns={previewSkuCols}
                dataSource={preview.skus}
                pagination={false}
              />
            </Card>
            <Text type="secondary" style={{ fontSize: 12 }}>
              前台路径：/catalog?flash=true · 接口：GET /flash-sales/active、/upcoming
            </Text>
          </Space>
        ) : null}
      </Modal>
    </div>
  );
};

export default AdminFlashSales;
