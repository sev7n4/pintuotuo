import { useState } from 'react';
import { Card, Typography, Form, Input, InputNumber, Button, Space, message } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import api from '@/services/api';

const { Title } = Typography;

type SkuLine = {
  sku_id?: number;
  flash_price?: number;
  stock_limit?: number;
  per_user_limit?: number;
};

const AdminFlashSales = () => {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const onFinish = async (values: {
    name: string;
    description?: string;
    start_time: string;
    end_time: string;
    skus: SkuLine[];
  }) => {
    setLoading(true);
    try {
      const skus = (values.skus || []).map((row) => ({
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
                  <Space key={field.key} align="baseline" wrap style={{ marginBottom: 8 }}>
                    <Form.Item
                      name={[field.name, 'sku_id']}
                      label="SKU ID"
                      rules={[{ required: true, message: '必填' }]}
                    >
                      <InputNumber min={1} placeholder="SKU id" style={{ width: 120 }} />
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
                    <MinusCircleOutlined onClick={() => remove(field.name)} />
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
