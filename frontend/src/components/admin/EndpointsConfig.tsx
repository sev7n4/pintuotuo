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
} from '@ant-design/icons';
import { validateEndpoints } from '@utils/routeConfigValidation';

const { Text } = Typography;

interface EndpointConfig {
  domestic?: string;
  overseas?: string;
  [key: string]: string | undefined;
}

interface EndpointsConfigProps {
  value?: Record<string, EndpointConfig>;
  onChange?: (value: Record<string, EndpointConfig>) => void;
}

const routeTypes = [
  { value: 'litellm', label: 'LiteLLM', icon: <ThunderboltOutlined />, color: 'blue' },
  { value: 'proxy', label: '代理', icon: <GlobalOutlined />, color: 'green' },
  { value: 'direct', label: '直连', icon: <ApiOutlined />, color: 'orange' },
];

const EndpointsConfig: React.FC<EndpointsConfigProps> = ({ value = {}, onChange }) => {
  const [testingEndpoint, setTestingEndpoint] = useState<string | null>(null);

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
    try {
      const response = await fetch(endpoint, { method: 'HEAD', signal: AbortSignal.timeout(5000) });
      if (response.ok) {
        message.success('端点连接正常');
      } else {
        message.warning(`端点响应异常: ${response.status}`);
      }
    } catch (error) {
      message.error('端点连接失败');
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
    </Card>
  );
};

export default EndpointsConfig;
