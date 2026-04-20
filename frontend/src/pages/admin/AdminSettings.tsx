import { useEffect, useState } from 'react';
import {
  Card,
  Typography,
  Form,
  Switch,
  InputNumber,
  Button,
  Space,
  message,
  Alert,
  Descriptions,
} from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import api from '@services/api';

const { Title } = Typography;

interface PlatformLimits {
  interval_seconds_min: number;
  interval_seconds_max: number;
  batch_min: number;
  batch_max: number;
}

interface PlatformSettingsPayload {
  health_scheduler_enabled: boolean;
  health_scheduler_interval_seconds: number;
  health_scheduler_batch: number;
  limits?: PlatformLimits;
}

const AdminSettings = () => {
  const [form] = Form.useForm<PlatformSettingsPayload>();
  const [loading, setLoading] = useState(false);
  const [limits, setLimits] = useState<PlatformLimits | null>(null);

  const fetchSettings = async () => {
    setLoading(true);
    try {
      const { data } = await api.get<PlatformSettingsPayload & { code?: number }>(
        '/admin/platform-settings'
      );
      form.setFieldsValue({
        health_scheduler_enabled: data.health_scheduler_enabled,
        health_scheduler_interval_seconds: data.health_scheduler_interval_seconds,
        health_scheduler_batch: data.health_scheduler_batch,
      });
      if (data.limits) {
        setLimits(data.limits);
      }
    } catch {
      message.error('加载系统设置失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSettings();
    // 仅挂载时拉取；避免将 fetchSettings 放入依赖导致重复请求
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const onFinish = async (values: PlatformSettingsPayload) => {
    setLoading(true);
    try {
      await api.put('/admin/platform-settings', {
        health_scheduler_enabled: values.health_scheduler_enabled,
        health_scheduler_interval_seconds: values.health_scheduler_interval_seconds,
        health_scheduler_batch: values.health_scheduler_batch,
      });
      message.success('已保存');
      await fetchSettings();
    } catch (e: unknown) {
      const err = e as { response?: { data?: { message?: string } } };
      message.error(err.response?.data?.message || '保存失败');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={2}>系统设置</Title>

      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="主动健康探测调度"
        description="开启后，按下方间隔与批量对商户 API Key 做定时健康探测并写入健康历史。该设置仅影响健康状态调度，不替代商户侧“验证密钥”流程；真实对话仍会触发被动健康更新。"
      />

      <Card title="健康调度" loading={loading}>
        {limits && (
          <Descriptions size="small" column={2} style={{ marginBottom: 16 }}>
            <Descriptions.Item label="探测间隔允许范围">
              {limits.interval_seconds_min}～{limits.interval_seconds_max} 秒
            </Descriptions.Item>
            <Descriptions.Item label="每轮批量上限">
              {limits.batch_min}～{limits.batch_max} 个 Key
            </Descriptions.Item>
          </Descriptions>
        )}

        <Form form={form} layout="vertical" onFinish={onFinish}>
          <Form.Item name="health_scheduler_enabled" label="启用主动探测" valuePropName="checked">
            <Switch checkedChildren="开" unCheckedChildren="关" />
          </Form.Item>

          <Form.Item
            name="health_scheduler_interval_seconds"
            label="探测周期（秒）"
            rules={[{ required: true, message: '请输入周期' }]}
          >
            <InputNumber
              min={limits?.interval_seconds_min ?? 60}
              max={limits?.interval_seconds_max ?? 86400}
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item
            name="health_scheduler_batch"
            label="每轮最多探测 Key 数"
            rules={[{ required: true, message: '请输入批量' }]}
          >
            <InputNumber
              min={limits?.batch_min ?? 1}
              max={limits?.batch_max ?? 50}
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={loading}>
                保存
              </Button>
              <Button onClick={() => fetchSettings()} disabled={loading}>
                重新加载
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
};

export default AdminSettings;
