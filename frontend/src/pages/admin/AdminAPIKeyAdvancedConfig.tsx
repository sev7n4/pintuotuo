import React, { useEffect, useState } from 'react';
import { Card, Table, Button, Form, Input, Select, Tag, Space, Modal, message, Descriptions } from 'antd';
import { EditOutlined, SaveOutlined } from '@ant-design/icons';
import api from '@/services/api';

interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface APIKey {
  id: number;
  name: string;
  provider: string;
  status: string;
  region: string;
  security_level: string;
  route_preference: RoutePreference | null;
}

interface RoutePreference {
  strategy: 'performance' | 'price' | 'reliability' | 'balanced' | 'security';
  weight: number;
  max_retries: number;
  timeout_ms: number;
  circuit_break_threshold: number;
  load_balance_weight: number;
}

const defaultRoutePreference: RoutePreference = {
  strategy: 'balanced',
  weight: 1,
  max_retries: 3,
  timeout_ms: 30000,
  circuit_break_threshold: 0.5,
  load_balance_weight: 1,
};

const AdminAPIKeyAdvancedConfig: React.FC = () => {
  const [apiKeys, setApiKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [currentKey, setCurrentKey] = useState<APIKey | null>(null);
  const [form] = Form.useForm();

  const fetchAPIKeys = async () => {
    try {
      setLoading(true);
      const response = await api.get<APIResponse<APIKey[]>>('/admin/merchants/api-keys');
      if (response.data && response.data.code === 0) {
        setApiKeys(response.data.data || []);
      }
    } catch (error) {
      message.error('获取API Key列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAPIKeys();
  }, []);

  const handleEdit = (apiKey: APIKey) => {
    setCurrentKey(apiKey);
    const routePref = apiKey.route_preference || defaultRoutePreference;
    form.setFieldsValue({
      region: apiKey.region || 'domestic',
      security_level: apiKey.security_level || 'standard',
      strategy: routePref.strategy,
      weight: routePref.weight,
      max_retries: routePref.max_retries,
      timeout_ms: routePref.timeout_ms,
      circuit_break_threshold: routePref.circuit_break_threshold,
      load_balance_weight: routePref.load_balance_weight,
    });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (!currentKey) return;

      const patch = {
        region: values.region,
        security_level: values.security_level,
        route_preference: {
          strategy: values.strategy,
          weight: Number(values.weight),
          max_retries: Number(values.max_retries),
          timeout_ms: Number(values.timeout_ms),
          circuit_break_threshold: Number(values.circuit_break_threshold),
          load_balance_weight: Number(values.load_balance_weight),
        },
      };

      const response = await api.patch<APIResponse<unknown>>(`/admin/merchants/api-keys/${currentKey.id}`, patch);
      if (response.data.code === 0) {
        message.success('配置更新成功');
        setModalVisible(false);
        fetchAPIKeys();
      }
    } catch (error) {
      message.error('配置更新失败');
    }
  };

  const columns = [
    {
      title: 'API Key ID',
      dataIndex: 'id',
      key: 'id',
      width: 100,
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 150,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      key: 'provider',
      width: 120,
      render: (provider: string) => <Tag color="blue">{provider}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 80,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      width: 100,
      render: (region: string) => (
        <Tag color="orange">{region === 'domestic' ? '国内' : '海外'}</Tag>
      ),
    },
    {
      title: '安全等级',
      dataIndex: 'security_level',
      key: 'security_level',
      width: 100,
      render: (level: string) => (
        <Tag color={level === 'high' ? 'red' : 'default'}>
          {level === 'high' ? '高' : '标准'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: APIKey) => (
        <Space>
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            配置
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <SaveOutlined />
          API Key 高级配置
        </Space>
      }
    >
      <Table
        columns={columns}
        dataSource={apiKeys}
        rowKey="id"
        loading={loading}
        pagination={{
          pageSize: 20,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        scroll={{ x: 'max-content' }}
      />

      <Modal
        title="API Key 高级配置"
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleSubmit}
        width={600}
      >
        {currentKey && (
          <>
            <Descriptions bordered column={2} style={{ marginBottom: 24 }}>
              <Descriptions.Item label="API Key ID">{currentKey.id}</Descriptions.Item>
              <Descriptions.Item label="名称">{currentKey.name}</Descriptions.Item>
              <Descriptions.Item label="提供商">{currentKey.provider}</Descriptions.Item>
              <Descriptions.Item label="状态">
                <Tag color={currentKey.status === 'active' ? 'green' : 'default'}>
                  {currentKey.status === 'active' ? '启用' : '禁用'}
                </Tag>
              </Descriptions.Item>
            </Descriptions>

            <Form form={form} layout="vertical">
              <Form.Item
                name="region"
                label="区域"
                rules={[{ required: true, message: '请选择区域' }]}
              >
                <Select placeholder="选择区域">
                  <Select.Option value="domestic">国内</Select.Option>
                  <Select.Option value="overseas">海外</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="security_level"
                label="安全等级"
                rules={[{ required: true, message: '请选择安全等级' }]}
              >
                <Select placeholder="选择安全等级">
                  <Select.Option value="standard">标准</Select.Option>
                  <Select.Option value="high">高安全</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="strategy"
                label="路由策略"
                rules={[{ required: true, message: '请选择路由策略' }]}
              >
                <Select placeholder="选择路由策略">
                  <Select.Option value="performance">性能优先</Select.Option>
                  <Select.Option value="price">价格优先</Select.Option>
                  <Select.Option value="reliability">可靠性优先</Select.Option>
                  <Select.Option value="balanced">均衡策略</Select.Option>
                  <Select.Option value="security">安全优先</Select.Option>
                </Select>
              </Form.Item>

              <Form.Item
                name="weight"
                label="策略权重"
                rules={[{ required: true, message: '请输入策略权重' }]}
              >
                <Input type="number" min={0} max={10} step={0.1} placeholder="输入策略权重 (0-10)" />
              </Form.Item>

              <Form.Item
                name="max_retries"
                label="最大重试次数"
                rules={[{ required: true, message: '请输入最大重试次数' }]}
              >
                <Input type="number" min={0} max={10} step={1} placeholder="输入最大重试次数 (0-10)" />
              </Form.Item>

              <Form.Item
                name="timeout_ms"
                label="超时时间 (ms)"
                rules={[{ required: true, message: '请输入超时时间' }]}
              >
                <Input type="number" min={1000} max={60000} step={1000} placeholder="输入超时时间 (毫秒)" />
              </Form.Item>

              <Form.Item
                name="circuit_break_threshold"
                label="熔断阈值"
                rules={[{ required: true, message: '请输入熔断阈值' }]}
              >
                <Input type="number" min={0} max={1} step={0.1} placeholder="输入熔断阈值 (0-1)" />
              </Form.Item>

              <Form.Item
                name="load_balance_weight"
                label="负载均衡权重"
                rules={[{ required: true, message: '请输入负载均衡权重' }]}
              >
                <Input type="number" min={0} max={10} step={0.1} placeholder="输入负载均衡权重 (0-10)" />
              </Form.Item>
            </Form>
          </>
        )}
      </Modal>
    </Card>
  );
};

export default AdminAPIKeyAdvancedConfig;