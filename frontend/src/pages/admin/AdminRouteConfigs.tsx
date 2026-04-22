import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  message,
  Tag,
  Tabs,
  Row,
  Col,
  Divider,
  Alert,
  Tooltip,
  InputNumber,
} from 'antd';
import { EditOutlined, ApiOutlined, GlobalOutlined, ExperimentOutlined } from '@ant-design/icons';
import routeConfigService, {
  ProviderRouteConfig,
  MerchantRouteConfig,
  RouteTestResult,
} from '@services/routeConfig';

const { TabPane } = Tabs;

const AdminRouteConfigs: React.FC = () => {
  const [providers, setProviders] = useState<ProviderRouteConfig[]>([]);
  const [merchants, setMerchants] = useState<MerchantRouteConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [providerModalVisible, setProviderModalVisible] = useState(false);
  const [merchantModalVisible, setMerchantModalVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<ProviderRouteConfig | null>(null);
  const [editingMerchant, setEditingMerchant] = useState<MerchantRouteConfig | null>(null);
  const [testResult, setTestResult] = useState<RouteTestResult | null>(null);
  const [providerForm] = Form.useForm();
  const [merchantForm] = Form.useForm();
  const [testForm] = Form.useForm();

  useEffect(() => {
    fetchProviders();
    fetchMerchants();
  }, []);

  const fetchProviders = async () => {
    setLoading(true);
    try {
      const data = await routeConfigService.getProviderRouteConfigs();
      setProviders(data);
    } catch (error) {
      message.error('获取厂商配置失败');
    } finally {
      setLoading(false);
    }
  };

  const fetchMerchants = async () => {
    try {
      const data = await routeConfigService.getMerchantRouteConfigs();
      setMerchants(data);
    } catch (error) {
      message.error('获取商户配置失败');
    }
  };

  const handleEditProvider = (record: ProviderRouteConfig) => {
    setEditingProvider(record);
    providerForm.setFieldsValue({
      provider_region: record.provider_region,
      route_strategy: JSON.stringify(record.route_strategy || {}, null, 2),
      endpoints: JSON.stringify(record.endpoints || {}, null, 2),
    });
    setProviderModalVisible(true);
  };

  const handleSaveProvider = async () => {
    if (!editingProvider) return;
    try {
      const values = await providerForm.validateFields();
      const routeStrategy = JSON.parse(values.route_strategy || '{}');
      const endpoints = JSON.parse(values.endpoints || '{}');

      await routeConfigService.updateProviderRouteConfig(editingProvider.code, {
        provider_region: values.provider_region,
        route_strategy: routeStrategy,
        endpoints: endpoints,
      });
      message.success('更新成功');
      setProviderModalVisible(false);
      fetchProviders();
    } catch (error) {
      message.error('更新失败，请检查 JSON 格式');
    }
  };

  const handleEditMerchant = (record: MerchantRouteConfig) => {
    setEditingMerchant(record);
    merchantForm.setFieldsValue({
      merchant_type: record.merchant_type,
      region: record.region,
      route_preference: JSON.stringify(record.route_preference || {}, null, 2),
    });
    setMerchantModalVisible(true);
  };

  const handleSaveMerchant = async () => {
    if (!editingMerchant) return;
    try {
      const values = await merchantForm.validateFields();
      const routePreference = JSON.parse(values.route_preference || '{}');

      await routeConfigService.updateMerchantRouteConfig(editingMerchant.id, {
        merchant_type: values.merchant_type,
        region: values.region,
        route_preference: routePreference,
      });
      message.success('更新成功');
      setMerchantModalVisible(false);
      fetchMerchants();
    } catch (error) {
      message.error('更新失败，请检查 JSON 格式');
    }
  };

  const handleTestRoute = async () => {
    try {
      const values = await testForm.validateFields();
      const result = await routeConfigService.testRouteDecision(
        values.provider_code,
        values.merchant_id
      );
      setTestResult(result);
    } catch (error) {
      message.error('测试失败');
    }
  };

  const providerColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '厂商代码',
      dataIndex: 'code',
      key: 'code',
    },
    {
      title: '厂商名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '区域',
      dataIndex: 'provider_region',
      key: 'provider_region',
      render: (region: string) => (
        <Tag color={region === 'overseas' ? 'blue' : 'green'} icon={<GlobalOutlined />}>
          {region === 'overseas' ? '海外' : '国内'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: ProviderRouteConfig) => (
        <Button type="link" icon={<EditOutlined />} onClick={() => handleEditProvider(record)}>
          编辑
        </Button>
      ),
    },
  ];

  const merchantColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
    },
    {
      title: '商户名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '商户类型',
      dataIndex: 'merchant_type',
      key: 'merchant_type',
      render: (type: string) => (
        <Tag color={type === 'enterprise' ? 'gold' : 'default'}>
          {type === 'enterprise' ? '企业' : '标准'}
        </Tag>
      ),
    },
    {
      title: '区域',
      dataIndex: 'region',
      key: 'region',
      render: (region: string) => (
        <Tag color={region === 'overseas' ? 'blue' : 'green'} icon={<GlobalOutlined />}>
          {region === 'overseas' ? '海外' : '国内'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_: any, record: MerchantRouteConfig) => (
        <Button type="link" icon={<EditOutlined />} onClick={() => handleEditMerchant(record)}>
          编辑
        </Button>
      ),
    },
  ];

  return (
    <Card title="统一路由配置管理">
      <Alert
        message="统一路由配置说明"
        description="通过配置厂商区域、路由策略和端点信息，实现基于商户类型和区域的智能路由决策。系统会自动选择最优路由模式（直连/LiteLLM/代理）。"
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
      />

      <Tabs defaultActiveKey="providers">
        <TabPane tab="厂商配置" key="providers">
          <Table
            columns={providerColumns}
            dataSource={providers}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 20 }}
          />
        </TabPane>

        <TabPane tab="商户配置" key="merchants">
          <Table
            columns={merchantColumns}
            dataSource={merchants}
            rowKey="id"
            loading={loading}
            pagination={{ pageSize: 20 }}
          />
        </TabPane>

        <TabPane
          tab={
            <span>
              <ExperimentOutlined />
              路由测试
            </span>
          }
          key="test"
        >
          <Card>
            <Form form={testForm} layout="inline">
              <Form.Item
                name="provider_code"
                label="厂商代码"
                rules={[{ required: true, message: '请输入厂商代码' }]}
              >
                <Input placeholder="如: openai" style={{ width: 150 }} />
              </Form.Item>
              <Form.Item
                name="merchant_id"
                label="商户ID"
                rules={[{ required: true, message: '请输入商户ID' }]}
              >
                <InputNumber placeholder="商户ID" style={{ width: 150 }} />
              </Form.Item>
              <Form.Item>
                <Button type="primary" onClick={handleTestRoute}>
                  测试路由
                </Button>
              </Form.Item>
            </Form>

            {testResult && (
              <div style={{ marginTop: 24 }}>
                <Divider>测试结果</Divider>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <Card size="small" title="路由决策">
                      <p>
                        <strong>路由模式:</strong> <Tag color="blue">{testResult.mode}</Tag>
                      </p>
                      <p>
                        <strong>端点:</strong> {testResult.endpoint || '-'}
                      </p>
                      <p>
                        <strong>降级模式:</strong> {testResult.fallback_mode || '-'}
                      </p>
                      <p>
                        <strong>降级端点:</strong> {testResult.fallback_endpoint || '-'}
                      </p>
                      <p>
                        <strong>决策原因:</strong> {testResult.reason}
                      </p>
                    </Card>
                  </Col>
                  <Col span={12}>
                    <Card size="small" title="厂商配置">
                      <pre style={{ fontSize: 12, maxHeight: 200, overflow: 'auto' }}>
                        {JSON.stringify(testResult.provider_config, null, 2)}
                      </pre>
                    </Card>
                  </Col>
                  <Col span={24}>
                    <Card size="small" title="商户配置">
                      <pre style={{ fontSize: 12, maxHeight: 200, overflow: 'auto' }}>
                        {JSON.stringify(testResult.merchant_config, null, 2)}
                      </pre>
                    </Card>
                  </Col>
                </Row>
              </div>
            )}
          </Card>
        </TabPane>
      </Tabs>

      <Modal
        title={`编辑厂商配置 - ${editingProvider?.name}`}
        open={providerModalVisible}
        onOk={handleSaveProvider}
        onCancel={() => setProviderModalVisible(false)}
        width={800}
      >
        <Form form={providerForm} layout="vertical">
          <Form.Item
            name="provider_region"
            label="厂商区域"
            rules={[{ required: true, message: '请选择区域' }]}
          >
            <Select
              options={[
                { value: 'domestic', label: '国内' },
                { value: 'overseas', label: '海外' },
              ]}
            />
          </Form.Item>
          <Form.Item
            name="route_strategy"
            label={
              <span>
                路由策略 (JSON)
                <Tooltip title='配置不同用户类型的路由策略，如: {"domestic_users": {"mode": "litellm"}}'>
                  <ApiOutlined style={{ marginLeft: 8, color: '#999' }} />
                </Tooltip>
              </span>
            }
          >
            <Input.TextArea rows={8} placeholder='{"domestic_users": {"mode": "litellm"}}' />
          </Form.Item>
          <Form.Item
            name="endpoints"
            label={
              <span>
                端点配置 (JSON)
                <Tooltip title="配置不同路由模式的端点URL">
                  <ApiOutlined style={{ marginLeft: 8, color: '#999' }} />
                </Tooltip>
              </span>
            }
          >
            <Input.TextArea
              rows={8}
              placeholder='{"litellm": {"domestic": "http://litellm:4000/v1"}}'
            />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`编辑商户配置 - ${editingMerchant?.name}`}
        open={merchantModalVisible}
        onOk={handleSaveMerchant}
        onCancel={() => setMerchantModalVisible(false)}
        width={800}
      >
        <Form form={merchantForm} layout="vertical">
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="merchant_type"
                label="商户类型"
                rules={[{ required: true, message: '请选择类型' }]}
              >
                <Select
                  options={[
                    { value: 'standard', label: '标准' },
                    { value: 'enterprise', label: '企业' },
                  ]}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="region"
                label="商户区域"
                rules={[{ required: true, message: '请选择区域' }]}
              >
                <Select
                  options={[
                    { value: 'domestic', label: '国内' },
                    { value: 'overseas', label: '海外' },
                  ]}
                />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item
            name="route_preference"
            label={
              <span>
                路由偏好 (JSON)
                <Tooltip title="配置商户的路由偏好设置">
                  <ApiOutlined style={{ marginLeft: 8, color: '#999' }} />
                </Tooltip>
              </span>
            }
          >
            <Input.TextArea rows={6} placeholder='{"preferred_mode": "litellm"}' />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
};

export default AdminRouteConfigs;
