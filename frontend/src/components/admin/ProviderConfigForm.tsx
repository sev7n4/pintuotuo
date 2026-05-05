import React, { useEffect } from 'react';
import { Modal, Form, Select, Divider, Typography, Space, Tooltip, Alert, Tabs } from 'antd';
import {
  QuestionCircleOutlined,
  InfoCircleOutlined,
  EyeOutlined,
  EditOutlined,
} from '@ant-design/icons';
import RouteStrategyConfig from './RouteStrategyConfig';
import EndpointsConfig from './EndpointsConfig';
import RoutePreview from './RoutePreview';

const { Text } = Typography;
const { TabPane } = Tabs;

interface ProviderRouteConfig {
  id: number;
  code: string;
  name: string;
  provider_region: string;
  route_strategy: Record<string, any>;
  endpoints: Record<string, any>;
  status: string;
}

interface ProviderConfigFormProps {
  visible: boolean;
  provider: ProviderRouteConfig | null;
  onCancel: () => void;
  onOk: (values: any) => void;
  loading?: boolean;
}

const ProviderConfigForm: React.FC<ProviderConfigFormProps> = ({
  visible,
  provider,
  onCancel,
  onOk,
  loading,
}) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (visible && provider) {
      form.setFieldsValue({
        provider_region: provider.provider_region || 'domestic',
        route_strategy: provider.route_strategy || {},
        endpoints: provider.endpoints || {},
      });
    }
  }, [visible, provider, form]);

  const handleOk = async () => {
    try {
      const values = await form.validateFields();
      onOk(values);
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  return (
    <Modal
      title={
        <Space>
          <span>配置厂商路由</span>
          {provider && (
            <Text type="secondary" style={{ fontSize: 14 }}>
              {provider.name} ({provider.code})
            </Text>
          )}
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      onOk={handleOk}
      confirmLoading={loading}
      width={900}
      style={{ top: 20 }}
      bodyStyle={{ maxHeight: '70vh', overflowY: 'auto' }}
    >
      <Form form={form} layout="vertical">
        <Tabs defaultActiveKey="config">
          <TabPane
            tab={
              <Space>
                <EditOutlined />
                配置
              </Space>
            }
            key="config"
          >
            <Alert
              message="配置说明"
              description="通过配置厂商区域、路由策略和端点信息，系统会根据商户类型和区域自动选择最优路由模式。"
              type="info"
              showIcon
              style={{ marginBottom: 16 }}
            />

            <Form.Item
              name="provider_region"
              label={
                <Space>
                  <span>厂商区域</span>
                  <Tooltip title="标识厂商服务器所在区域，影响路由决策">
                    <QuestionCircleOutlined style={{ color: '#999' }} />
                  </Tooltip>
                </Space>
              }
            >
              <Select
                options={[
                  { value: 'domestic', label: '国内' },
                  { value: 'overseas', label: '海外' },
                ]}
              />
            </Form.Item>

            <Divider orientation="left">
              <Space>
                <span>路由策略</span>
                <Tooltip title="为不同用户类型配置路由策略">
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </Space>
            </Divider>

            <Form.Item name="route_strategy" noStyle>
              <RouteStrategyConfig providerRegion={form.getFieldValue('provider_region')} />
            </Form.Item>

            <Divider orientation="left">
              <Space>
                <span>端点配置</span>
                <Tooltip title="配置不同路由模式的端点地址">
                  <InfoCircleOutlined style={{ color: '#1890ff' }} />
                </Tooltip>
              </Space>
            </Divider>

            <Form.Item name="endpoints" noStyle>
              <EndpointsConfig providerCode={provider?.code} />
            </Form.Item>
          </TabPane>

          <TabPane
            tab={
              <Space>
                <EyeOutlined />
                预览
              </Space>
            }
            key="preview"
          >
            <Form.Item shouldUpdate noStyle>
              {({ getFieldValue }) => (
                <RoutePreview
                  routeStrategy={getFieldValue('route_strategy')}
                  endpoints={getFieldValue('endpoints')}
                  providerRegion={getFieldValue('provider_region')}
                />
              )}
            </Form.Item>
          </TabPane>
        </Tabs>
      </Form>
    </Modal>
  );
};

export default ProviderConfigForm;
