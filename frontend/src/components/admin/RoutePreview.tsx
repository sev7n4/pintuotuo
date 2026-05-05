import React from 'react';
import { Card, Descriptions, Tag, Space, Typography, Divider, Alert, Table } from 'antd';
import {
  CheckCircleOutlined,
  GlobalOutlined,
  ApiOutlined,
  ThunderboltOutlined,
  UserOutlined,
  CrownOutlined,
} from '@ant-design/icons';

const { Text, Title } = Typography;

interface RouteStrategyItem {
  mode: string;
  weight?: number;
  fallback_mode?: string;
  conditions?: Record<string, any>;
}

interface EndpointConfig {
  domestic?: string;
  overseas?: string;
  [key: string]: string | undefined;
}

interface RoutePreviewProps {
  routeStrategy?: Record<string, RouteStrategyItem>;
  endpoints?: Record<string, EndpointConfig>;
  providerRegion?: string;
}

const userTypeConfig = {
  domestic_users: {
    label: '国内用户',
    icon: <UserOutlined />,
    color: 'blue',
    description: '中国大陆地区用户',
  },
  overseas_users: {
    label: '海外用户',
    icon: <GlobalOutlined />,
    color: 'green',
    description: '海外地区用户',
  },
  enterprise_users: {
    label: '企业用户',
    icon: <CrownOutlined />,
    color: 'gold',
    description: '企业级用户',
  },
  default_mode: {
    label: '默认模式',
    icon: <ApiOutlined />,
    color: 'default',
    description: '默认路由策略',
  },
};

const routeModeConfig = {
  direct: {
    label: '直连',
    icon: <ApiOutlined />,
    color: 'orange',
    description: '直接访问厂商API',
  },
  litellm: {
    label: 'LiteLLM',
    icon: <ThunderboltOutlined />,
    color: 'blue',
    description: '通过LiteLLM网关访问',
  },
  proxy: {
    label: '代理',
    icon: <GlobalOutlined />,
    color: 'green',
    description: '通过代理服务器访问',
  },
  auto: {
    label: '自动',
    icon: <CheckCircleOutlined />,
    color: 'purple',
    description: '系统自动选择最优路由',
  },
};

const RoutePreview: React.FC<RoutePreviewProps> = ({
  routeStrategy,
  endpoints,
  providerRegion,
}) => {
  const getRouteModeDisplay = (mode: string) => {
    const config = routeModeConfig[mode as keyof typeof routeModeConfig] || {
      label: mode,
      icon: <ApiOutlined />,
      color: 'default',
      description: '未知路由模式',
    };
    return (
      <Tag color={config.color} icon={config.icon}>
        {config.label}
      </Tag>
    );
  };

  const getEndpointDisplay = (mode: string, region?: string) => {
    if (!endpoints || !endpoints[mode]) {
      return <Text type="secondary">未配置</Text>;
    }

    const endpoint = endpoints[mode];
    const regionKey = region === 'overseas' ? 'overseas' : 'domestic';
    const url = endpoint[regionKey];

    if (!url) {
      return <Text type="secondary">未配置</Text>;
    }

    return (
      <Space direction="vertical" size="small">
        <div>
          <Tag color="blue">{regionKey === 'domestic' ? '国内' : '海外'}</Tag>
          <Text code style={{ fontSize: 12 }}>
            {url}
          </Text>
        </div>
      </Space>
    );
  };

  const columns = [
    {
      title: '用户类型',
      dataIndex: 'userType',
      key: 'userType',
      render: (userType: string) => {
        const config = userTypeConfig[userType as keyof typeof userTypeConfig];
        return (
          <Space>
            {config?.icon}
            <Text strong>{config?.label || userType}</Text>
          </Space>
        );
      },
    },
    {
      title: '路由模式',
      dataIndex: 'mode',
      key: 'mode',
      render: (mode: string) => getRouteModeDisplay(mode),
    },
    {
      title: '权重',
      dataIndex: 'weight',
      key: 'weight',
      render: (weight: number) => (
        <Tag color={weight >= 80 ? 'green' : weight >= 50 ? 'blue' : 'orange'}>{weight || 100}</Tag>
      ),
    },
    {
      title: '降级模式',
      dataIndex: 'fallback_mode',
      key: 'fallback_mode',
      render: (fallback_mode: string) =>
        fallback_mode ? getRouteModeDisplay(fallback_mode) : <Text type="secondary">无</Text>,
    },
    {
      title: '端点',
      key: 'endpoint',
      render: (_: any, record: any) => getEndpointDisplay(record.mode, providerRegion),
    },
  ];

  const tableData = Object.entries(routeStrategy || {}).map(([userType, item]) => ({
    userType,
    mode: item.mode,
    weight: item.weight,
    fallback_mode: item.fallback_mode,
  }));

  return (
    <Card
      size="small"
      title={
        <Space>
          <Title level={5} style={{ margin: 0 }}>
            路由配置预览
          </Title>
          <Tag color="blue">实时预览</Tag>
        </Space>
      }
    >
      {!routeStrategy || Object.keys(routeStrategy).length === 0 ? (
        <Alert
          message="暂无路由配置"
          description="请在上方配置路由策略和端点信息"
          type="info"
          showIcon
        />
      ) : (
        <Space direction="vertical" style={{ width: '100%' }} size="large">
          <div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              厂商区域: {providerRegion === 'overseas' ? '海外' : '国内'}
            </Text>
          </div>

          <Divider style={{ margin: '12px 0' }} />

          <Table
            columns={columns}
            dataSource={tableData}
            pagination={false}
            size="small"
            rowKey="userType"
          />

          <Divider style={{ margin: '12px 0' }} />

          <Descriptions title="路由决策说明" column={1} size="small">
            <Descriptions.Item label="决策逻辑">
              系统会根据用户类型自动匹配对应的路由策略，选择最优的访问路径
            </Descriptions.Item>
            <Descriptions.Item label="降级机制">
              当主路由模式失败时，系统会自动切换到降级模式（如果配置）
            </Descriptions.Item>
            <Descriptions.Item label="端点选择">
              根据厂商区域和用户区域，自动选择对应的端点地址
            </Descriptions.Item>
          </Descriptions>
        </Space>
      )}
    </Card>
  );
};

export default RoutePreview;
