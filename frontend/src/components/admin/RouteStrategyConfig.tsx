import React from 'react';
import {
  Card,
  Form,
  Select,
  InputNumber,
  Switch,
  Space,
  Tooltip,
  Row,
  Col,
  Typography,
} from 'antd';
import { QuestionCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';

const { Text } = Typography;

interface RouteStrategyItem {
  mode: string;
  weight?: number;
  fallback_mode?: string;
  conditions?: Record<string, any>;
}

interface RouteStrategyConfigProps {
  value?: Record<string, RouteStrategyItem>;
  onChange?: (value: Record<string, RouteStrategyItem>) => void;
  providerRegion?: string;
}

const userTypes = [
  { value: 'domestic_users', label: '国内用户' },
  { value: 'overseas_users', label: '海外用户' },
  { value: 'enterprise_users', label: '企业用户' },
  { value: 'default_mode', label: '默认模式' },
];

const routeModes = [
  { value: 'direct', label: '直连', desc: '直接访问厂商API' },
  { value: 'litellm', label: 'LiteLLM', desc: '通过LiteLLM网关访问' },
  { value: 'proxy', label: '代理', desc: '通过代理服务器访问' },
  { value: 'auto', label: '自动', desc: '系统自动选择最优路由' },
];

const RouteStrategyConfig: React.FC<RouteStrategyConfigProps> = ({
  value = {},
  onChange,
  providerRegion,
}) => {
  const handleStrategyChange = (userType: string, field: string, fieldValue: any) => {
    const newStrategy = {
      ...value,
      [userType]: {
        ...(value[userType] || {}),
        [field]: fieldValue,
      },
    };
    onChange?.(newStrategy);
  };

  const getStrategyValue = (userType: string, field: string) => {
    return (value?.[userType] as any)?.[field];
  };

  return (
    <Card
      size="small"
      title={
        <Space>
          <span>路由策略配置</span>
          <Tooltip title="为不同类型的用户配置不同的路由策略，系统会根据用户类型自动选择最优路由">
            <QuestionCircleOutlined style={{ color: '#999', fontSize: 14 }} />
          </Tooltip>
        </Space>
      }
      style={{ marginBottom: 16 }}
    >
      <Row gutter={[16, 16]}>
        {userTypes.map((userType) => (
          <Col xs={24} sm={24} md={12} key={userType.value}>
            <Card
              size="small"
              type="inner"
              title={
                <Space>
                  <Text strong>{userType.label}</Text>
                  <Tooltip title={`为${userType.label}配置专属路由策略`}>
                    <InfoCircleOutlined style={{ color: '#1890ff', fontSize: 12 }} />
                  </Tooltip>
                </Space>
              }
              style={{ backgroundColor: '#fafafa' }}
            >
              <Form layout="vertical" size="small">
                <Form.Item label="路由模式" style={{ marginBottom: 12 }}>
                  <Select
                    value={getStrategyValue(userType.value, 'mode') || 'auto'}
                    onChange={(v) => handleStrategyChange(userType.value, 'mode', v)}
                    options={routeModes.map((m) => ({
                      value: m.value,
                      label: (
                        <Space>
                          <span>{m.label}</span>
                          <Text type="secondary" style={{ fontSize: 12 }}>
                            {m.desc}
                          </Text>
                        </Space>
                      ),
                    }))}
                  />
                </Form.Item>

                <Form.Item label="权重" style={{ marginBottom: 12 }}>
                  <Space>
                    <InputNumber
                      min={1}
                      max={100}
                      value={getStrategyValue(userType.value, 'weight') || 100}
                      onChange={(v) => handleStrategyChange(userType.value, 'weight', v || 100)}
                      style={{ width: 80 }}
                    />
                    <Tooltip title="权重越高，优先级越高（1-100）">
                      <InfoCircleOutlined style={{ color: '#999' }} />
                    </Tooltip>
                  </Space>
                </Form.Item>

                <Form.Item label="降级模式" style={{ marginBottom: 12 }}>
                  <Select
                    value={getStrategyValue(userType.value, 'fallback_mode')}
                    onChange={(v) => handleStrategyChange(userType.value, 'fallback_mode', v)}
                    placeholder="可选：主路由失败时的降级方案"
                    allowClear
                    options={routeModes
                      .filter((m) => m.value !== 'auto')
                      .map((m) => ({
                        value: m.value,
                        label: m.label,
                      }))}
                  />
                </Form.Item>

                <Form.Item label="启用" style={{ marginBottom: 0 }}>
                  <Switch
                    checked={getStrategyValue(userType.value, 'enabled') !== false}
                    onChange={(v) => handleStrategyChange(userType.value, 'enabled', v)}
                    checkedChildren="开"
                    unCheckedChildren="关"
                  />
                </Form.Item>
              </Form>
            </Card>
          </Col>
        ))}
      </Row>

      {providerRegion === 'overseas' && (
        <Card
          size="small"
          type="inner"
          style={{ marginTop: 16, backgroundColor: '#fff7e6', borderColor: '#ffd591' }}
        >
          <Space>
            <InfoCircleOutlined style={{ color: '#fa8c16' }} />
            <Text type="warning">
              该厂商为海外厂商，国内用户访问时建议配置 LiteLLM 或代理模式以获得更好的访问体验
            </Text>
          </Space>
        </Card>
      )}
    </Card>
  );
};

export default RouteStrategyConfig;
