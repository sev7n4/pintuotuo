import React, { useEffect } from 'react';
import { Modal, Form, Select, Card, Row, Col, Typography, Space, Tooltip, Alert, Tag } from 'antd';
import {
  QuestionCircleOutlined,
  InfoCircleOutlined,
  GlobalOutlined,
  CrownOutlined,
} from '@ant-design/icons';

const { Text } = Typography;

interface MerchantRouteConfig {
  id: number;
  name: string;
  merchant_type: string;
  region: string;
  route_preference: Record<string, any>;
  status: string;
}

interface MerchantConfigFormProps {
  visible: boolean;
  merchant: MerchantRouteConfig | null;
  onCancel: () => void;
  onOk: (values: any) => void;
  loading?: boolean;
}

const routeModes = [
  { value: 'direct', label: '直连', desc: '直接访问厂商API', color: 'orange' },
  { value: 'litellm', label: 'LiteLLM', desc: '通过LiteLLM网关访问', color: 'blue' },
  { value: 'proxy', label: '代理', desc: '通过代理服务器访问', color: 'green' },
  { value: 'auto', label: '自动', desc: '系统自动选择最优路由', color: 'default' },
];

const MerchantConfigForm: React.FC<MerchantConfigFormProps> = ({
  visible,
  merchant,
  onCancel,
  onOk,
  loading,
}) => {
  const [form] = Form.useForm();

  useEffect(() => {
    if (visible && merchant) {
      form.setFieldsValue({
        merchant_type: merchant.merchant_type || 'standard',
        region: merchant.region || 'domestic',
        route_preference: merchant.route_preference || {},
      });
    }
  }, [visible, merchant, form]);

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
          <span>配置商户路由</span>
          {merchant && (
            <Text type="secondary" style={{ fontSize: 14 }}>
              {merchant.name} (ID: {merchant.id})
            </Text>
          )}
        </Space>
      }
      open={visible}
      onCancel={onCancel}
      onOk={handleOk}
      confirmLoading={loading}
      width={700}
      style={{ top: 20 }}
    >
      <Form form={form} layout="vertical">
        <Alert
          message="配置说明"
          description="配置商户的类型和区域，系统会根据这些信息自动选择最优路由策略。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />

        <Row gutter={16}>
          <Col xs={24} sm={12}>
            <Form.Item
              name="merchant_type"
              label={
                <Space>
                  <span>商户类型</span>
                  <Tooltip title="标识商户类型，影响路由优先级">
                    <QuestionCircleOutlined style={{ color: '#999' }} />
                  </Tooltip>
                </Space>
              }
            >
              <Select
                options={[
                  {
                    value: 'standard',
                    label: (
                      <Space>
                        <Tag>标准</Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          普通商户
                        </Text>
                      </Space>
                    ),
                  },
                  {
                    value: 'enterprise',
                    label: (
                      <Space>
                        <Tag color="gold" icon={<CrownOutlined />}>
                          企业
                        </Tag>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          企业客户，优先级更高
                        </Text>
                      </Space>
                    ),
                  },
                ]}
              />
            </Form.Item>
          </Col>

          <Col xs={24} sm={12}>
            <Form.Item
              name="region"
              label={
                <Space>
                  <span>商户区域</span>
                  <Tooltip title="标识商户所在区域，影响路由选择">
                    <QuestionCircleOutlined style={{ color: '#999' }} />
                  </Tooltip>
                </Space>
              }
            >
              <Select
                options={[
                  {
                    value: 'domestic',
                    label: (
                      <Space>
                        <GlobalOutlined style={{ color: '#52c41a' }} />
                        <span>国内</span>
                      </Space>
                    ),
                  },
                  {
                    value: 'overseas',
                    label: (
                      <Space>
                        <GlobalOutlined style={{ color: '#1890ff' }} />
                        <span>海外</span>
                      </Space>
                    ),
                  },
                ]}
              />
            </Form.Item>
          </Col>
        </Row>

        <Card
          size="small"
          title={
            <Space>
              <span>路由偏好</span>
              <Tooltip title="可选：设置商户的首选路由模式">
                <InfoCircleOutlined style={{ color: '#1890ff' }} />
              </Tooltip>
            </Space>
          }
          style={{ marginTop: 16 }}
        >
          <Form.Item name={['route_preference', 'preferred_mode']} label="首选路由模式">
            <Select
              placeholder="不设置则使用系统默认策略"
              allowClear
              options={routeModes.map((m) => ({
                value: m.value,
                label: (
                  <Space>
                    <Tag color={m.color}>{m.label}</Tag>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {m.desc}
                    </Text>
                  </Space>
                ),
              }))}
            />
          </Form.Item>

          <Form.Item name={['route_preference', 'priority']} label="路由优先级">
            <Select
              placeholder="不设置则使用默认优先级"
              allowClear
              options={[
                { value: 'high', label: '高优先级' },
                { value: 'normal', label: '普通优先级' },
                { value: 'low', label: '低优先级' },
              ]}
            />
          </Form.Item>
        </Card>
      </Form>
    </Modal>
  );
};

export default MerchantConfigForm;
