import { useEffect, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tag,
  DatePicker,
  Alert,
  Typography,
} from 'antd';
import { PlusOutlined, DeleteOutlined, EditOutlined, EyeOutlined } from '@ant-design/icons';
import { entitlementPackageService } from '@/services/entitlementPackage';
import { skuService } from '@/services/sku';
import type { EntitlementPackage } from '@/types/entitlementPackage';
import type { SKUWithSPU } from '@/types/sku';
import dayjs from 'dayjs';

const { Paragraph, Text } = Typography;

type FormValues = {
  package_code?: string;
  name: string;
  description?: string;
  status: 'active' | 'inactive';
  sort_order: number;
  start_at?: dayjs.Dayjs;
  end_at?: dayjs.Dayjs;
  is_featured?: boolean;
  badge_text?: string;
  items: Array<{ sku_id: number; default_quantity: number }>;
};

type PreviewContext = {
  /** 草稿预览（未保存） */
  isDraft: boolean;
  /** 列表行预览（已落库） */
  warnings: string[];
};

function packageTotalPrice(pkg: EntitlementPackage): number {
  return (pkg.items || []).reduce(
    (sum, it) => sum + Number(it.retail_price || 0) * Number(it.default_quantity || 1),
    0
  );
}

function collectPackagePreviewWarnings(pkg: EntitlementPackage): string[] {
  const w: string[] = [];
  if (pkg.status !== 'active') {
    w.push('当前状态非「在售」，用户端 /entitlement-packages 不会展示该包。');
  }
  const now = dayjs();
  if (pkg.start_at && dayjs(pkg.start_at).isAfter(now)) {
    w.push(
      `未到上架时间（开始：${dayjs(pkg.start_at).format('YYYY-MM-DD HH:mm')}），用户端暂不展示。`
    );
  }
  if (pkg.end_at && !dayjs(pkg.end_at).isAfter(now)) {
    w.push(
      `已过下架时间（结束：${dayjs(pkg.end_at).format('YYYY-MM-DD HH:mm')}），用户端暂不展示。`
    );
  }
  return w;
}

function buildDraftPackagePreview(v: FormValues, skuList: SKUWithSPU[]): EntitlementPackage {
  const byId = new Map(skuList.map((s) => [s.id, s]));
  const items = (v.items || []).map((line, idx) => {
    const s = byId.get(line.sku_id);
    return {
      id: -(idx + 1),
      sku_id: line.sku_id,
      sku_code: s?.sku_code ?? '—',
      spu_name: s?.spu_name ?? '（请选择有效 SKU）',
      sku_type: s?.sku_type ?? '',
      default_quantity: line.default_quantity,
      retail_price: s != null ? Number(s.retail_price) : 0,
    };
  });
  return {
    id: 0,
    package_code: v.package_code?.trim() || '（未保存）',
    name: v.name?.trim() || '（未命名）',
    description: v.description?.trim(),
    status: v.status,
    sort_order: v.sort_order ?? 0,
    is_featured: !!v.is_featured,
    badge_text: v.badge_text?.trim(),
    start_at: v.start_at?.toISOString(),
    end_at: v.end_at?.toISOString(),
    items,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

export default function AdminEntitlementPackages() {
  const [loading, setLoading] = useState(false);
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<EntitlementPackage | null>(null);
  const [rows, setRows] = useState<EntitlementPackage[]>([]);
  const [skus, setSkus] = useState<SKUWithSPU[]>([]);
  const [form] = Form.useForm<FormValues>();
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewPkg, setPreviewPkg] = useState<EntitlementPackage | null>(null);
  const [previewCtx, setPreviewCtx] = useState<PreviewContext>({ isDraft: false, warnings: [] });

  const loadData = async () => {
    setLoading(true);
    try {
      const [pkgRes, skuRes] = await Promise.all([
        entitlementPackageService.listAdmin(),
        skuService.getSKUs({ per_page: 200, scope: 'all' }),
      ]);
      setRows(pkgRes.data.data || []);
      setSkus(skuRes.data.data || []);
    } catch {
      message.error('加载权益包数据失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const openCreate = () => {
    setEditing(null);
    form.resetFields();
    form.setFieldsValue({
      status: 'active',
      sort_order: 0,
      is_featured: false,
      items: [{ sku_id: 0, default_quantity: 1 }],
    });
    setOpen(true);
  };

  const openEdit = (row: EntitlementPackage) => {
    setEditing(row);
    form.setFieldsValue({
      package_code: row.package_code,
      name: row.name,
      description: row.description,
      status: row.status,
      sort_order: row.sort_order,
      start_at: row.start_at ? dayjs(row.start_at) : undefined,
      end_at: row.end_at ? dayjs(row.end_at) : undefined,
      is_featured: row.is_featured,
      badge_text: row.badge_text,
      items: (row.items || []).map((it) => ({
        sku_id: it.sku_id,
        default_quantity: it.default_quantity,
      })),
    });
    setOpen(true);
  };

  const submit = async () => {
    try {
      const v = await form.validateFields();
      const skuIDs = (v.items || []).map((i) => i.sku_id);
      const dedup = new Set(skuIDs);
      if (dedup.size !== skuIDs.length) {
        message.error('同一权益包内不能重复选择同一个 SKU');
        return;
      }
      if (v.start_at && v.end_at && !v.end_at.isAfter(v.start_at)) {
        message.error('结束时间必须晚于开始时间');
        return;
      }
      const payload = {
        ...v,
        start_at: v.start_at ? v.start_at.toISOString() : undefined,
        end_at: v.end_at ? v.end_at.toISOString() : undefined,
      };
      if (editing) {
        await entitlementPackageService.updateAdmin(editing.id, payload);
        message.success('权益包已更新');
      } else {
        await entitlementPackageService.createAdmin(payload);
        message.success('权益包已创建');
      }
      setOpen(false);
      loadData();
    } catch {
      // form validation or api error
    }
  };

  const remove = async (id: number) => {
    try {
      await entitlementPackageService.deleteAdmin(id);
      message.success('权益包已删除');
      loadData();
    } catch {
      message.error('删除失败');
    }
  };

  const openSavedPreview = (row: EntitlementPackage) => {
    setPreviewPkg(row);
    setPreviewCtx({ isDraft: false, warnings: collectPackagePreviewWarnings(row) });
    setPreviewOpen(true);
  };

  const openDraftPreview = async () => {
    try {
      const v = await form.validateFields();
      const skuIDs = (v.items || []).map((i) => i.sku_id);
      const dedup = new Set(skuIDs);
      if (dedup.size !== skuIDs.length) {
        message.error('同一权益包内不能重复选择同一个 SKU');
        return;
      }
      if (v.start_at && v.end_at && !v.end_at.isAfter(v.start_at)) {
        message.error('结束时间必须晚于开始时间');
        return;
      }
      const draft = buildDraftPackagePreview(v, skus);
      const missingSku = (v.items || []).some((line) => !skus.some((s) => s.id === line.sku_id));
      const extra: string[] = [];
      if (missingSku)
        extra.push('部分 SKU 在当前列表中未匹配到价格/名称，保存前请确认 SKU 选择正确。');
      setPreviewPkg(draft);
      setPreviewCtx({
        isDraft: true,
        warnings: [...collectPackagePreviewWarnings(draft), ...extra],
      });
      setPreviewOpen(true);
    } catch {
      message.info('请先完善表单必填项后再预览');
    }
  };

  return (
    <Card
      title="权益包管理"
      extra={
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
          新建权益包
        </Button>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        dataSource={rows}
        columns={[
          { title: '编码', dataIndex: 'package_code', key: 'package_code', width: 140 },
          { title: '名称', dataIndex: 'name', key: 'name', width: 220 },
          {
            title: '明细',
            key: 'items',
            render: (_, r) => (
              <Space wrap>
                {(r.items || []).map((it) => (
                  <Tag key={it.id}>
                    {it.sku_code} x{it.default_quantity}
                  </Tag>
                ))}
              </Space>
            ),
          },
          {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            width: 100,
            render: (s: string) => <Tag color={s === 'active' ? 'success' : 'default'}>{s}</Tag>,
          },
          {
            title: '运营',
            key: 'ops',
            width: 220,
            render: (_, r) => (
              <Space wrap>
                {r.is_featured ? <Tag color="gold">推荐</Tag> : null}
                {r.badge_text ? <Tag color="purple">{r.badge_text}</Tag> : null}
                {r.start_at || r.end_at ? (
                  <Tag>
                    {r.start_at ? dayjs(r.start_at).format('MM-DD HH:mm') : '不限'} ~{' '}
                    {r.end_at ? dayjs(r.end_at).format('MM-DD HH:mm') : '不限'}
                  </Tag>
                ) : (
                  <Tag>长期有效</Tag>
                )}
              </Space>
            ),
          },
          { title: '排序', dataIndex: 'sort_order', key: 'sort_order', width: 90 },
          {
            title: '操作',
            key: 'action',
            width: 240,
            render: (_, r) => (
              <Space wrap>
                <Button type="link" icon={<EyeOutlined />} onClick={() => openSavedPreview(r)}>
                  预览
                </Button>
                <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(r)}>
                  编辑
                </Button>
                <Popconfirm title="确认删除该权益包？" onConfirm={() => remove(r.id)}>
                  <Button type="link" danger icon={<DeleteOutlined />}>
                    删除
                  </Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal
        open={open}
        title={editing ? '编辑权益包' : '新建权益包'}
        onCancel={() => setOpen(false)}
        onOk={submit}
        width={860}
        footer={[
          <Button key="preview" icon={<EyeOutlined />} onClick={openDraftPreview}>
            预览用户端
          </Button>,
          <Button key="cancel" onClick={() => setOpen(false)}>
            取消
          </Button>,
          <Button key="ok" type="primary" onClick={submit}>
            保存
          </Button>,
        ]}
      >
        <Form form={form} layout="vertical">
          <Alert
            type="info"
            showIcon
            style={{ marginBottom: 12 }}
            message="配置建议：同一权益包内避免重复 SKU；可通过时间窗控制活动期，通过推荐/角标提升曝光。"
          />
          <Form.Item
            name="package_code"
            label="包编码"
            rules={[{ required: !editing, message: '请输入包编码' }]}
          >
            <Input placeholder="如 P1_CODING" disabled={!!editing} />
          </Form.Item>
          <Form.Item
            name="name"
            label="包名称"
            rules={[{ required: true, message: '请输入包名称' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Space style={{ width: '100%' }} size={16}>
            <Form.Item name="status" label="状态" rules={[{ required: true }]}>
              <Select
                style={{ width: 180 }}
                options={[{ value: 'active' }, { value: 'inactive' }]}
              />
            </Form.Item>
            <Form.Item name="sort_order" label="排序">
              <InputNumber min={0} precision={0} />
            </Form.Item>
            <Form.Item name="is_featured" label="推荐位">
              <Select
                style={{ width: 140 }}
                options={[
                  { label: '否', value: false },
                  { label: '是', value: true },
                ]}
              />
            </Form.Item>
          </Space>
          <Space style={{ width: '100%' }} size={16}>
            <Form.Item name="badge_text" label="角标文案">
              <Input placeholder="如 限时特惠 / 新客推荐" />
            </Form.Item>
            <Form.Item name="start_at" label="开始时间">
              <DatePicker showTime />
            </Form.Item>
            <Form.Item name="end_at" label="结束时间">
              <DatePicker showTime />
            </Form.Item>
          </Space>

          <Form.List name="items">
            {(fields, { add, remove: removeField }) => (
              <>
                {fields.map((f) => (
                  <Space key={f.key} align="baseline" style={{ display: 'flex', marginBottom: 8 }}>
                    <Form.Item
                      name={[f.name, 'sku_id']}
                      rules={[{ required: true, message: '请选择SKU' }]}
                      style={{ minWidth: 420 }}
                    >
                      <Select
                        showSearch
                        optionFilterProp="label"
                        placeholder="选择SKU"
                        options={skus.map((s) => ({
                          value: s.id,
                          label: `${s.sku_code} / ${s.spu_name}`,
                        }))}
                      />
                    </Form.Item>
                    <Form.Item
                      name={[f.name, 'default_quantity']}
                      rules={[{ required: true, message: '数量' }]}
                    >
                      <InputNumber min={1} precision={0} />
                    </Form.Item>
                    <Button type="link" danger onClick={() => removeField(f.name)}>
                      删除
                    </Button>
                  </Space>
                ))}
                <Button type="dashed" onClick={() => add({ sku_id: 0, default_quantity: 1 })}>
                  添加SKU明细
                </Button>
              </>
            )}
          </Form.List>
        </Form>
      </Modal>

      <Modal
        title="用户端预览（权益包卡片）"
        open={previewOpen}
        onCancel={() => {
          setPreviewOpen(false);
          setPreviewPkg(null);
        }}
        footer={null}
        width={560}
        destroyOnClose
      >
        {previewPkg ? (
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            {previewCtx.isDraft ? (
              <Alert
                type="warning"
                showIcon
                message="草稿预览"
                description="展示内容为当前表单未保存状态，与用户保存后一致的前提是字段已填写完整。"
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
              title={previewPkg.name}
              extra={
                <Space size={4} wrap>
                  {previewPkg.is_featured ? <Tag color="gold">推荐</Tag> : null}
                  {previewPkg.badge_text ? (
                    <Tag color="purple">{previewPkg.badge_text}</Tag>
                  ) : (
                    <Tag color="blue">权益包</Tag>
                  )}
                </Space>
              }
            >
              {previewPkg.description ? (
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  {previewPkg.description}
                </Paragraph>
              ) : null}
              {(previewPkg.start_at || previewPkg.end_at) && (
                <Paragraph type="secondary" style={{ marginBottom: 8 }}>
                  有效期：
                  {previewPkg.start_at
                    ? dayjs(previewPkg.start_at).format('YYYY-MM-DD HH:mm')
                    : '不限'}{' '}
                  ~{' '}
                  {previewPkg.end_at ? dayjs(previewPkg.end_at).format('YYYY-MM-DD HH:mm') : '不限'}
                </Paragraph>
              )}
              <Paragraph style={{ marginBottom: 8 }}>
                组合总价：<Text strong>¥{packageTotalPrice(previewPkg).toFixed(2)}</Text>
              </Paragraph>
              <Paragraph type="secondary" style={{ marginBottom: 4 }}>
                明细（与前台「一键组合下单」行数一致）
              </Paragraph>
              <Space wrap>
                {(previewPkg.items || []).map((it) => (
                  <Tag key={`${it.sku_id}-${it.id}`} color="green">
                    {it.spu_name} / {it.sku_code} ×{it.default_quantity}
                  </Tag>
                ))}
              </Space>
            </Card>
            <Text type="secondary" style={{ fontSize: 12 }}>
              前台路径：/packages · 接口：GET /entitlement-packages
            </Text>
          </Space>
        ) : null}
      </Modal>
    </Card>
  );
}
