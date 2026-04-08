import { useEffect } from 'react';
import { Drawer, Form, Input, Select, InputNumber, Switch, Button, Space } from 'antd';
import type { Category } from '@/types';

export interface CatalogFilterValues {
  q?: string;
  category?: string;
  model_name?: string;
  provider?: string;
  tier?: string;
  sku_type?: string;
  group_enabled?: boolean;
  price_min?: number | null;
  price_max?: number | null;
  valid_days_min?: number | null;
  valid_days_max?: number | null;
  sort?: string;
}

interface CatalogFilterDrawerProps {
  open: boolean;
  onClose: () => void;
  categories: Category[];
  initialValues: CatalogFilterValues;
  onApply: (values: CatalogFilterValues) => void;
}

export function CatalogFilterDrawer({
  open,
  onClose,
  categories,
  initialValues,
  onApply,
}: CatalogFilterDrawerProps) {
  const [form] = Form.useForm<CatalogFilterValues>();

  useEffect(() => {
    if (open) {
      form.setFieldsValue({
        group_enabled: false,
        ...initialValues,
      });
    }
  }, [open, form, initialValues]);

  const handleFinish = (values: CatalogFilterValues) => {
    onApply(values);
    onClose();
  };

  return (
    <Drawer
      title="筛选条件"
      placement="right"
      width={360}
      open={open}
      onClose={onClose}
      destroyOnClose
    >
      <Form form={form} layout="vertical" onFinish={handleFinish}>
        <Form.Item name="q" label="关键词">
          <Input placeholder="模型名、套餐名、SKU 编码" allowClear />
        </Form.Item>
        <Form.Item name="category" label="品类 / SPU 名称">
          <Select
            allowClear
            placeholder="选择分类"
            showSearch
            optionFilterProp="label"
            options={categories.map((c) => ({ label: c.name, value: c.name }))}
          />
        </Form.Item>
        <Space style={{ width: '100%' }} size="middle" wrap>
          <Form.Item name="price_min" label="最低价（元）">
            <InputNumber min={0} style={{ width: 140 }} placeholder="元" />
          </Form.Item>
          <Form.Item name="price_max" label="最高价（元）">
            <InputNumber min={0} style={{ width: 140 }} placeholder="元" />
          </Form.Item>
        </Space>
        <Form.Item name="model_name" label="模型名称">
          <Input placeholder="如 GLM、DeepSeek" allowClear />
        </Form.Item>
        <Form.Item name="provider" label="厂商">
          <Select
            allowClear
            placeholder="厂商"
            options={[
              { label: 'OpenAI', value: 'openai' },
              { label: 'Anthropic', value: 'anthropic' },
              { label: 'Google', value: 'google' },
              { label: 'DeepSeek', value: 'deepseek' },
              { label: '智谱', value: 'zhipu' },
            ]}
          />
        </Form.Item>
        <Form.Item name="tier" label="模型层级">
          <Select
            allowClear
            options={[
              { label: 'Pro', value: 'pro' },
              { label: 'Lite', value: 'lite' },
              { label: 'Mini', value: 'mini' },
              { label: 'Vision', value: 'vision' },
            ]}
          />
        </Form.Item>
        <Form.Item name="sku_type" label="套餐类型">
          <Select
            allowClear
            options={[
              { label: 'Token包', value: 'token_pack' },
              { label: '订阅', value: 'subscription' },
              { label: '并发', value: 'concurrent' },
              { label: '试用', value: 'trial' },
            ]}
          />
        </Form.Item>
        <Form.Item name="group_enabled" label="仅看支持拼团" valuePropName="checked">
          <Switch />
        </Form.Item>
        <Space style={{ width: '100%' }} size="middle" wrap>
          <Form.Item name="valid_days_min" label="有效期≥(天)">
            <InputNumber min={0} style={{ width: 140 }} />
          </Form.Item>
          <Form.Item name="valid_days_max" label="有效期≤(天)">
            <InputNumber min={0} style={{ width: 140 }} />
          </Form.Item>
        </Space>
        <Form.Item name="sort" label="排序">
          <Select
            allowClear
            placeholder="默认推荐"
            options={[
              { label: '热销', value: 'hot' },
              { label: '最新', value: 'new' },
              { label: '价格从低到高', value: 'price_asc' },
              { label: '价格从高到低', value: 'price_desc' },
            ]}
          />
        </Form.Item>
        <Space wrap>
          <Button type="primary" htmlType="submit">
            应用筛选
          </Button>
          <Button
            onClick={() => {
              form.resetFields();
            }}
          >
            重置表单
          </Button>
        </Space>
      </Form>
    </Drawer>
  );
}
