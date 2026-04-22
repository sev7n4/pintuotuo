import React, { useState, useEffect } from 'react';
import {
  Card,
  Table,
  Button,
  Space,
  message,
  Tag,
  Tabs,
  Row,
  Col,
  Divider,
  Alert,
  Tooltip,
  InputNumber,
  Input,
  Typography,
  Statistic,
} from 'antd';
import {
  EditOutlined,
  ApiOutlined,
  GlobalOutlined,
  ExperimentOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CrownOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import routeConfigService, {
  ProviderRouteConfig,
  MerchantRouteConfig,
  RouteTestResult,
} from '@services/routeConfig';
import ProviderConfigForm from '@components/admin/ProviderConfigForm';
import MerchantConfigForm from '@components/admin/MerchantConfigForm';

const { TabPane } = Tabs;
const { Text } = Typography;

const AdminRouteConfigs: React.FC = () => {
  const [providers, setProviders] = useState<ProviderRouteConfig[]>([]);
  const [merchants, setMerchants] = useState<MerchantRouteConfig[]>([]);
  const [loading, setLoading] = useState(false);
  const [providerModalVisible, setProviderModalVisible] = useState(false);
  const [merchantModalVisible, setMerchantModalVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<ProviderRouteConfig | null>(null);
  const [editingMerchant, setEditingMerchant] = useState<MerchantRouteConfig | null>(null);
  const [testResult, setTestResult] = useState<RouteTestResult | null>(null);
  const [testLoading, setTestLoading] = useState(false);
  const [saveLoading, setSaveLoading] = useState(false);
  const [testForm, setTestForm] = useState({ provider_code: '', merchant_id: 1 });

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
    setProviderModalVisible(true);
  };

  const handleSaveProvider = async (values: any) => {
    if (!editingProvider) return;
    setSaveLoading(true);
    try {
      await routeConfigService.updateProviderRouteConfig(editingProvider.code, values);
      message.success('更新成功');
      setProviderModalVisible(false);
      fetchProviders();
    } catch (error) {
      message.error('更新失败');
    } finally {
      setSaveLoading(false);
    }
  };

  const handleEditMerchant = (record: MerchantRouteConfig) => {
    setEditingMerchant(record);
    setMerchantModalVisible(true);
  };

  const handleSaveMerchant = async (values: any) => {
    if (!editingMerchant) return;
    setSaveLoading(true);
    try {
      await routeConfigService.updateMerchantRouteConfig(editingMerchant.id, values);
      message.success('更新成功');
      setMerchantModalVisible(false);
      fetchMerchants();
    } catch (error) {
      message.error('更新失败');
    } finally {
      setSaveLoading(false);
    }
  };

  const handleTestRoute = async () => {
    if (!testForm.provider_code || !testForm.merchant_id) {
      message.warning('请填写厂商代码和商户ID');
      return;
    }
    setTestLoading(true);
    try {
      const result = await routeConfigService.testRouteDecision(
        testForm.provider_code,
        testForm.merchant_id
      );
      setTestResult(result);
    } catch (error) {
      message.error('测试失败，请检查输入是否正确');
    } finally {
      setTestLoading(false);
    }
  };

  const providerColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 60,
      responsive: ['md'] as any,
    },
    {
      title: '厂商代码',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => <Tag color="blue">{code}</Tag>,
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
        <Tag
          color={status === 'active' ? 'success' : 'default'}
          icon={status === 'active' ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
        >
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: any, record: ProviderRouteConfig) => (
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => handleEditProvider(record)}
          size="small"
        >
          配置
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
        <Tag
          color={type === 'enterprise' ? 'gold' : 'default'}
          icon={type === 'enterprise' ? <CrownOutlined /> : undefined}
        >
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
        <Tag
          color={status === 'active' ? 'success' : 'default'}
          icon={status === 'active' ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
        >
          {status === 'active' ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: any, record: MerchantRouteConfig) => (
        <Button
          type="link"
          icon={<EditOutlined />}
          onClick={() => handleEditMerchant(record)}
          size="small"
        >
          配置
        </Button>
      ),
    },
  ];

  const domesticProviders = providers.filter((p) => p.provider_region === 'domestic');
  const overseasProviders = providers.filter((p) => p.provider_region === 'overseas');
  const domesticMerchants = merchants.filter((m) => m.region === 'domestic');
  const overseasMerchants = merchants.filter((m) => m.region === 'overseas');

  return (
    <div style={{ padding: '0 0 24px 0' }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="厂商总数"
              value={providers.length}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="海外厂商"
              value={overseasProviders.length}
              prefix={<GlobalOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="商户总数"
              value={merchants.length}
              prefix={<ApiOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card size="small">
            <Statistic
              title="企业商户"
              value={merchants.filter((m) => m.merchant_type === 'enterprise').length}
              prefix={<CrownOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 16 }}>
        <Alert
          message={
            <Space>
              <span>统一路由配置管理</span>
              <Tooltip title="通过配置厂商区域、路由策略和端点信息，实现基于商户类型和区域的智能路由决策">
                <InfoCircleOutlined style={{ color: '#1890ff' }} />
              </Tooltip>
            </Space>
          }
          description="系统会根据商户类型和区域自动选择最优路由模式（直连/LiteLLM/代理），无需手动干预。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Tabs defaultActiveKey="providers">
          <TabPane
            tab={
              <Space>
                <ApiOutlined />
                <span>厂商配置</span>
                <Tag color="blue">{providers.length}</Tag>
              </Space>
            }
            key="providers"
          >
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Card
                size="small"
                title={
                  <Space>
                    <GlobalOutlined style={{ color: '#52c41a' }} />
                    国内厂商
                  </Space>
                }
              >
                <Table
                  columns={providerColumns}
                  dataSource={domesticProviders}
                  rowKey="id"
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ x: 600 }}
                />
              </Card>

              <Card
                size="small"
                title={
                  <Space>
                    <GlobalOutlined style={{ color: '#1890ff' }} />
                    海外厂商
                  </Space>
                }
              >
                <Table
                  columns={providerColumns}
                  dataSource={overseasProviders}
                  rowKey="id"
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ x: 600 }}
                />
              </Card>
            </Space>
          </TabPane>

          <TabPane
            tab={
              <Space>
                <ApiOutlined />
                <span>商户配置</span>
                <Tag color="purple">{merchants.length}</Tag>
              </Space>
            }
            key="merchants"
          >
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <Card
                size="small"
                title={
                  <Space>
                    <GlobalOutlined style={{ color: '#52c41a' }} />
                    国内商户
                  </Space>
                }
              >
                <Table
                  columns={merchantColumns}
                  dataSource={domesticMerchants}
                  rowKey="id"
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ x: 600 }}
                />
              </Card>

              <Card
                size="small"
                title={
                  <Space>
                    <GlobalOutlined style={{ color: '#1890ff' }} />
                    海外商户
                  </Space>
                }
              >
                <Table
                  columns={merchantColumns}
                  dataSource={overseasMerchants}
                  rowKey="id"
                  loading={loading}
                  pagination={false}
                  size="small"
                  scroll={{ x: 600 }}
                />
              </Card>
            </Space>
          </TabPane>

          <TabPane
            tab={
              <Space>
                <ExperimentOutlined />
                <span>路由测试</span>
              </Space>
            }
            key="test"
          >
            <Card>
              <Space direction="vertical" style={{ width: '100%' }} size="large">
                <Row gutter={16}>
                  <Col xs={24} sm={12} md={8}>
                    <Input
                      placeholder="厂商代码（如: openai）"
                      value={testForm.provider_code}
                      onChange={(e) => setTestForm({ ...testForm, provider_code: e.target.value })}
                      prefix={<ApiOutlined />}
                    />
                  </Col>
                  <Col xs={24} sm={12} md={8}>
                    <InputNumber
                      placeholder="商户ID"
                      value={testForm.merchant_id}
                      onChange={(v) => setTestForm({ ...testForm, merchant_id: v || 1 })}
                      style={{ width: '100%' }}
                      min={1}
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8}>
                    <Button
                      type="primary"
                      onClick={handleTestRoute}
                      loading={testLoading}
                      icon={<ExperimentOutlined />}
                      block
                    >
                      测试路由
                    </Button>
                  </Col>
                </Row>

                {testResult && (
                  <>
                    <Divider>测试结果</Divider>
                    <Row gutter={[16, 16]}>
                      <Col xs={24} md={12}>
                        <Card
                          size="small"
                          title="路由决策"
                          style={{ backgroundColor: '#f6ffed', borderColor: '#b7eb8f' }}
                        >
                          <Space direction="vertical" style={{ width: '100%' }}>
                            <div>
                              <Text strong>路由模式: </Text>
                              <Tag color="blue">{testResult.mode}</Tag>
                            </div>
                            <div>
                              <Text strong>端点: </Text>
                              <Text copyable>{testResult.endpoint || '-'}</Text>
                            </div>
                            <div>
                              <Text strong>降级模式: </Text>
                              <Tag color="orange">{testResult.fallback_mode || '无'}</Tag>
                            </div>
                            <div>
                              <Text strong>降级端点: </Text>
                              <Text copyable>{testResult.fallback_endpoint || '-'}</Text>
                            </div>
                            <div>
                              <Text strong>决策原因: </Text>
                              <Text type="secondary">{testResult.reason}</Text>
                            </div>
                          </Space>
                        </Card>
                      </Col>

                      <Col xs={24} md={12}>
                        <Card size="small" title="厂商配置">
                          <pre
                            style={{ fontSize: 11, maxHeight: 200, overflow: 'auto', margin: 0 }}
                          >
                            {JSON.stringify(testResult.provider_config, null, 2)}
                          </pre>
                        </Card>
                      </Col>

                      <Col xs={24}>
                        <Card size="small" title="商户配置">
                          <pre
                            style={{ fontSize: 11, maxHeight: 200, overflow: 'auto', margin: 0 }}
                          >
                            {JSON.stringify(testResult.merchant_config, null, 2)}
                          </pre>
                        </Card>
                      </Col>
                    </Row>
                  </>
                )}
              </Space>
            </Card>
          </TabPane>
        </Tabs>
      </Card>

      <ProviderConfigForm
        visible={providerModalVisible}
        provider={editingProvider}
        onCancel={() => setProviderModalVisible(false)}
        onOk={handleSaveProvider}
        loading={saveLoading}
      />

      <MerchantConfigForm
        visible={merchantModalVisible}
        merchant={editingMerchant}
        onCancel={() => setMerchantModalVisible(false)}
        onOk={handleSaveMerchant}
        loading={saveLoading}
      />
    </div>
  );
};

export default AdminRouteConfigs;
