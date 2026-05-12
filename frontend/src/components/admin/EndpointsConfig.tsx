import React, { useState, useMemo } from 'react';
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  Tooltip,
  Row,
  Col,
  Typography,
  Tag,
  message,
  Spin,
  Alert,
  Modal,
  Descriptions,
  Checkbox,
  Select,
  Divider,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  QuestionCircleOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  GlobalOutlined,
  ApiOutlined,
  ThunderboltOutlined,
  WarningOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons';
import { validateEndpoints } from '@utils/routeConfigValidation';
import { adminService, ProbeEndpointResponse } from '@services/admin';
import {
  ENDPOINT_TYPE_LABELS,
  BILLING_UNIT_LABELS,
  type EndpointType,
  type BillingUnit,
} from '@/types/sku';

const { Text } = Typography;

interface EndpointConfig {
  domestic?: string;
  overseas?: string;
  [key: string]: string | undefined;
}

interface EndpointTypeConfig {
  enabled: boolean;
  url_template: string;
  billing_unit: BillingUnit;
}

interface EndpointsConfigProps {
  value?: Record<string, EndpointConfig>;
  onChange?: (value: Record<string, EndpointConfig>) => void;
  providerCode?: string;
  apiKey?: string;
  endpointTypes?: Record<string, EndpointTypeConfig>;
  onEndpointTypesChange?: (value: Record<string, EndpointTypeConfig>) => void;
}

const routeTypes = [
  { value: 'litellm', label: 'LiteLLM', icon: <ThunderboltOutlined />, color: 'blue' },
  { value: 'proxy', label: '代理', icon: <GlobalOutlined />, color: 'green' },
  { value: 'direct', label: '直连', icon: <ApiOutlined />, color: 'orange' },
];

const endpointTypeOptions: {
  value: EndpointType;
  label: string;
  defaultUrl: string;
  defaultBillingUnit: BillingUnit;
}[] = [
  {
    value: 'chat_completions',
    label: '对话补全',
    defaultUrl: '/v1/chat/completions',
    defaultBillingUnit: 'token',
  },
  {
    value: 'responses',
    label: 'Response API',
    defaultUrl: '/v1/responses',
    defaultBillingUnit: 'token',
  },
  { value: 'embeddings', label: '嵌入', defaultUrl: '/v1/embeddings', defaultBillingUnit: 'token' },
  {
    value: 'images_generations',
    label: '图像生成',
    defaultUrl: '/v1/images/generations',
    defaultBillingUnit: 'image',
  },
  {
    value: 'images_variations',
    label: '图像变体',
    defaultUrl: '/v1/images/variations',
    defaultBillingUnit: 'image',
  },
  {
    value: 'images_edits',
    label: '图像编辑',
    defaultUrl: '/v1/images/edits',
    defaultBillingUnit: 'image',
  },
  {
    value: 'audio_transcriptions',
    label: '语音转文字',
    defaultUrl: '/v1/audio/transcriptions',
    defaultBillingUnit: 'second',
  },
  {
    value: 'audio_translations',
    label: '音频翻译',
    defaultUrl: '/v1/audio/translations',
    defaultBillingUnit: 'second',
  },
  {
    value: 'audio_speech',
    label: '语音合成',
    defaultUrl: '/v1/audio/speech',
    defaultBillingUnit: 'character',
  },
  {
    value: 'moderations',
    label: '内容审核',
    defaultUrl: '/v1/moderations',
    defaultBillingUnit: 'request',
  },
];

const billingUnitOptions = Object.entries(BILLING_UNIT_LABELS).map(([value, label]) => ({
  value,
  label,
}));

const EndpointsConfig: React.FC<EndpointsConfigProps> = ({
  value = {},
  onChange,
  providerCode,
  apiKey,
  endpointTypes = {},
  onEndpointTypesChange,
}) => {
  const [testingEndpoint, setTestingEndpoint] = useState<string | null>(null);
  const [probeResult, setProbeResult] = useState<ProbeEndpointResponse | null>(null);
  const [resultModalVisible, setResultModalVisible] = useState(false);

  const validationResult = useMemo(() => {
    const endpoints: Record<string, { url: string }> = {};
    Object.entries(value).forEach(([key, config]) => {
      if (config.domestic) {
        endpoints[`${key}_domestic`] = { url: config.domestic };
      }
      if (config.overseas) {
        endpoints[`${key}_overseas`] = { url: config.overseas };
      }
    });
    return validateEndpoints(endpoints);
  }, [value]);

  const [addModalVisible, setAddModalVisible] = useState(false);
  const [selectedRouteType, setSelectedRouteType] = useState<string>('');

  const handleAddEndpoint = () => {
    setAddModalVisible(true);
  };

  const handleConfirmAdd = () => {
    if (!selectedRouteType) {
      message.warning('请选择路由类型');
      return;
    }
    if (value[selectedRouteType]) {
      message.warning(
        `${routeTypes.find((t) => t.value === selectedRouteType)?.label || selectedRouteType}端点已存在`
      );
      return;
    }
    const newEndpoints = {
      ...value,
      [selectedRouteType]: { domestic: '', overseas: '' },
    };
    onChange?.(newEndpoints);
    setAddModalVisible(false);
    setSelectedRouteType('');
  };

  const handleRemoveEndpoint = (key: string) => {
    const newEndpoints = { ...value };
    delete newEndpoints[key];
    onChange?.(newEndpoints);
  };

  const handleEndpointChange = (routeType: string, region: string, endpointValue: string) => {
    const newEndpoints = {
      ...value,
      [routeType]: {
        ...value[routeType],
        [region]: endpointValue,
      },
    };
    onChange?.(newEndpoints);
  };

  const handleEndpointTypeToggle = (checkedValues: EndpointType[]) => {
    const newEndpointTypes: Record<string, EndpointTypeConfig> = {};
    checkedValues.forEach((et) => {
      const existing = endpointTypes[et];
      const option = endpointTypeOptions.find((o) => o.value === et);
      newEndpointTypes[et] = existing || {
        enabled: true,
        url_template: option?.defaultUrl || `/v1/${et.replace(/_/g, '/')}`,
        billing_unit: option?.defaultBillingUnit || 'token',
      };
    });
    Object.entries(endpointTypes).forEach(([key]) => {
      if (!checkedValues.includes(key as EndpointType)) {
        return;
      }
    });
    onEndpointTypesChange?.(newEndpointTypes);
  };

  const handleEndpointTypeConfigChange = (
    et: string,
    field: 'url_template' | 'billing_unit',
    val: string
  ) => {
    const newEndpointTypes = {
      ...endpointTypes,
      [et]: {
        ...endpointTypes[et],
        [field]: val,
      },
    };
    onEndpointTypesChange?.(newEndpointTypes);
  };

  const handleTestEndpoint = async (endpoint: string) => {
    if (!endpoint) {
      message.warning('请先输入端点地址');
      return;
    }

    if (!providerCode) {
      message.warning('请先保存厂商配置后再测试端点');
      return;
    }

    setTestingEndpoint(endpoint);
    setProbeResult(null);
    try {
      const response = await adminService.probeEndpoint(providerCode, endpoint, apiKey);
      const resultData = response.data.data;
      if (response.data.code === 0 && resultData) {
        setProbeResult(resultData);
        setResultModalVisible(true);
        if (resultData.success) {
          message.success('端点连接正常');
        } else {
          message.warning(`端点响应异常: ${resultData.error_msg || resultData.status_code}`);
        }
      } else {
        message.error(`探测失败: ${response.data.message || 'Unknown error'}`);
      }
    } catch (error) {
      message.error('端点连接失败');
      const result: ProbeEndpointResponse = {
        success: false,
        status_code: 0,
        latency_ms: 0,
        error_msg: error instanceof Error ? error.message : 'Unknown error',
      };
      setProbeResult(result);
      setResultModalVisible(true);
    } finally {
      setTestingEndpoint(null);
    }
  };

  const enabledEndpointTypes = Object.keys(endpointTypes).filter((k) => endpointTypes[k].enabled);

  return (
    <Card
      size="small"
      title={
        <Space>
          <span>端点配置</span>
          <Tooltip title="配置不同路由模式的端点地址，支持国内和海外两个区域">
            <QuestionCircleOutlined style={{ color: '#999', fontSize: 14 }} />
          </Tooltip>
        </Space>
      }
      extra={
        <Button type="dashed" size="small" icon={<PlusOutlined />} onClick={handleAddEndpoint}>
          添加端点
        </Button>
      }
      style={{ marginBottom: 16 }}
    >
      <Divider orientation="left" style={{ fontSize: 13, marginTop: 0 }}>
        端点类型启用
      </Divider>
      <Checkbox.Group
        value={enabledEndpointTypes as string[]}
        onChange={(vals) => handleEndpointTypeToggle(vals as EndpointType[])}
        style={{ width: '100%' }}
      >
        <Row gutter={[8, 8]}>
          {endpointTypeOptions.map((opt) => (
            <Col xs={12} sm={8} md={6} key={opt.value}>
              <Checkbox value={opt.value}>{opt.label}</Checkbox>
            </Col>
          ))}
        </Row>
      </Checkbox.Group>

      {enabledEndpointTypes.length > 0 && (
        <>
          <Divider orientation="left" style={{ fontSize: 13 }}>
            端点类型详细配置
          </Divider>
          <Row gutter={[16, 12]}>
            {enabledEndpointTypes.map((et) => {
              const config = endpointTypes[et];
              const label = ENDPOINT_TYPE_LABELS[et] || et;
              return (
                <Col xs={24} sm={12} md={8} key={et}>
                  <Card size="small" type="inner" style={{ backgroundColor: '#fafafa' }}>
                    <Space direction="vertical" style={{ width: '100%' }} size={8}>
                      <Tag color="blue">{label}</Tag>
                      <Form.Item label="URL 模板" style={{ marginBottom: 0 }}>
                        <Input
                          size="small"
                          placeholder="/v1/..."
                          value={config?.url_template || ''}
                          onChange={(e) =>
                            handleEndpointTypeConfigChange(et, 'url_template', e.target.value)
                          }
                        />
                      </Form.Item>
                      <Form.Item label="计费单位" style={{ marginBottom: 0 }}>
                        <Select
                          size="small"
                          value={config?.billing_unit || 'token'}
                          onChange={(v) => handleEndpointTypeConfigChange(et, 'billing_unit', v)}
                          options={billingUnitOptions}
                        />
                      </Form.Item>
                    </Space>
                  </Card>
                </Col>
              );
            })}
          </Row>
        </>
      )}

      <Divider orientation="left" style={{ fontSize: 13 }}>
        路由模式端点地址
      </Divider>

      {validationResult.errors.length > 0 && (
        <Alert
          message="端点配置验证失败"
          description={
            <div>
              {validationResult.errors.map((error, index) => (
                <div key={index}>
                  <CloseCircleOutlined style={{ color: '#ff4d4f', marginRight: 8 }} />
                  {error}
                </div>
              ))}
            </div>
          }
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {validationResult.warnings.length > 0 && (
        <Alert
          message="端点配置警告"
          description={
            <div>
              {validationResult.warnings.map((warning, index) => (
                <div key={index}>
                  <WarningOutlined style={{ color: '#faad14', marginRight: 8 }} />
                  {warning}
                </div>
              ))}
            </div>
          }
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={[16, 16]}>
        {Object.entries(value).map(([key, endpoints]) => {
          const routeType = routeTypes.find((t) => t.value === key) || {
            label: key,
            icon: <ApiOutlined />,
            color: 'default',
          };

          return (
            <Col xs={24} sm={24} md={12} lg={8} key={key}>
              <Card
                size="small"
                type="inner"
                title={
                  <Space>
                    <Tag color={routeType.color} icon={routeType.icon}>
                      {routeType.label}
                    </Tag>
                    <Tooltip title={`${routeType.label}端点配置`}>
                      <InfoCircleOutlined style={{ color: '#1890ff', fontSize: 12 }} />
                    </Tooltip>
                  </Space>
                }
                extra={
                  !routeTypes.find((t) => t.value === key) && (
                    <Button
                      type="text"
                      danger
                      size="small"
                      icon={<DeleteOutlined />}
                      onClick={() => handleRemoveEndpoint(key)}
                    />
                  )
                }
                style={{ backgroundColor: '#fafafa' }}
              >
                <Form layout="vertical" size="small">
                  <Form.Item label="国内端点" style={{ marginBottom: 12 }}>
                    <Space.Compact style={{ width: '100%' }}>
                      <Input
                        placeholder="http://localhost:4000/v1"
                        value={endpoints.domestic || ''}
                        onChange={(e) => handleEndpointChange(key, 'domestic', e.target.value)}
                      />
                      <Button
                        icon={
                          testingEndpoint === endpoints.domestic ? (
                            <Spin size="small" />
                          ) : endpoints.domestic ? (
                            <CheckCircleOutlined style={{ color: '#52c41a' }} />
                          ) : (
                            <CloseCircleOutlined style={{ color: '#d9d9d9' }} />
                          )
                        }
                        onClick={() => handleTestEndpoint(endpoints.domestic || '')}
                        disabled={!endpoints.domestic || testingEndpoint === endpoints.domestic}
                      />
                    </Space.Compact>
                  </Form.Item>

                  <Form.Item label="海外端点" style={{ marginBottom: 0 }}>
                    <Space.Compact style={{ width: '100%' }}>
                      <Input
                        placeholder="https://api.openai.com/v1"
                        value={endpoints.overseas || ''}
                        onChange={(e) => handleEndpointChange(key, 'overseas', e.target.value)}
                      />
                      <Button
                        icon={
                          testingEndpoint === endpoints.overseas ? (
                            <Spin size="small" />
                          ) : endpoints.overseas ? (
                            <CheckCircleOutlined style={{ color: '#52c41a' }} />
                          ) : (
                            <CloseCircleOutlined style={{ color: '#d9d9d9' }} />
                          )
                        }
                        onClick={() => handleTestEndpoint(endpoints.overseas || '')}
                        disabled={!endpoints.overseas || testingEndpoint === endpoints.overseas}
                      />
                    </Space.Compact>
                  </Form.Item>
                </Form>
              </Card>
            </Col>
          );
        })}

        {Object.keys(value).length === 0 && (
          <Col span={24}>
            <Card
              size="small"
              type="inner"
              style={{ textAlign: 'center', backgroundColor: '#fafafa' }}
            >
              <Space direction="vertical">
                <ApiOutlined style={{ fontSize: 32, color: '#d9d9d9' }} />
                <Text type="secondary">暂无路由端点配置，点击右上角添加</Text>
              </Space>
            </Card>
          </Col>
        )}
      </Row>

      <Modal
        title="添加端点配置"
        open={addModalVisible}
        onCancel={() => {
          setAddModalVisible(false);
          setSelectedRouteType('');
        }}
        onOk={handleConfirmAdd}
        okText="添加"
        cancelText="取消"
      >
        <Space direction="vertical" style={{ width: '100%' }} size="middle">
          <Alert
            message="选择要添加的路由类型端点"
            description="每种路由类型只能添加一个端点配置。如果已存在则无法重复添加。"
            type="info"
            showIcon
          />
          {routeTypes.map((rt) => {
            const exists = !!value[rt.value];
            return (
              <Card
                key={rt.value}
                size="small"
                hoverable={!exists}
                onClick={() => !exists && setSelectedRouteType(rt.value)}
                style={{
                  borderColor: selectedRouteType === rt.value ? '#1890ff' : undefined,
                  backgroundColor: exists ? '#f5f5f5' : undefined,
                  cursor: exists ? 'not-allowed' : 'pointer',
                }}
              >
                <Space>
                  <Tag color={rt.color} icon={rt.icon}>
                    {rt.label}
                  </Tag>
                  {exists ? (
                    <Text type="secondary">已配置</Text>
                  ) : (
                    <Text>{selectedRouteType === rt.value ? '已选择' : '点击选择'}</Text>
                  )}
                </Space>
              </Card>
            );
          })}
        </Space>
      </Modal>

      <Modal
        title="端点探测结果"
        open={resultModalVisible}
        onCancel={() => setResultModalVisible(false)}
        footer={[
          <Button key="close" type="primary" onClick={() => setResultModalVisible(false)}>
            关闭
          </Button>,
        ]}
      >
        {probeResult && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="探测状态">
              {probeResult.success ? (
                <Tag color="success" icon={<CheckCircleOutlined />}>
                  连接成功
                </Tag>
              ) : (
                <Tag color="error" icon={<CloseCircleOutlined />}>
                  连接失败
                </Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="响应状态码">
              {probeResult.status_code > 0 ? (
                <Tag
                  color={
                    probeResult.status_code >= 200 && probeResult.status_code < 300
                      ? 'success'
                      : 'warning'
                  }
                >
                  {probeResult.status_code}
                </Tag>
              ) : (
                <Tag>N/A</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="延迟">
              {probeResult.latency_ms > 0 ? (
                <Space>
                  <ClockCircleOutlined />
                  {probeResult.latency_ms} ms
                  {probeResult.latency_ms < 100 && <Tag color="success">优秀</Tag>}
                  {probeResult.latency_ms >= 100 && probeResult.latency_ms < 500 && (
                    <Tag color="warning">正常</Tag>
                  )}
                  {probeResult.latency_ms >= 500 && <Tag color="error">较慢</Tag>}
                </Space>
              ) : (
                <Tag>N/A</Tag>
              )}
            </Descriptions.Item>
            {probeResult.error_code && (
              <Descriptions.Item label="错误代码">
                <Tag color="error">{probeResult.error_code}</Tag>
              </Descriptions.Item>
            )}
            {probeResult.error_msg && (
              <Descriptions.Item label="错误信息">
                <Text type="danger">{probeResult.error_msg}</Text>
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>
    </Card>
  );
};

export default EndpointsConfig;
