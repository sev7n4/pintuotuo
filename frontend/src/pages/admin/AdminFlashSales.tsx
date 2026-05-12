import { useCallback, useEffect, useRef, useState } from 'react';
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
  Select,
  Spin,
} from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import api from '@/services/api';
import { skuService } from '@/services/sku';
import type { SKUWithSPU } from '@/types/sku';

const { Title } = Typography;

type SkuLine = {
  sku_id?: number;
  flash_price?: number;
  stock_limit?: number;
  per_user_limit?: number;
};

function FlashSkuSelect({
  value,
  onChange,
}: {
  value?: number;
  onChange?: (v: number | undefined) => void;
}) {
  const [options, setOptions] = useState<{ label: string; value: number }[]>([]);
  const [fetching, setFetching] = useState(false);
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const loadById = useCallback(async (id: number) => {
    setFetching(true);
    try {
      const res = await skuService.getSKU(id);
      const sku = res.data.data as SKUWithSPU;
      const label = `${sku.id} · ${sku.spu_name || '—'} · ${sku.sku_code}（库存 ${sku.stock ?? '—'}）`;
      setOptions((prev) => {
        const rest = prev.filter((o) => o.value !== id);
        return [{ value: id, label }, ...rest];
      });
    } catch {
      setOptions((prev) => {
        const rest = prev.filter((o) => o.value !== id);
        return [{ value: id, label: `SKU #${id}（加载失败）` }, ...rest];
      });
    } finally {
      setFetching(false);
    }
  }, []);

  useEffect(() => {
    if (value && value > 0) {
      void loadById(value);
    }
  }, [value, loadById]);

  const onSearch = (q: string) => {
    const term = q.trim();
    if (searchTimer.current) clearTimeout(searchTimer.current);
    if (!term) {
      return;
    }
    searchTimer.current = setTimeout(async () => {
      setFetching(true);
      try {
        const res = await skuService.getSKUs({
          q: term,
          page: 1,
          per_page: 40,
          scope: 'all',
          status: 'all',
        });
        const rows = res.data.data || [];
        setOptions(
          rows.map((r) => ({
            value: r.id,
            label: `${r.id} · ${r.spu_name || '—'} · ${r.sku_code}（库存 ${r.stock ?? '—'}）`,
          }))
        );
      } catch {
        setOptions([]);
      } finally {
        setFetching(false);
      }
    }, 320);
  };

  return (
    <Select
      showSearch
      allowClear
      filterOption={false}
      value={value}
      onChange={(v) => onChange?.(v ?? undefined)}
      onSearch={onSearch}
      notFoundContent={fetching ? <Spin size="small" /> : '输入名称或编码搜索'}
      placeholder="搜索 SKU 名称 / 编码 / ID"
      style={{ minWidth: 280, maxWidth: '100%' }}
      options={options}
    />
  );
}

const AdminFlashSales = () => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const isStrictModelSKU = (provider?: string, modelName?: string, providerModelID?: string) => {
    const p = String(provider || '')
      .trim()
      .toLowerCase();
    if (!p || p === 'internal' || p === 'virtual_goods') return false;
    return Boolean(String(providerModelID || '').trim() || String(modelName || '').trim());
  };

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

    const rows = values.skus || [];
    const skuIds = rows.map((r) => Number(r.sku_id)).filter((id) => id > 0);
    const uniq = new Set(skuIds);
    if (uniq.size !== skuIds.length) {
      message.warning('同一活动中不可重复选择相同 SKU');
      return;
    }

    setLoading(true);
    try {
      for (const row of rows) {
        const skuID = Number(row.sku_id);
        const res = await skuService.getSKU(skuID);
        const sku = res.data.data as SKUWithSPU;
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
      const skus = rows.map((row) => ({
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
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      message.error(err?.response?.data?.message || '创建失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={2}>秒杀配置</Title>
      <Alert
        type="warning"
        showIcon
        style={{ marginBottom: 12 }}
        message="加油包限制"
        description="加油包不可单独购买。秒杀活动中若配置纯加油包 SKU，会在提交前拦截并提示。"
      />
      <Card>
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
                      label="SKU"
                      rules={[{ required: true, message: '请选择 SKU' }]}
                    >
                      <FlashSkuSelect />
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
    </div>
  );
};

export default AdminFlashSales;
