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

const { Text } = Typography;

interface EndpointConfig {
  domestic?: string;
  overseas?: string;
  [key: string]: string | undefined;
}

interface EndpointsConfigProps {
  value?: Record<string, EndpointConfig>;
  onChange?: (value: Record<string, EndpointConfig>) => void;
  providerCode?: string;
  apiKey?: string;
}

const routeTypes = [
  { value: 'litellm', label: 'LiteLLM', icon: <ThunderboltOutlined />, color: 'blue' },
  { value: 'proxy', label: '代理', icon: <GlobalOutlined />, color: 'green' },
  { value: 'direct', label: '直连', icon: <ApiOutlined />, color: 'orange' },
];

const EndpointsConfig: React.FC<EndpointsConfigProps> = ({ value = {}, onChange, providerCode, apiKey }) => {
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

  const handleAddEndpoint = () => {
    const newKey = `endpoint_${Date.now()}`;
    const newEndpoints = {
      ...value,
      [newKey]: { domestic: '', overseas: '' },
    };
    onChange?.(newEndpoints);
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

  const handleTestEndpoint = async (endpoint: string) => {
    if (!endpoint) {
      message.warning('请先输入端点地址');
      return;
    }

    setTestingEndpoint(endpoint);
    setProbeResult(null);
    try {
      if (providerCode) {
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
      } else {
        const response = await fetch(endpoint, { method: 'HEAD', signal: AbortSignal.timeout(5000) });
        if (response.ok) {
          message.success('端点连接正常');
        } else {
          message.warning(`端点响应异常: ${response.status}`);
        }
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
                <Text type="secondary">暂无端点配置，点击右上角添加</Text>
              </Space>
            </Card>
          </Col>
        )}
      </Row>

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
                <Tag color="success" icon={<CheckCircleOutlined />}>连接成功</Tag>
              ) : (
                <Tag color="error" icon={<CloseCircleOutlined />}>连接失败</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="响应状态码">
              {probeResult.status_code > 0 ? (
                <Tag color={probeResult.status_code >= 200 && probeResult.status_code < 300 ? 'success' : 'warning'}>
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
                  {probeResult.latency_ms >= 100 && probeResult.latency_ms < 500 && <Tag color="warning">正常</Tag>}
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
